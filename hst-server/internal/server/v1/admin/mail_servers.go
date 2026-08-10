package admin

import (
	"errors"
	"net"
	"strconv"
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

// CrtMailServer creates one mail server configuration.
type CrtMailServer struct {
	Enabled      bool   `json:"enabled"`
	Name         string `json:"name" validate:"required,max=128"`
	SenderEmail  string `json:"sender_email" validate:"required,email,max=255"`
	SenderName   string `json:"sender_name" validate:"max=128"`
	SmtpServer   string `json:"smtp_server" validate:"required,max=255"`
	SmtpLogin    string `json:"smtp_login" validate:"max=255"`
	SmtpPassword string `json:"smtp_password" validate:"max=255"`
	IsDefault    bool   `json:"is_default"`
}

// UptMailServer patches one mail server configuration. An absent password keeps the stored one.
type UptMailServer struct {
	Enabled      *bool   `json:"enabled"`
	Name         *string `json:"name" validate:"omitempty,max=128"`
	SenderEmail  *string `json:"sender_email" validate:"omitempty,email,max=255"`
	SenderName   *string `json:"sender_name" validate:"omitempty,max=128"`
	SmtpServer   *string `json:"smtp_server" validate:"omitempty,max=255"`
	SmtpLogin    *string `json:"smtp_login" validate:"omitempty,max=255"`
	SmtpPassword *string `json:"smtp_password" validate:"omitempty,max=255"`
	IsDefault    *bool   `json:"is_default"`
}

// ViewMailServer is one configuration plus the send statistics the list shows.
type ViewMailServer struct {
	model.MailServer
	TimeAvgMs int64 `json:"time_avg_ms"`
	Queue     int64 `json:"queue"`
}

// mailServersSortable are the real columns of mail_servers.
var mailServersSortable = utils.NewSortable(
	"mail_server_id", "name", "sender_email", "enabled", "is_default", "total_sent", "created_at")

// generated from the struct so the select and the scan cannot drift apart
const mailServerColumns = `mail_server_id, enabled, name, sender_email, sender_name,
	smtp_server, smtp_login, smtp_password, is_default,
	total_sent, total_errors, time_min_ms, time_max_ms, time_sum_ms, created_at, updated_at`

func scanMailServer(row pgx.Row, out *model.MailServer) error {
	return row.Scan(&out.MailServerId, &out.Enabled, &out.Name, &out.SenderEmail, &out.SenderName,
		&out.SmtpServer, &out.SmtpLogin, &out.SmtpPassword, &out.IsDefault,
		&out.TotalSent, &out.TotalErrors, &out.TimeMinMs, &out.TimeMaxMs, &out.TimeSumMs,
		&out.CreatedAt, &out.UpdatedAt)
}

// checkSmtpServer accepts what the dialog accepts: host:port, ports 25, 465 and 587.
func checkSmtpServer(addr string) error {
	host, port, err := net.SplitHostPort(strings.TrimSpace(addr))
	if err != nil || host == "" {
		return errs.ErrInvalidSmtpServer
	}

	n, err := strconv.Atoi(port)
	if err != nil || n <= 0 || n > 65535 {
		return errs.ErrInvalidSmtpPort
	}

	return nil
}

// ListMailServers returns the configurations with their send statistics.
//
//	@Id			ListMailServers
//	@Tags		MailServers
//	@Produce	json
//	@Success	200	{object}	Response{data=[]ViewMailServer}
//	@Failure	400	{object}	Response
//	@Failure	403	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/mail-servers [get]
func (s *Server) ListMailServers(c *fiber.Ctx) error {
	ctx := c.UserContext()

	q, err := utils.QueryFilter(c, mailServersSortable, "mail_server_id")
	if err != nil {
		return s.App.HttpResponseBadQueryParams(c, err)
	}

	// sort_by is validated against an allowlist in QueryFilter.
	rows, err := s.DB.DB.Query(ctx,
		`SELECT `+mailServerColumns+`
		   FROM hst.mail_servers
		  WHERE ($1 = '' OR name ILIKE '%'||$1||'%' OR sender_email ILIKE '%'||$1||'%')
		  ORDER BY `+q.SortBy+`
		  LIMIT $2 OFFSET $3`, q.Search, q.Limit, q.Offset)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	out := make([]ViewMailServer, 0)
	for rows.Next() {
		var m model.MailServer
		if err := scanMailServer(rows, &m); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		out = append(out, ViewMailServer{MailServer: m, TimeAvgMs: m.TimeAvgMs()})
	}
	if err := rows.Err(); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// one grouped count rather than a query per configuration
	queued := map[int64]int64{}
	qrows, err := s.DB.DB.Query(ctx,
		`SELECT mail_server_id, COUNT(*) FROM hst.outbox WHERE state = $1 GROUP BY mail_server_id`,
		model.OutboxState_queued)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer qrows.Close()

	for qrows.Next() {
		var id, n int64
		if err := qrows.Scan(&id, &n); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		queued[id] = n
	}

	for i := range out {
		out[i].Queue = queued[out[i].MailServerId]
	}

	return s.App.HttpResponseOK(c, out)
}

// GetMailServer returns one configuration.
//
//	@Id			GetMailServer
//	@Tags		MailServers
//	@Produce	json
//	@Param		id	path		int	true	"mail server id"
//	@Success	200	{object}	Response{data=ViewMailServer}
//	@Failure	404	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/mail-servers/{id} [get]
func (s *Server) GetMailServer(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}

	var out model.MailServer
	if err := scanMailServer(s.DB.DB.QueryRow(c.UserContext(),
		`SELECT `+mailServerColumns+` FROM hst.mail_servers WHERE mail_server_id = $1`, id),
		&out); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, ViewMailServer{MailServer: out, TimeAvgMs: out.TimeAvgMs()})
}

