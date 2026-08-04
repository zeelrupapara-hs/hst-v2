package admin

import (
	"errors"
	"fmt"
	v1 "hstserver/internal/server/v1"
	"sort"
	"strings"
	"time"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/pkg/journal"
	"hstserver/pkg/logger"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

// holidayOrderLock serialises the two writers that pick a config_index.
// ponytail: one advisory lock for the whole calendar; the table is tiny and edits
const holidayOrderLock int64 = 0x484f4c49 // "HOLI"

// CrtHoliday creates one holiday.
type CrtHoliday struct {
	Year        int32    `json:"year" validate:"gte=0,lte=9999"`
	Month       int16    `json:"month" validate:"required,gte=1,lte=12"`
	Day         int16    `json:"day" validate:"required,gte=1,lte=31"`
	From        int32    `json:"from" validate:"gte=0,lte=1439"`
	To          int32    `json:"to" validate:"gte=0,lte=1439"`
	Description string   `json:"description" validate:"max=128"`
	Mode        int16    `json:"mode" validate:"oneof=0 1"`
	Symbols     []string `json:"symbols" validate:"required,min=1,max=128,dive,required,max=128"`
}

// UptHoliday patches one holiday.
type UptHoliday struct {
	Year        *int32    `json:"year" validate:"omitempty,gte=0,lte=9999"`
	Month       *int16    `json:"month" validate:"omitempty,gte=1,lte=12"`
	Day         *int16    `json:"day" validate:"omitempty,gte=1,lte=31"`
	From        *int32    `json:"from" validate:"omitempty,gte=0,lte=1439"`
	To          *int32    `json:"to" validate:"omitempty,gte=0,lte=1439"`
	Description *string   `json:"description" validate:"omitempty,max=128"`
	Mode        *int16    `json:"mode" validate:"omitempty,oneof=0 1"`
	Symbols     *[]string `json:"symbols" validate:"omitempty,min=1,max=128,dive,required,max=128"`
}

// ReorderHolidays carries the new list order, most significant first.
type ReorderHolidays struct {
	HolidayIds []int64 `json:"holiday_ids" validate:"required,min=1"`
}

// HolidayWindow is one range of minutes the server stays open. To is inclusive.
type HolidayWindow struct {
	From int32 `json:"from"`
	To   int32 `json:"to"`
}

// ViewHolidayCheck is what the resolver answers for one symbol on one date.
type ViewHolidayCheck struct {
	Symbol  string          `json:"symbol"`
	Path    string          `json:"path"`
	Date    string          `json:"date"`
	Holiday bool            `json:"holiday"`
	Closed  bool            `json:"closed"`
	Windows []HolidayWindow `json:"windows"`
	Matched []int64         `json:"matched"`
}

// holidaysSortable are the real columns of holidays.
var holidaysSortable = utils.NewSortable(
	"holiday_id", "config_index", "year", "month", "day", "timestamp")

// generated from the struct so the select and the scan cannot drift apart
const holidayColumns = `holiday_id, year, month, day, "from", "to",
	description, "timestamp", mode, symbols, config_index`

// CreateHoliday appends a holiday to the calendar. It lands last, which matters
// only for list order, not for the resolver.
//
//	@Id			CreateHoliday
//	@Tags		Holidays
//	@Accept		json
//	@Produce	json
//	@Param		body	body		CrtHoliday	true	"the holiday to create"
//	@Success	201		{object}	Response{data=model.Holiday}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/holidays [post]
func (s *Server) CreateHoliday(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var body CrtHoliday
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}
	if err := validateHoliday(body.Year, body.Month, body.Day, body.From, body.To); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// without the lock two appends can pick the same config_index and one loses to the unique index.
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, holidayOrderLock); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	var next int32
	if err := tx.QueryRow(ctx,
		`SELECT COALESCE(MAX(config_index) + 1, 0) FROM hst.holidays`).Scan(&next); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	var out model.Holiday
	if err := tx.QueryRow(ctx,
		`INSERT INTO hst.holidays (year, month, day, "from", "to", description,
		    "timestamp", mode, symbols, config_index)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		 RETURNING `+holidayColumns,
		body.Year, body.Month, body.Day, body.From, body.To, body.Description,
		time.Now().UnixNano(), body.Mode, body.Symbols, next).
		Scan(&out.HolidayId, &out.Year, &out.Month, &out.Day, &out.From, &out.To,
			&out.Description, &out.Timestamp, &out.Mode, &out.Symbols, &out.ConfigIndex); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "holiday created",
		"actor", snap.Login, "target", out.HolidayId,
		"date", holidayDate(out), "work_time", workTime(out), "symbols", out.Symbols)

	s.NotifyWS(model.SubjectHoliday, model.EventHolidayCreated, out)
	s.NotifySystem(model.SubjectSystemHolidayCreated, out)
	s.JournalEntry(c, logger.TypeCfg, logger.CodeOK, journal.HolidayCreatedMsg(snap.Login, int(out.HolidayId)), out)

	return s.App.HttpResponseCreated(c, out)
}

