package v1

import (
	"errors"
	"sync"
	"time"

	"context"
	"strings"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	nethttp "hstserver/pkg/http"
	"hstserver/pkg/journal"
	"hstserver/pkg/logger"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

// CrtUser creates a login, its account row, and optionally its manager row.
type CrtUser struct {
	ClientId         *int64 `json:"client_id"`
	Group            string `json:"group" validate:"required,max=128"`
	Rights           int64  `json:"rights"`
	Name             string `json:"name" validate:"required,max=128"`
	FirstName        string `json:"first_name" validate:"max=64"`
	LastName         string `json:"last_name" validate:"max=64"`
	Email            string `json:"email" validate:"required,email,max=255"`
	Phone            string `json:"phone" validate:"max=64"`
	Country          string `json:"country" validate:"max=64"`
	City             string `json:"city" validate:"max=64"`
	Comment          string `json:"comment" validate:"max=4096"`
	PasswordMain     string `json:"password_main" validate:"required,min=8,max=128"`
	PasswordInvestor string `json:"password_investor" validate:"required,min=8,max=128"`
	PasswordApi      string `json:"password_api" validate:"omitempty,min=8,max=128"`
}

// UptUser patches a login.
type UptUser struct {
	Group    *string `json:"group" validate:"omitempty,max=128"`
	Rights   *int64  `json:"rights"`
	Name     *string `json:"name" validate:"omitempty,max=128"`
	Email    *string `json:"email" validate:"omitempty,email,max=255"`
	Phone    *string `json:"phone" validate:"omitempty,max=64"`
	Country  *string `json:"country" validate:"omitempty,max=64"`
	City     *string `json:"city" validate:"omitempty,max=64"`
	Comment  *string `json:"comment" validate:"omitempty,max=4096"`
	Leverage *int32  `json:"leverage" validate:"omitempty,gte=1,lte=10000"`
}

// ViewUser is what the panel renders. Password hashes never appear.
type ViewUser struct {
	Login      int64             `json:"login"`
	ClientId   int64             `json:"client_id"`
	Group      string            `json:"group"`
	Rights     model.UsersRights `json:"rights"`
	Name       string            `json:"name"`
	Email      string            `json:"email"`
	Phone      string            `json:"phone"`
	Country    string            `json:"country"`
	City       string            `json:"city"`
	Leverage   int32             `json:"leverage"`
	Balance    float64           `json:"balance"`
	Credit     float64           `json:"credit"`
	IsManager  bool              `json:"is_manager"`
	LastAccess int64             `json:"last_access"`
	UpdatedAt  int64             `json:"updated_at"`
}

// usersSortable are the real columns of hst.users.
var usersSortable = utils.NewSortable(
	"login", "client_id", "name", "email", "registration",
	"last_access", "updated_at", "balance")

const userColumns = `u.login, COALESCE(u.client_id, 0), u."group", u.rights, u.name,
	u.email, u.phone, u.country, u.city, u.leverage, u.balance, u.credit,
	(m.login IS NOT NULL), u.last_access, u.updated_at`

const userJoin = ` FROM hst.users u LEFT JOIN hst.managers m ON m.login = u.login`

// CreateUser creates the identity and its 1:1 account row in one transaction.
//
//	@Id			CreateUser
//	@Tags		Users
//	@Accept		json
//	@Produce	json
//	@Param		body	body		CrtUser	true	"the user to create, an account row is created with it"
//	@Success	201		{object}	Response{data=ViewUser}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/users [post]
func (s *HttpServer) CreateUser(c *fiber.Ctx) error {
	var body CrtUser
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	// the request and the core carry the same fields, so the body is the argument
	login, status, err := s.OpenLogin(c.UserContext(), NewLogin(body))
	if err != nil {
		return s.App.HttpResponseStatus(c, status, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "user created",
		"actor", snap.Login, "target", login)

	return s.getUserByLogin(c, login, s.NotifyUser(model.EventUserCreated))
}

// NewLogin is a login to open, from a manager creating one or from a public signup.
type NewLogin struct {
	ClientId         *int64
	Group            string
	Rights           int64
	Name             string
	FirstName        string
	LastName         string
	Email            string
	Phone            string
	Country          string
	City             string
	Comment          string
	PasswordMain     string
	PasswordInvestor string
	PasswordApi      string
}

// OpenLogin creates a login and the account row that holds its money, in one transaction.
//
// The group decides what the account opens with, so a demo signup lands funded and everything
// else lands empty. A login without an account row has no money state, so neither is optional.
func (s *HttpServer) OpenLogin(ctx context.Context, n NewLogin) (int64, int, error) {
	// the group is what the account opens with, so it is read before anything is written
	deposit, leverage, status, err := s.OpeningBalance(ctx, n.Group)
	if err != nil {
		return 0, status, err
	}

	// hash before opening the transaction.
	hashes, err := s.hashPasswords(n.PasswordMain, n.PasswordInvestor, n.PasswordApi)
	if err != nil {
		return 0, nethttp.StatusInternalServerError, err
	}

	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return 0, nethttp.StatusInternalServerError, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	now := time.Now().UnixNano()

	var login int64
	if err := tx.QueryRow(ctx,
		`INSERT INTO hst.users
		   (client_id, "group", rights, name, first_name, last_name, email, phone,
		    country, city, comment, leverage,
		    password_main, password_investor, password_api,
		    registration, last_pass_change, updated_at, balance)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$16,$16,$17)
		 RETURNING login`,
		n.ClientId, n.Group, n.Rights, n.Name, n.FirstName, n.LastName,
		n.Email, n.Phone, n.Country, n.City, n.Comment, leverage,
		hashes[0], hashes[1], hashes[2], now, deposit).Scan(&login); err != nil {
		return 0, nethttp.StatusInternalServerError, err
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO hst.accounts (login, currency_digits, margin_leverage, balance, equity, updated_at)
		 VALUES ($1, 2, $2, $3, $3, $4)`, login, leverage, deposit, now); err != nil {
		return 0, nethttp.StatusInternalServerError, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, nethttp.StatusInternalServerError, err
	}

	return login, nethttp.StatusCreated, nil
}

// ListUsers returns a page of logins.
//
//	@Id			ListUsers
//	@Tags		Users
//	@Produce	json
//	@Param		page	query		int		false	"page number, from 1"
//	@Param		limit	query		int		false	"rows per page, max 500"
//	@Param		search	query		string	false	"matches name or email"
//	@Param		sort_by	query		string	false	"login, client_id, name, email, registration, last_access, updated_at, balance"	Enums(login, client_id, name, email, registration, last_access, updated_at, balance)
//	@Param		order	query		string	false	"asc or desc"																	Enums(asc, desc)
//	@Success	200		{object}	Response{data=[]ViewUser}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/users [get]
func (s *HttpServer) ListUsers(c *fiber.Ctx) error {
	q, err := utils.QueryFilter(c, usersSortable, "login")
	if err != nil {
		return s.App.HttpResponseBadQueryParams(c, err)
	}

	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	// only the logins inside this manager's groups, the same masks the websocket routes by
	access, args := utils.GroupAccessFor(snap.IsManager, snap.ManagerGroups, `u."group"`, 4)
	args = append([]any{q.Search, q.Limit, q.Offset}, args...)

	rows, err := s.DB.DB.Query(c.UserContext(),
		`SELECT `+userColumns+userJoin+`
		  WHERE ($1 = '' OR u.name ILIKE '%'||$1||'%' OR u.email ILIKE '%'||$1||'%')
		    AND `+access+`
		  ORDER BY u.`+q.SortBy+`
		  LIMIT $2 OFFSET $3`, args...)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	out := []ViewUser{}
	for rows.Next() {
		var v ViewUser
		if err := rows.Scan(&v.Login, &v.ClientId, &v.Group, &v.Rights, &v.Name,
			&v.Email, &v.Phone, &v.Country, &v.City, &v.Leverage, &v.Balance,
			&v.Credit, &v.IsManager, &v.LastAccess, &v.UpdatedAt); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		out = append(out, v)
	}
	if rows.Err() != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, rows.Err())
	}

	return s.App.HttpResponseOK(c, out)
}

// GetUser returns one login.
//
//	@Id			GetUser
//	@Tags		Users
//	@Produce	json
//	@Param		login	path		int	true	"login number"
//	@Success	200		{object}	Response{data=ViewUser}
//	@Failure	404		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/users/{login} [get]
func (s *HttpServer) GetUser(c *fiber.Ctx) error {
	login, err := c.ParamsInt("login")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	return s.getUserByLogin(c, int64(login), s.App.HttpResponseOK)
}

// UpdateUser patches a login.
//
//	@Id			UpdateUser
//	@Tags		Users
//	@Accept		json
//	@Produce	json
//	@Param		login	path		int		true	"login number"
//	@Param		body	body		UptUser	true	"only the fields to change"
//	@Success	200		{object}	Response{data=ViewUser}
//	@Failure	400		{object}	Response
//	@Failure	404		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/users/{login} [patch]
func (s *HttpServer) UpdateUser(c *fiber.Ctx) error {
	ctx := c.UserContext()

	login, err := c.ParamsInt("login")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	var body UptUser
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	// the previous group is needed before the write, the losers have to be told
	var oldGroup string
	if body.Group != nil {
		// a move lands the login in a real group or nowhere at all
		if err := s.GroupExists(ctx, *body.Group); err != nil {
			return s.App.HttpResponseBadRequest(c, err)
		}

		if err := s.DB.DB.QueryRow(ctx,
			`SELECT "group" FROM hst.users WHERE login = $1`, login).Scan(&oldGroup); err != nil &&
			!errors.Is(err, pgx.ErrNoRows) {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		if oldGroup == *body.Group {
			oldGroup = ""
		}
	}

	tag, err := s.DB.DB.Exec(ctx,
		`UPDATE hst.users SET
		    "group"    = COALESCE($2, "group"),
		    rights     = COALESCE($3, rights),
		    name       = COALESCE($4, name),
		    email      = COALESCE($5, email),
		    phone      = COALESCE($6, phone),
		    country    = COALESCE($7, country),
		    city       = COALESCE($8, city),
		    comment    = COALESCE($9, comment),
		    leverage   = COALESCE($10, leverage),
		    updated_at = $11
		  WHERE login = $1`,
		login, body.Group, body.Rights, body.Name, body.Email, body.Phone,
		body.Country, body.City, body.Comment, body.Leverage, time.Now().UnixNano())
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if tag.RowsAffected() == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	if body.Rights != nil || body.Group != nil {
		if err := s.OAuth2.InvalidateLogin(ctx, int64(login), model.SessionRevokedRightsChanged); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "user updated",
		"actor", snap.Login, "target", login)

	// a group change has two audiences: those who lost the record and those who gained it
	if oldGroup != "" {
		s.NotifyWS(model.SubjectUser(oldGroup), model.EventUserMoved,
			ViewUserRef{Login: int64(login), Group: oldGroup})
		s.JournalEntry(c, logger.CodeOK, journal.UserMovedMsg(snap.Login, int64(login)), oldGroup)
	}

	return s.getUserByLogin(c, int64(login), s.NotifyUser(model.EventUserUpdated))
}

// DeleteUser removes a login.
//
//	@Id			DeleteUser
//	@Tags		Users
//	@Produce	json
//	@Param		login	path		int	true	"login number"
//	@Success	204		{object}	Response
//	@Failure	404		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/users/{login} [delete]
func (s *HttpServer) DeleteUser(c *fiber.Ctx) error {
	ctx := c.UserContext()

	login, err := c.ParamsInt("login")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	// returning the group saves a read: it is needed to announce the delete and it does not exist afterwards
	var gone string
	err = s.DB.DB.QueryRow(ctx,
		`DELETE FROM hst.users WHERE login = $1 RETURNING "group"`, login).Scan(&gone)
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if err := s.OAuth2.InvalidateLogin(ctx, int64(login), model.SessionRevokedRightsChanged); err != nil {
		s.Log.Log(logger.TypeUser, logger.CodeWarn, "failed to drop sessions of deleted user",
			"login", login, "error", err.Error())
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeWarn, "user deleted",
		"actor", snap.Login, "target", login)

	ref := ViewUserRef{Login: int64(login), Group: gone}
	s.NotifyWS(model.SubjectUser(gone), model.EventUserDeleted, ref)
	s.NotifyWS(model.SubjectTraderProfile(int64(login)), model.EventUserDeleted, ref)
	s.NotifySystem(model.SubjectSystemUserDeleted, ref)
	s.JournalEntry(c, logger.CodeWarn, journal.UserDeletedMsg(snap.Login, int64(login)), ref)

	return s.App.HttpResponseNoContent(c)
}

// returning the group saves a read: it is needed to announce the delete
func (s *HttpServer) NotifyUser(event string) func(*fiber.Ctx, interface{}) error {
	return func(c *fiber.Ctx, v interface{}) error {
		if u, ok := v.(*ViewUser); ok {
			s.NotifyWS(model.SubjectUser(u.Group), event, u)
			// the account itself is told about its own record, on the root only it can hear
			s.NotifyWS(model.SubjectTraderProfile(u.Login), event, u)
			s.NotifySystem(SystemUserSubject(event), u)
			s.JournalEntry(c, logger.CodeOK, UserMsg(c, event, u), u)
		}
		if event == model.EventUserCreated {
			return s.App.HttpResponseCreated(c, v)
		}
		return s.App.HttpResponseOK(c, v)
	}
}

// getUserByLogin reads one row and answers with the given responder.
func (s *HttpServer) getUserByLogin(c *fiber.Ctx, login int64,
	respond func(*fiber.Ctx, interface{}) error) error {

	v := &ViewUser{}
	err := s.DB.DB.QueryRow(c.UserContext(),
		`SELECT `+userColumns+userJoin+` WHERE u.login = $1`, login).
		Scan(&v.Login, &v.ClientId, &v.Group, &v.Rights, &v.Name, &v.Email,
			&v.Phone, &v.Country, &v.City, &v.Leverage, &v.Balance, &v.Credit,
			&v.IsManager, &v.LastAccess, &v.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return respond(c, v)
}

// hashPasswords hashes the three slots concurrently.
func (s *HttpServer) hashPasswords(passwords ...string) ([]string, error) {
	out := make([]string, len(passwords))
	errs := make([]error, len(passwords))

	var wg sync.WaitGroup
	for i, p := range passwords {
		if p == "" {
			continue
		}
		wg.Add(1)
		go func(i int, p string) {
			defer wg.Done()
			out[i], errs[i] = s.OAuth2.Hasher.HashPassword(p)
		}(i, p)
	}
	wg.Wait()

	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}

// SystemUserSubject is the service side of a user event.
func SystemUserSubject(event string) string {
	if event == model.EventUserCreated {
		return model.SubjectSystemUserCreated
	}
	return model.SubjectSystemUserUpdated
}

// UserMsg is the journal line for a user event.
func UserMsg(c *fiber.Ctx, event string, u *ViewUser) string {
	snap, _ := utils.GetClient(c)

	var actor int64
	if snap != nil {
		actor = snap.Login
	}

	if event == model.EventUserCreated {
		return journal.UserCreatedMsg(actor, u.Login)
	}
	return journal.UserUpdatedMsg(actor, u.Login)
}

// demoSection is the group tree whose accounts open funded, as MT5 names it.
const demoSection = "demo"

// preliminarySection is the tree a real signup waits in until it is approved.
const preliminarySection = "preliminary"

// defaultLeverage is what an account gets when the group names none: 1, which is no leverage at all.
const defaultLeverage int32 = 1

// OpeningBalance is what an account in this group starts with.
//
// A demo group opens its accounts on the house: the deposit and the leverage are the group's, and
// unset means no money and no leverage rather than zero and zero, which would be an account that
// cannot trade at all. A live group opens empty and is funded by a real transfer.
func (s *HttpServer) OpeningBalance(ctx context.Context, group string) (float64, int32, int, error) {
	var (
		deposit  *float64
		leverage *int32
	)

	// the group is read whichever tree it is in: a login belongs to a group, so a group that does
	// not exist is a bad request rather than a login left pointing at nothing
	err := s.DB.DB.QueryRow(ctx,
		`SELECT demo_deposit, demo_leverage FROM hst.groups WHERE "group" = $1`, group).
		Scan(&deposit, &leverage)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, 0, nethttp.StatusBadRequest, errs.ErrGroupNotFound
	}
	if err != nil {
		return 0, 0, nethttp.StatusInternalServerError, err
	}

	// only the demo tree opens on the house
	if !IsDemoGroup(group) {
		return 0, defaultLeverage, nethttp.StatusOK, nil
	}

	return ptrOr(deposit, 0), ptrOr(leverage, defaultLeverage), nethttp.StatusOK, nil
}

// GroupExists reports whether a group path names a group that is there.
func (s *HttpServer) GroupExists(ctx context.Context, group string) error {
	var exists bool
	if err := s.DB.DB.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM hst.groups WHERE "group" = $1)`, group).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return errs.ErrGroupNotFound
	}
	return nil
}

// IsPreliminaryGroup reports whether the path opens onto the tree where accounts await approval.
func IsPreliminaryGroup(group string) bool {
	head, _, _ := strings.Cut(strings.TrimSpace(group), model.GroupSep)
	return strings.EqualFold(head, preliminarySection)
}

// IsDemoGroup reports whether the path opens onto the demo tree.
func IsDemoGroup(group string) bool {
	head, _, _ := strings.Cut(strings.TrimSpace(group), model.GroupSep)
	return strings.EqualFold(head, demoSection)
}