// CreateMailServer adds a configuration.
//
//	@Id			CreateMailServer
//	@Tags		MailServers
//	@Accept		json
//	@Produce	json
//	@Param		body	body		CrtMailServer	true	"the mail server to create"
//	@Success	201		{object}	Response{data=ViewMailServer}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/mail-servers [post]
func (s *Server) CreateMailServer(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var body CrtMailServer
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}
	if err := checkSmtpServer(body.SmtpServer); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// only one configuration is the default, so raising this one stands the old one down
	if body.IsDefault {
		if _, err := tx.Exec(ctx,
			`UPDATE hst.mail_servers SET is_default = FALSE WHERE is_default`); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
	}

	now := time.Now().UnixNano()

	var out model.MailServer
	if err := scanMailServer(tx.QueryRow(ctx,
		`INSERT INTO hst.mail_servers (enabled, name, sender_email, sender_name,
		     smtp_server, smtp_login, smtp_password, is_default, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$9)
		 RETURNING `+mailServerColumns,
		body.Enabled, body.Name, body.SenderEmail, body.SenderName,
		strings.TrimSpace(body.SmtpServer), body.SmtpLogin, body.SmtpPassword,
		body.IsDefault, now), &out); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	view := ViewMailServer{MailServer: out, TimeAvgMs: out.TimeAvgMs()}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "mail server created",
		"actor", snap.Login, "target", out.MailServerId,
		"name", out.Name, "smtp_server", out.SmtpServer, "default", out.IsDefault)

	s.NotifyWS(model.SubjectMailServer, model.EventMailServerCreated, view)
	s.NotifySystem(model.SubjectSystemMailServerCreated, view)
	s.JournalEntry(c, model.JournalType_mail, logger.CodeOK,
		journal.MailServerCreatedMsg(int(out.MailServerId)), view)

	return s.App.HttpResponseCreated(c, view)
}

// UpdateMailServer patches a configuration. An absent password keeps the stored one, so the
// panel can save the form without ever having been shown the secret.
//
//	@Id			UpdateMailServer
//	@Tags		MailServers
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int				true	"mail server id"
//	@Param		body	body		UptMailServer	true	"the fields to change"
//	@Success	200		{object}	Response{data=ViewMailServer}
//	@Failure	400		{object}	Response
//	@Failure	404		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/mail-servers/{id} [patch]
func (s *Server) UpdateMailServer(c *fiber.Ctx) error {
	ctx := c.UserContext()

	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}

	var body UptMailServer
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}
	if body.SmtpServer != nil {
		if err := checkSmtpServer(*body.SmtpServer); err != nil {
			return s.App.HttpResponseBadRequest(c, err)
		}
	}

	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// only one configuration is the default, so raising this one stands the old one down
	if body.IsDefault != nil && *body.IsDefault {
		if _, err := tx.Exec(ctx,
			`UPDATE hst.mail_servers SET is_default = FALSE WHERE is_default AND mail_server_id <> $1`,
			id); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
	}

	var out model.MailServer
	if err := scanMailServer(tx.QueryRow(ctx,
		`UPDATE hst.mail_servers SET
		     enabled       = COALESCE($2, enabled),
		     name          = COALESCE($3, name),
		     sender_email  = COALESCE($4, sender_email),
		     sender_name   = COALESCE($5, sender_name),
		     smtp_server   = COALESCE($6, smtp_server),
		     smtp_login    = COALESCE($7, smtp_login),
		     smtp_password = COALESCE($8, smtp_password),
		     is_default    = COALESCE($9, is_default),
		     updated_at    = $10
		   WHERE mail_server_id = $1
		 RETURNING `+mailServerColumns,
		id, body.Enabled, body.Name, body.SenderEmail, body.SenderName,
		body.SmtpServer, body.SmtpLogin, body.SmtpPassword, body.IsDefault,
		time.Now().UnixNano()), &out); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	view := ViewMailServer{MailServer: out, TimeAvgMs: out.TimeAvgMs()}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "mail server updated",
		"actor", snap.Login, "target", out.MailServerId, "name", out.Name)

	s.NotifyWS(model.SubjectMailServer, model.EventMailServerUpdated, view)
	s.NotifySystem(model.SubjectSystemMailServerUpdated, view)
	s.JournalEntry(c, model.JournalType_mail, logger.CodeOK,
		journal.MailServerUpdatedMsg(int(out.MailServerId)), view)

	return s.App.HttpResponseOK(c, view)
}

// DeleteMailServer removes a configuration. Queued mail addressed to it is left alone; the
// sender falls back to the default server rather than dropping a password nobody else holds.
//
//	@Id			DeleteMailServer
//	@Tags		MailServers
//	@Produce	json
//	@Param		id	path		int	true	"mail server id"
//	@Success	200	{object}	Response
//	@Failure	404	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/mail-servers/{id} [delete]
func (s *Server) DeleteMailServer(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}

	tag, err := s.DB.DB.Exec(c.UserContext(),
		`DELETE FROM hst.mail_servers WHERE mail_server_id = $1`, id)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if tag.RowsAffected() == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	view := model.MailServer{MailServerId: int64(id)}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "mail server deleted",
		"actor", snap.Login, "target", id)

	s.NotifyWS(model.SubjectMailServer, model.EventMailServerDeleted, view)
	s.NotifySystem(model.SubjectSystemMailServerDeleted, view)
	s.JournalEntry(c, model.JournalType_mail, logger.CodeOK, journal.MailServerDeletedMsg(id), view)

	return s.App.HttpResponseOK(c, view)
}
