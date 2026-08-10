package trader

import (
	"context"
	"strconv"

	v1 "hstserver/internal/server/v1"
	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/pkg/journal"
	"hstserver/pkg/logger"
	"hstserver/pkg/mailer"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
)

// The account types a signup may ask for. The type picks a configured group; a caller never
// names a path, or it could register itself into the managers tree.
const (
	AccountTypeDemo = "demo"
	AccountTypeReal = "real"
)

// CrtRegister is a public signup. It carries no password: MT5 generates both and mails them,
// so a signup cannot choose a weak one and no plaintext ever crosses the wire inbound.
type CrtRegister struct {
	Type    string `json:"type" validate:"required,oneof=demo real"`
	Name    string `json:"name" validate:"required,max=128"`
	Email   string `json:"email" validate:"required,email,max=255"`
	Phone   string `json:"phone" validate:"max=64"`
	Country string `json:"country" validate:"max=64"`
	City    string `json:"city" validate:"max=64"`
}

// generatedPasswordLength is what a welcome mail carries; long enough that it is worth mailing
// rather than remembering.
const generatedPasswordLength = 12

// ViewRegister is what a signup gets back: the login it may now authenticate with.
type ViewRegister struct {
	Login int64  `json:"login"`
	Group string `json:"group"`
	Type  string `json:"type"`
}

// Register opens an account from the public side.
//
//	@Id			Register
//	@Tags		Trader
//	@Accept		json
//	@Produce	json
//	@Param		body	body		CrtRegister	true	"the account to open"
//	@Success	201		{object}	Response{data=ViewRegister}
//	@Failure	400		{object}	Response
//	@Failure	429		{object}	Response
//	@Failure	500		{object}	Response
//	@Router		/auth/trader/v1/register [post]
func (s *Server) Register(c *fiber.Ctx) error {
	ctx := c.UserContext()
	ip := utils.GetRealIP(c)

	// turned away before argon2 runs, so a signup flood costs a lookup rather than a hash
	if s.OAuth2.IPThrottled(ctx, ip) {
		return s.App.HttpResponseTooManyRequests(c, errs.ErrTooManyRequests)
	}

	var body CrtRegister
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	group, err := s.RegistrationGroup(body.Type)
	if err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	// a signup is enabled and owns its password, and nothing else: rights are not the caller's to choose
	rights := int64(model.UsersRights_enabled | model.UsersRights_password)

	// a group demanding more than the mail length gets a longer password, not a rejected signup
	length := generatedPasswordLength
	if min := int(s.GroupPasswordMin(ctx, 0, group)); min > length {
		length = min
	}

	passwordMain, err := utils.NewPassword(length)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	passwordInvestor, err := utils.NewPassword(length)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	login, status, err := s.OpenLogin(ctx, v1.NewLogin{
		Group:            group,
		Rights:           rights,
		Name:             body.Name,
		Email:            body.Email,
		Phone:            body.Phone,
		Country:          body.Country,
		City:             body.City,
		PasswordMain:     passwordMain,
		PasswordInvestor: passwordInvestor,
	})
	if err != nil {
		return s.App.HttpResponseStatus(c, status, err)
	}

	s.Log.Log(logger.TypeUser, logger.CodeOK, "account registered",
		"login", login, "group", group, "type", body.Type, "ip", ip)

	s.SendWelcome(ctx, login, group, body.Name, body.Email, passwordMain, passwordInvestor)

	view := ViewRegister{Login: login, Group: group, Type: body.Type}

	// the managers whose access covers the tree see the signup arrive
	s.NotifyWS(model.SubjectUser(group), model.EventUserCreated, view)
	s.NotifySystem(model.SubjectSystemUserCreated, view)
	s.WriteJournal(ctx, login, ip, model.JournalType_accounts, logger.CodeOK, journal.RegisteredMsg(body.Type), view)

	return s.App.HttpResponseCreated(c, view)
}

// RegistrationGroup is where a signup of this type lands.
func (s *Server) RegistrationGroup(accountType string) (string, error) {
	var group string
	switch accountType {
	case AccountTypeDemo:
		group = s.Cfg.Register.DemoGroup
	case AccountTypeReal:
		// a real signup is preliminary until somebody approves it and moves it to the live tree
		group = s.Cfg.Register.PreliminaryGroup
	default:
		return "", errs.ErrBadRequest
	}

	// blank disables the type rather than registering into an empty path
	if group == "" {
		return "", errs.ErrRegistrationClosed
	}

	return group, nil
}

// SendWelcome queues the mail carrying the generated credentials. MT5 sends it however the
// account was opened, and it is the only place those passwords exist in plaintext.
func (s *Server) SendWelcome(ctx context.Context, login int64, group, name, email, main, investor string) {
	// development only, and the sole way to read a password when no mail server is configured
	if s.Cfg.Auth.LogCredentials {
		s.Log.Log(logger.TypeUser, logger.CodeWarn, "generated account credentials",
			"login", login, "password_main", main, "password_investor", investor)
	}

	if email == "" {
		return
	}

	company, catalog := s.CompanyOf(ctx, group)

	body, err := s.Mailer.Render(mailer.KindGreeting, catalog, map[string]string{
		mailer.MacroName:             name,
		mailer.MacroLogin:            strconv.FormatInt(login, 10),
		mailer.MacroPasswordMain:     main,
		mailer.MacroPasswordInvestor: investor,
		mailer.MacroGroup:            group,
		mailer.MacroCompany:          company,
	})
	if err != nil {
		s.Log.Log(logger.TypeUser, logger.CodeErr, "welcome mail not rendered",
			"login", login, "error", err.Error())
		return
	}

	if err := s.Mailer.Queue(ctx, email, "Your trading account", body); err != nil {
		s.Log.Log(logger.TypeUser, logger.CodeErr, "welcome mail not queued",
			"login", login, "error", err.Error())
	}
}

// CompanyOf reads the group's company name and its template folder. Both are blank for a
// group that never set them, which lands the caller on the default templates.
func (s *Server) CompanyOf(ctx context.Context, group string) (company, catalog string) {
	_ = s.DB.DB.QueryRow(ctx,
		`SELECT company, company_catalog FROM hst.groups WHERE "group" = $1`, group).
		Scan(&company, &catalog)

	return company, catalog
}
