package admin

import (
	"time"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/pkg/journal"
	"hstserver/pkg/logger"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
)

// The rollover: swaps, held profit, period commissions and expired orders, all at one hour of
// the day. Every engine pod holds the same answer, so a change goes to all of them at once and
// is written down as well, because a pod that restarts must not quietly go back to the default.

// endOfDayKey names the setting in hst.settings.
const endOfDayKey = "end_of_day_at"

// endOfDayLayout is how the hour is written, on the wire and in the setting.
const endOfDayLayout = "15:04"

// UptEndOfDay moves the hour the rollover runs at.
type UptEndOfDay struct {
	At string `json:"at" validate:"required,len=5"`
}

// ViewEndOfDay is when the rollover runs.
type ViewEndOfDay struct {
	At        string `json:"at"`
	UpdatedAt int64  `json:"updated_at"`
}

// GetEndOfDay reports the hour the rollover runs at.
//
//	@Id			GetEndOfDay
//	@Tags		System
//	@Produce	json
//	@Success	200	{object}	Response{data=ViewEndOfDay}
//	@Failure	403	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/system/end-of-day [get]
func (s *Server) GetEndOfDay(c *fiber.Ctx) error {
	var v ViewEndOfDay

	err := s.DB.DB.QueryRow(c.UserContext(),
		`SELECT value, updated_at FROM hst.settings WHERE key = $1`, endOfDayKey).
		Scan(&v.At, &v.UpdatedAt)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, v)
}

// UpdateEndOfDay moves the hour the rollover runs at, on every engine pod.
//
//	@Id			UpdateEndOfDay
//	@Tags		System
//	@Accept		json
//	@Produce	json
//	@Param		body	body		UptEndOfDay	true	"the hour, as 15:04"
//	@Success	200		{object}	Response{data=ViewEndOfDay}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/system/end-of-day [put]
func (s *Server) UpdateEndOfDay(c *fiber.Ctx) error {
	var body UptEndOfDay
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	if _, err := time.Parse(endOfDayLayout, body.At); err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrInvalidEndOfDayTime)
	}

	now := time.Now().UnixNano()

	var v ViewEndOfDay

	err := s.DB.DB.QueryRow(c.UserContext(),
		`INSERT INTO hst.settings (key, value, updated_at) VALUES ($1,$2,$3)
		 ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = $3
		 RETURNING value, updated_at`, endOfDayKey, body.At, now).
		Scan(&v.At, &v.UpdatedAt)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// written down first, so a pod coming up mid-change reads the new hour rather than the old
	s.NotifySystem(model.SubjectSystemEndOfDayTime, fiber.Map{"at": v.At})

	s.JournalEntry(c, model.JournalType_system, logger.CodeOK, journal.EndOfDayTimeMsg(v.At), v)

	return s.App.HttpResponseOK(c, v)
}

// RunEndOfDay runs the rollover now, for an operator who cannot wait for the hour.
//
//	@Id			RunEndOfDay
//	@Tags		System
//	@Produce	json
//	@Success	200	{object}	Response
//	@Failure	403	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/system/end-of-day/run [post]
func (s *Server) RunEndOfDay(c *fiber.Ctx) error {
	s.NotifySystem(model.SubjectSystemEndOfDay, fiber.Map{"at": time.Now().UnixNano()})

	s.JournalEntry(c, model.JournalType_system, logger.CodeAtt, journal.EndOfDayRunMsg(), nil)

	return s.App.HttpResponseOK(c, nil)
}
