package admin

import (
	"time"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/pkg/mailer"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
)

// CrtMailTemplate saves a compose template. Saving under a taken name overwrites, so the
// dialog's Save always lands without a rename dance.
type CrtMailTemplate struct {
	Name    string `json:"name" validate:"required,max=128"`
	Subject string `json:"subject" validate:"max=128"`
	Body    string `json:"body" validate:"max=65536"`
}

// ListMailTemplates returns the caller's own templates.
//
//	@Id			ListMailTemplates
//	@Tags		Mails
//	@Produce	json
//	@Success	200	{object}	Response{data=[]model.MailTemplate}
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/mail-templates [get]
func (s *Server) ListMailTemplates(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	rows, err := s.DB.DB.Query(c.UserContext(),
		`SELECT template_id, manager_login, name, subject, body, created_at, updated_at
		   FROM hst.mail_templates
		  WHERE manager_login = $1
		  ORDER BY name`, snap.Login)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	out := make([]model.MailTemplate, 0)
	for rows.Next() {
		var t model.MailTemplate
		if err := rows.Scan(&t.TemplateId, &t.ManagerLogin, &t.Name, &t.Subject, &t.Body,
			&t.CreatedAt, &t.UpdatedAt); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, out)
}

// SaveMailTemplate creates or overwrites one of the caller's templates.
//
//	@Id			SaveMailTemplate
//	@Tags		Mails
//	@Accept		json
//	@Produce	json
//	@Param		body	body		CrtMailTemplate	true	"the template to save"
//	@Success	200		{object}	Response{data=model.MailTemplate}
//	@Failure	400		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/mail-templates [post]
func (s *Server) SaveMailTemplate(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	var in CrtMailTemplate
	if err := c.BodyParser(&in); err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}
	if err := s.Validate.Struct(in); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	now := time.Now().UnixNano()

	var out model.MailTemplate
	if err := s.DB.DB.QueryRow(c.UserContext(),
		`INSERT INTO hst.mail_templates (manager_login, name, subject, body, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $5)
		 ON CONFLICT (manager_login, name)
		 DO UPDATE SET subject = EXCLUDED.subject, body = EXCLUDED.body, updated_at = EXCLUDED.updated_at
		 RETURNING template_id, manager_login, name, subject, body, created_at, updated_at`,
		snap.Login, in.Name, in.Subject, mailer.SanitizeHTML(in.Body), now).
		Scan(&out.TemplateId, &out.ManagerLogin, &out.Name, &out.Subject, &out.Body,
			&out.CreatedAt, &out.UpdatedAt); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, out)
}

// DeleteMailTemplate removes one of the caller's templates.
//
//	@Id			DeleteMailTemplate
//	@Tags		Mails
//	@Produce	json
//	@Param		id	path		int	true	"template id"
//	@Success	200	{object}	Response
//	@Failure	404	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/mail-templates/{id} [delete]
func (s *Server) DeleteMailTemplate(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}

	tag, err := s.DB.DB.Exec(c.UserContext(),
		`DELETE FROM hst.mail_templates WHERE template_id = $1 AND manager_login = $2`,
		id, snap.Login)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if tag.RowsAffected() == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	return s.App.HttpResponseOK(c, model.MailTemplate{TemplateId: int64(id)})
}
