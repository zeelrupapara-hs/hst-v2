package trader

import (
	v1 "hstserver/internal/server/v1"
	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/pkg/journal"
	"hstserver/pkg/logger"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
)

// The account types a signup may ask for. The type picks a configured group; a caller never
// names a path, or it could register itself into the managers tree.
const (
	AccountTypeDemo = "demo"
	AccountTypeReal = "real"
)

// CrtRegister is a public signup.
type CrtRegister struct {
	Type             string `json:"type" validate:"required,oneof=demo real"`
	Name             string `json:"name" validate:"required,max=128"`
	Email            string `json:"email" validate:"required,email,max=255"`
	Phone            string `json:"phone" validate:"max=64"`
	Country          string `json:"country" validate:"max=64"`
	City             string `json:"city" validate:"max=64"`
	PasswordMain     string `json:"password_main" validate:"required,min=8,max=128"`
	PasswordInvestor string `json:"password_investor" validate:"required,min=8,max=128"`
}

// ViewRegister is what a signup gets back: the login it may now authenticate with.
type ViewRegister struct {
	Login int64  `json:"login"`
	Group string `json:"group"`
	Type  string `json:"type"`
}

// Register opens an account from the public side.
//
//	@Id			Register
//	@Tags		Auth
//	@Accept		json
//	@Produce	json
//	@Param		body	body		CrtRegister	true	"the account to open"
//	@Success	201		{object}	Response{data=ViewRegister}
//	@Failure	400		{object}	Response
//	@Failure	429		{object}	Response
//	@Failure	500		{object}	Response
//	@Router		/auth/v1/register [post]
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

	login, status, err := s.OpenLogin(ctx, v1.NewLogin{
		Group:            group,
		Rights:           rights,
		Name:             body.Name,
		Email:            body.Email,
		Phone:            body.Phone,
		Country:          body.Country,
		City:             body.City,
		PasswordMain:     body.PasswordMain,
		PasswordInvestor: body.PasswordInvestor,
	})
	if err != nil {
		return s.App.HttpResponseStatus(c, status, err)
	}

	s.Log.Log(logger.TypeUser, logger.CodeOK, "account registered",
		"login", login, "group", group, "type", body.Type, "ip", ip)

	view := ViewRegister{Login: login, Group: group, Type: body.Type}

	// the managers whose access covers the tree see the signup arrive
	s.NotifyWS(model.SubjectUser(group), model.EventUserCreated, view)
	s.NotifySystem(model.SubjectSystemUserCreated, view)
	s.WriteJournal(ctx, login, ip, logger.CodeOK, journal.RegisteredMsg(login, body.Type), view)

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
