package v1

import (
	"errors"
	"sync"
	"time"

	"hstserver/model"
	errs "hstserver/pkg/errors"
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
	Leverage         int32  `json:"leverage" validate:"gte=1,lte=10000"`
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
	ctx := c.UserContext()

	var body CrtUser
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	// hash before opening the transaction.
	hashes, err := s.hashPasswords(body.PasswordMain, body.PasswordInvestor, body.PasswordApi)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	now := time.Now().UnixNano()

	var login int64
	if err := tx.QueryRow(ctx,
		`INSERT INTO hst.users
		   (client_id, "group", rights, name, first_name, last_name, email, phone,
		    country, city, comment, leverage,
		    password_main, password_investor, password_api,
		    registration, last_pass_change, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$16,$16)
		 RETURNING login`,
		body.ClientId, body.Group, body.Rights, body.Name, body.FirstName, body.LastName,
		body.Email, body.Phone, body.Country, body.City, body.Comment, body.Leverage,
		hashes[0], hashes[1], hashes[2], now).Scan(&login); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// the account row is not optional: a login without one has no money state
	if _, err := tx.Exec(ctx,
		`INSERT INTO hst.accounts (login, currency_digits, margin_leverage, updated_at)
		 VALUES ($1, 2, $2, $3)`, login, body.Leverage, now); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "user created",
		"actor", snap.Login, "target", login)

	return s.getUserByLogin(c, login, s.notifyUser(model.ActionCreated))
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

	// only the logins inside this manager's groups, the same masks the
	// websocket routes by
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

	// the previous group is needed before the write: the managers who lose the
	// record have to be told it left, and afterwards it is gone
	var oldGroup string
	if body.Group != nil {
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

	// a group change has two audiences: the managers who just lost the record
	// hear it left, the ones who gained it hear the update below
	if oldGroup != "" {
		s.NotifyWS(model.FamilyUsers, oldGroup, model.ActionMoved,
			ViewUserRef{Login: int64(login), Group: oldGroup})
	}

	return s.getUserByLogin(c, int64(login), s.notifyUser(model.ActionUpdated))
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

	// returning the group saves a read: it is needed to announce the delete
	// and it does not exist afterwards
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

	s.NotifyWS(model.FamilyUsers, gone, model.ActionDeleted,
		ViewUserRef{Login: int64(login), Group: gone})

	return s.App.HttpResponseNoContent(c)
}

// notifyUser announces a user change on the way out, so the record is loaded
// once and both the caller and the listeners get the same view.
//
// The group in the subject is the user's own group, which is what decides
// which managers hear about it.
func (s *HttpServer) notifyUser(action model.Action) func(*fiber.Ctx, interface{}) error {
	return func(c *fiber.Ctx, v interface{}) error {
		if u, ok := v.(*ViewUser); ok {
			s.NotifyWS(model.FamilyUsers, u.Group, action, u)
		}
		if action == model.ActionCreated {
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