// ListHolidays returns a page of the calendar.
//
//	@Id			ListHolidays
//	@Tags		Holidays
//	@Produce	json
//	@Param		page	query		int		false	"page number, from 1"
//	@Param		limit	query		int		false	"rows per page, max 500"
//	@Param		search	query		string	false	"matches description"
//	@Param		sort_by	query		string	false	"holiday_id, config_index, year, month, day, timestamp"			Enums(holiday_id, config_index, year, month, day, timestamp)
//	@Param		order	query		string	false	"asc or desc, asc by default because the list order is data"	Enums(asc, desc)
//	@Success	200		{object}	Response{data=[]model.Holiday}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/holidays [get]
func (s *Server) ListHolidays(c *fiber.Ctx) error {
	q, err := utils.QueryFilter(c, holidaysSortable, "config_index")
	if err != nil {
		return s.App.HttpResponseBadQueryParams(c, err)
	}

	// the list order is panel order, so ascending is the useful default
	if c.Query("order") == "" {
		q.SortBy = strings.TrimSuffix(q.SortBy, " DESC") + " ASC"
	}

	// sort_by is validated against an allowlist in QueryFilter.
	rows, err := s.DB.DB.Query(c.UserContext(),
		`SELECT `+holidayColumns+`
		   FROM hst.holidays
		  WHERE ($1 = '' OR description ILIKE '%'||$1||'%')
		  ORDER BY `+q.SortBy+`
		  LIMIT $2 OFFSET $3`, q.Search, q.Limit, q.Offset)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	out, err := scanHolidays(rows)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, out)
}

// GetHoliday returns one holiday.
//
//	@Id			GetHoliday
//	@Tags		Holidays
//	@Produce	json
//	@Param		id	path		int	true	"holiday id"
//	@Success	200	{object}	Response{data=model.Holiday}
//	@Failure	400	{object}	Response
//	@Failure	403	{object}	Response
//	@Failure	404	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/holidays/{id} [get]
func (s *Server) GetHoliday(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	var out model.Holiday
	err = s.DB.DB.QueryRow(c.UserContext(),
		`SELECT `+holidayColumns+` FROM hst.holidays WHERE holiday_id = $1`, id).
		Scan(&out.HolidayId, &out.Year, &out.Month, &out.Day, &out.From, &out.To,
			&out.Description, &out.Timestamp, &out.Mode, &out.Symbols, &out.ConfigIndex)
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, out)
}

