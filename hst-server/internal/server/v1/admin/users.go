package admin

import (
	"errors"
	v1 "hstserver/internal/server/v1"
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
func (s *Server) CreateUser(c *fiber.Ctx) error {
	var body CrtUser
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	// the request and the core carry the same fields, so the body is the argument
	login, status, err := s.OpenLogin(c.UserContext(), v1.NewLogin(body))
	if err != nil {
		return s.App.HttpResponseStatus(c, status, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "user created",
		"actor", snap.Login, "target", login)

	return s.getUserByLogin(c, login, s.NotifyUser(model.EventUserCreated))
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
func (s *Server) ListUsers(c *fiber.Ctx) error {
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
		  WHERE `+searchable(snap.ManagerRights)+`
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
		maskDetails(c, &v)
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
func (s *Server) GetUser(c *fiber.Ctx) error {
	login, err := c.ParamsInt("login")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	// a login the masks do not cover reads as absent, so the id space cannot be walked
	reach, err := s.AccountInReach(c, int64(login))
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if !reach {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	return s.getUserByLogin(c, int64(login), s.respondUser)
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
func (s *Server) UpdateUser(c *fiber.Ctx) error {
	ctx := c.UserContext()

	login, err := c.ParamsInt("login")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	// a login the masks do not cover reads as absent, so the id space cannot be walked
	reach, err := s.AccountInReach(c, int64(login))
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if !reach {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
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
			v1.ViewUserRef{Login: int64(login), Group: oldGroup})
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
func (s *Server) DeleteUser(c *fiber.Ctx) error {
	ctx := c.UserContext()

	login, err := c.ParamsInt("login")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	// a login the masks do not cover reads as absent, so the id space cannot be walked
	reach, err := s.AccountInReach(c, int64(login))
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if !reach {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
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

	ref := v1.ViewUserRef{Login: int64(login), Group: gone}
	s.NotifyWS(model.SubjectUser(gone), model.EventUserDeleted, ref)
	s.NotifyWS(model.SubjectTraderProfile(int64(login)), model.EventUserDeleted, ref)
	s.NotifySystem(model.SubjectSystemUserDeleted, ref)
	s.JournalEntry(c, logger.CodeWarn, journal.UserDeletedMsg(snap.Login, int64(login)), ref)

	return s.App.HttpResponseNoContent(c)
}

// returning the group saves a read: it is needed to announce the delete
func (s *Server) NotifyUser(event string) func(*fiber.Ctx, interface{}) error {
	return func(c *fiber.Ctx, v interface{}) error {
		if u, ok := v.(*ViewUser); ok {
			s.NotifyWS(model.SubjectUser(u.Group), event, u)
			// the account itself is told about its own record, on the root only it can hear
			s.NotifyWS(model.SubjectTraderProfile(u.Login), event, u)
			s.NotifySystem(SystemUserSubject(event), u)
			s.JournalEntry(c, logger.CodeOK, UserMsg(c, event, u), u)
		}
		body := v
		if u, ok := v.(*ViewUser); ok {
			masked := *u
			maskDetails(c, &masked)
			body = &masked
		}

		if event == model.EventUserCreated {
			return s.App.HttpResponseCreated(c, body)
		}
		return s.App.HttpResponseOK(c, body)
	}
}

// searchable is the free text predicate, over the columns this caller may read.
//
// Searching a column that is masked in the answer would still confirm what it holds, one guess
// at a time, so a caller who may not see an email may not search by one either.
func searchable(r model.ManagerRights) string {
	cols := []string{}
	if r.Has(model.MgrRightAccDetailsName) {
		cols = append(cols, `u.name ILIKE '%'||$1||'%'`)
	}
	if r.Has(model.MgrRightAccDetailsEmail) {
		cols = append(cols, `u.email ILIKE '%'||$1||'%'`)
	}

	if len(cols) == 0 {
		return `($1 = '' OR u.login::text = $1)`
	}

	return `($1 = '' OR u.login::text = $1 OR ` + strings.Join(cols, " OR ") + `)`
}

// maskDetails blanks the personal fields the caller was not granted.
//
// The platform grants sight of a client's name, location, address, document, email and phone
// separately, so a dealer who must see the book need not see who is behind it.
func maskDetails(c *fiber.Ctx, v *ViewUser) {
	snap, ok := utils.GetClient(c)
	if !ok {
		*v = ViewUser{Login: v.Login}
		return
	}

	r := snap.ManagerRights

	if !r.Has(model.MgrRightAccDetailsName) {
		v.Name = ""
	}
	if !r.Has(model.MgrRightAccDetailsLocation) {
		v.Country, v.City = "", ""
	}
	if !r.Has(model.MgrRightAccDetailsEmail) {
		v.Email = ""
	}
	if !r.Has(model.MgrRightAccDetailsPhone) {
		v.Phone = ""
	}
}

// getUserByLogin reads one row and answers with the given responder.
func (s *Server) getUserByLogin(c *fiber.Ctx, login int64,
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

// respondUser writes one user to the caller, masked to what they may read.
func (s *Server) respondUser(c *fiber.Ctx, v interface{}) error {
	if u, ok := v.(*ViewUser); ok {
		masked := *u
		maskDetails(c, &masked)
		return s.App.HttpResponseOK(c, &masked)
	}

	return s.App.HttpResponseOK(c, v)
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