// UpdateHoliday patches one holiday.
//
//	@Id			UpdateHoliday
//	@Tags		Holidays
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int			true	"holiday id"
//	@Param		body	body		UptHoliday	true	"only the fields to change. a symbols list replaces the whole mask set"
//	@Success	200		{object}	Response{data=model.Holiday}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	404		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/holidays/{id} [patch]
func (s *Server) UpdateHoliday(c *fiber.Ctx) error {
	ctx := c.UserContext()

	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	var body UptHoliday
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// the date and the work time are checked as a whole, and any part of either may be the one that is changing.
	var before model.Holiday
	err = tx.QueryRow(ctx,
		`SELECT year, month, day, "from", "to" FROM hst.holidays
		  WHERE holiday_id = $1 FOR UPDATE`, id).
		Scan(&before.Year, &before.Month, &before.Day, &before.From, &before.To)
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	after := before
	if body.Year != nil {
		after.Year = *body.Year
	}
	if body.Month != nil {
		after.Month = *body.Month
	}
	if body.Day != nil {
		after.Day = *body.Day
	}
	if body.From != nil {
		after.From = *body.From
	}
	if body.To != nil {
		after.To = *body.To
	}
	if err := validateHoliday(after.Year, after.Month, after.Day, after.From, after.To); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	// an untyped NULL has no type for COALESCE to infer, hence the cast
	var symbols any
	if body.Symbols != nil {
		symbols = *body.Symbols
	}

	var out model.Holiday
	if err := tx.QueryRow(ctx,
		`UPDATE hst.holidays SET
		    year        = COALESCE($2, year),
		    month       = COALESCE($3, month),
		    day         = COALESCE($4, day),
		    "from"      = COALESCE($5, "from"),
		    "to"        = COALESCE($6, "to"),
		    description = COALESCE($7, description),
		    mode        = COALESCE($8, mode),
		    symbols     = COALESCE($9::text[], symbols),
		    "timestamp" = $10
		  WHERE holiday_id = $1
		 RETURNING `+holidayColumns,
		id, body.Year, body.Month, body.Day, body.From, body.To,
		body.Description, body.Mode, symbols, time.Now().UnixNano()).
		Scan(&out.HolidayId, &out.Year, &out.Month, &out.Day, &out.From, &out.To,
			&out.Description, &out.Timestamp, &out.Mode, &out.Symbols, &out.ConfigIndex); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "holiday updated",
		"actor", snap.Login, "target", out.HolidayId,
		"date", holidayDate(out), "work_time", workTime(out), "symbols", out.Symbols)

	s.NotifyWS(model.SubjectHoliday, model.EventHolidayUpdated, out)
	s.NotifySystem(model.SubjectSystemHolidayUpdated, out)
	s.JournalEntry(c, logger.TypeCfg, logger.CodeOK, journal.HolidayUpdatedMsg(snap.Login, int(out.HolidayId)), out)

	return s.App.HttpResponseOK(c, out)
}

// DeleteHoliday removes a holiday and closes the gap it leaves, so the list
// order stays a dense 0..n-1 sequence.
//
//	@Id			DeleteHoliday
//	@Tags		Holidays
//	@Produce	json
//	@Param		id	path		int	true	"holiday id"
//	@Success	204	{object}	Response
//	@Failure	400	{object}	Response
//	@Failure	403	{object}	Response
//	@Failure	404	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/holidays/{id} [delete]
func (s *Server) DeleteHoliday(c *fiber.Ctx) error {
	ctx := c.UserContext()

	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, holidayOrderLock); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	var gone int32
	err = tx.QueryRow(ctx,
		`DELETE FROM hst.holidays WHERE holiday_id = $1 RETURNING config_index`, id).Scan(&gone)
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// shifting down cannot collide: the vacated index is always free
	if _, err := tx.Exec(ctx,
		`UPDATE hst.holidays SET config_index = config_index - 1
		  WHERE config_index > $1`, gone); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeWarn, "holiday deleted",
		"actor", snap.Login, "target", id)

	ref := v1.ViewHolidayRef{HolidayId: id}
	s.NotifyWS(model.SubjectHoliday, model.EventHolidayDeleted, ref)
	s.NotifySystem(model.SubjectSystemHolidayDeleted, ref)
	s.JournalEntry(c, logger.TypeCfg, logger.CodeWarn, journal.HolidayDeletedMsg(snap.Login, id), ref)

	return s.App.HttpResponseNoContent(c)
}

// ReorderHolidays rewrites the list order. The body must list every holiday
// exactly once, because a partial order is ambiguous. The order is parity
// only, the resolver does not read it.
//
//	@Id			ReorderHolidays
//	@Tags		Holidays
//	@Accept		json
//	@Produce	json
//	@Param		body	body		ReorderHolidays	true	"every holiday id, exactly once, in the new list order"
//	@Success	200		{object}	Response{data=[]model.Holiday}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/holidays/reorder [put]
func (s *Server) ReorderHolidays(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var body ReorderHolidays
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	seen := make(map[int64]struct{}, len(body.HolidayIds))
	for _, hid := range body.HolidayIds {
		if _, dup := seen[hid]; dup {
			return s.App.HttpResponseBadRequest(c, errs.ErrReorderMustListEveryHoliday)
		}
		seen[hid] = struct{}{}
	}

	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// the lock stops a concurrent append slipping in between the count and the rewrite and leaving a holiday.
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, holidayOrderLock); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	var total int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM hst.holidays`).Scan(&total); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if total != len(body.HolidayIds) {
		return s.App.HttpResponseBadRequest(c, errs.ErrReorderMustListEveryHoliday)
	}

	// with the counts equal and no duplicates, every id existing proves the two sets are identical.
	var owned int
	if err := tx.QueryRow(ctx,
		`SELECT count(*) FROM hst.holidays WHERE holiday_id = ANY($1)`,
		body.HolidayIds).Scan(&owned); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if owned != len(body.HolidayIds) {
		return s.App.HttpResponseBadRequest(c, errs.ErrReorderMustListEveryHoliday)
	}

	// park the indexes out of range first, otherwise the unique constraint fires the moment two holidays swap.
	if _, err := tx.Exec(ctx,
		`UPDATE hst.holidays SET config_index = -(config_index + 1)`); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	now := time.Now().UnixNano()
	for i, hid := range body.HolidayIds {
		if _, err := tx.Exec(ctx,
			`UPDATE hst.holidays SET config_index = $1, "timestamp" = $2
			  WHERE holiday_id = $3`, i, now, hid); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "holidays reordered",
		"actor", snap.Login, "holidays", len(body.HolidayIds))

	s.NotifyWS(model.SubjectHoliday, model.EventHolidayReordered, body)
	s.NotifySystem(model.SubjectSystemHolidayReordered, body)
	s.JournalEntry(c, logger.TypeCfg, logger.CodeOK, journal.HolidayReorderedMsg(snap.Login), body)

	rows, err := s.DB.DB.Query(ctx,
		`SELECT `+holidayColumns+` FROM hst.holidays ORDER BY config_index`)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	out, err := scanHolidays(rows)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, out)
}

// CheckHoliday resolves the calendar for one symbol on one date, which is the
// only way to see what a stack of overriding records actually adds up to.
//
//	@Id			CheckHoliday
//	@Tags		Holidays
//	@Produce	json
//	@Param		symbol	query		string	true	"symbol name, its path is looked up so group masks match"
//	@Param		date	query		string	false	"YYYY-MM-DD, today by default"
//	@Success	200		{object}	Response{data=ViewHolidayCheck}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/holidays/check [get]
func (s *Server) CheckHoliday(c *fiber.Ctx) error {
	ctx := c.UserContext()

	symbol := c.Query("symbol")
	if symbol == "" {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	on := time.Now().UTC()
	if raw := c.Query("date"); raw != "" {
		parsed, err := time.Parse(time.DateOnly, raw)
		if err != nil {
			return s.App.HttpResponseBadQueryParams(c, fmt.Errorf("date must be YYYY-MM-DD: %w", err))
		}
		on = parsed
	}

	// group masks are matched against the path, so it has to be resolved.
	path := symbol
	if err := s.DB.DB.QueryRow(ctx,
		`SELECT path FROM hst.symbols WHERE symbol = $1`, symbol).Scan(&path); err != nil &&
		!errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	rows, err := s.DB.DB.Query(ctx,
		`SELECT `+holidayColumns+`
		   FROM hst.holidays
		  WHERE mode = 1 AND month = $1 AND day = $2 AND (year = 0 OR year = $3)
		  ORDER BY config_index`, int(on.Month()), on.Day(), on.Year())
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	candidates, err := scanHolidays(rows)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	windows, matched, holiday := HolidayWindows(candidates, symbol, path, on)

	return s.App.HttpResponseOK(c, ViewHolidayCheck{
		Symbol:  symbol,
		Path:    path,
		Date:    on.Format(time.DateOnly),
		Holiday: holiday,
		Closed:  holiday && len(windows) == 0,
		Windows: windows,
		Matched: matched,
	})
}

// HolidayWindows merges the work time every matching holiday leaves open for a symbol on one date.
func HolidayWindows(hs []model.Holiday, symbol, path string, on time.Time) (
	windows []HolidayWindow, matched []int64, holiday bool) {

	windows, matched = []HolidayWindow{}, []int64{}

	for _, h := range hs {
		if h.Mode != model.HolidayMode_enabled || !h.OnDate(on) || !h.Covers(symbol, path) {
			continue
		}
		matched = append(matched, h.HolidayId)
		// a record with no work time is the prohibiting kind, it opens nothing
		if h.Closed() {
			continue
		}
		windows = append(windows, HolidayWindow{From: h.From, To: h.To})
	}
	if len(matched) == 0 {
		return windows, matched, false
	}

	sort.Slice(windows, func(i, j int) bool { return windows[i].From < windows[j].From })

	merged := windows[:0:0]
	for _, w := range windows {
		// To is inclusive, so a window starting one minute later still touches
		last := len(merged) - 1
		if last >= 0 && w.From <= merged[last].To+1 {
			if w.To > merged[last].To {
				merged[last].To = w.To
			}
			continue
		}
		merged = append(merged, w)
	}

	return merged, matched, true
}

// scanHolidays reads a whole result set of holidayColumns.
func scanHolidays(rows pgx.Rows) ([]model.Holiday, error) {
	out := []model.Holiday{}

	for rows.Next() {
		var v model.Holiday
		if err := rows.Scan(&v.HolidayId, &v.Year, &v.Month, &v.Day, &v.From,
			&v.To, &v.Description, &v.Timestamp, &v.Mode, &v.Symbols,
			&v.ConfigIndex); err != nil {
			return nil, err
		}
		out = append(out, v)
	}

	return out, rows.Err()
}

// validateHoliday checks the invariants the schema cannot express.
func validateHoliday(year int32, month, day int16, from, to int32) error {
	if to < from {
		return fmt.Errorf("to must not be earlier than from, got %d and %d", from, to)
	}

	// day 31 does not exist in February and a CHECK constraint cannot tell.
	y := int(year)
	if y == 0 {
		y = 2024
	}
	d := time.Date(y, time.Month(month), int(day), 0, 0, 0, 0, time.UTC)
	if d.Month() != time.Month(month) || d.Day() != int(day) {
		return errs.ErrHolidayDateInvalid
	}

	return nil
}

// holidayDate renders the date for the audit log, **** for a yearly holiday.
func holidayDate(h model.Holiday) string {
	if h.Year == 0 {
		return fmt.Sprintf("****-%02d-%02d", h.Month, h.Day)
	}
	return fmt.Sprintf("%04d-%02d-%02d", h.Year, h.Month, h.Day)
}

// workTime renders the work time for the audit log as clock time, or closed.
func workTime(h model.Holiday) string {
	if h.Closed() {
		return "closed"
	}
	return fmt.Sprintf("%02d:%02d-%02d:%02d",
		h.From/60, h.From%60, h.To/60, h.To%60)
}
