// Package mailer sends the platform's outgoing email through a configured mail server.
// It is not hst.mails, which is the internal mailbox between a trading account and support.
package mailer

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strings"
	"sync"
	"time"

	"hstserver/model"
	"hstserver/pkg/db"
	"hstserver/pkg/logger"
)

// ErrNoMailServer means nothing is configured to send from, so the mail stays queued.
var ErrNoMailServer = errors.New("no enabled default mail server is configured")

// Mailer drains hst.outbox through the default mail server.
type Mailer struct {
	DB  *db.PostgresDB
	Log *logger.Logger

	// TemplatesDir is the root holding greeting/, verify_email/ and the rest.
	TemplatesDir string

	// Interval is how often the queue is drained.
	Interval time.Duration

	stop chan struct{}
	once sync.Once
}

func New(database *db.PostgresDB, log *logger.Logger, templatesDir string, interval time.Duration) *Mailer {
	return &Mailer{
		DB:           database,
		Log:          log,
		TemplatesDir: templatesDir,
		Interval:     interval,
		stop:         make(chan struct{}),
	}
}

// Queue accepts a mail for delivery. It is written to the outbox rather than sent inline: a
// welcome mail carries the only copy of a generated password, so a dead SMTP server must delay
// it, never drop it. serverName is a mail_servers.name; empty uses the default server.
func (m *Mailer) Queue(ctx context.Context, recipient, subject, body string) error {
	return m.QueueWithServer(ctx, "", recipient, subject, body)
}

// QueueWithServer queues mail on the named server, or the default when name is empty.
func (m *Mailer) QueueWithServer(ctx context.Context, serverName, recipient, subject, body string) error {
	server, err := m.serverByName(ctx, serverName)
	if err != nil {
		return err
	}

	_, err = m.DB.DB.Exec(ctx,
		`INSERT INTO hst.outbox (mail_server_id, recipient, subject, body, state, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		server.MailServerId, recipient, subject, body, model.OutboxState_queued, time.Now().UnixNano())

	return err
}

// Start drains the queue until Stop.
func (m *Mailer) Start() {
	go func() {
		t := time.NewTicker(m.Interval)
		defer t.Stop()

		for {
			select {
			case <-m.stop:
				return
			case <-t.C:
				m.Drain(context.Background())
			}
		}
	}()
}

func (m *Mailer) Stop() { m.once.Do(func() { close(m.stop) }) }

// Drain sends every queued mail that has attempts left, each through the server it was
// queued on. A missing or disabled server falls back to the default; no default at all
// leaves the mail queued rather than dropped.
func (m *Mailer) Drain(ctx context.Context) {
	rows, err := m.DB.DB.Query(ctx,
		`SELECT outbox_id, mail_server_id, recipient, subject, body, attempts
		   FROM hst.outbox
		  WHERE state = $1 AND attempts < $2
		  ORDER BY outbox_id
		  LIMIT 100`, model.OutboxState_queued, model.OutboxMaxAttempts)
	if err != nil {
		return
	}

	type pending struct {
		id, serverId             int64
		recipient, subject, body string
		attempts                 int32
	}

	var batch []pending
	for rows.Next() {
		var p pending
		if err := rows.Scan(&p.id, &p.serverId, &p.recipient, &p.subject, &p.body, &p.attempts); err != nil {
			rows.Close()
			return
		}
		batch = append(batch, p)
	}
	rows.Close()

	// resolved once per distinct id for the batch; nil means nothing can send this one now
	servers := map[int64]*model.MailServer{}
	resolve := func(id int64) *model.MailServer {
		if s, ok := servers[id]; ok {
			return s
		}

		s := (*model.MailServer)(nil)
		if id != 0 {
			s = m.serverById(ctx, id)
		}
		if s == nil {
			if d, err := m.defaultServer(ctx); err == nil {
				s = d
			}
		}
		servers[id] = s

		return s
	}

	for _, p := range batch {
		server := resolve(p.serverId)
		if server == nil {
			continue
		}

		started := time.Now()
		sendErr := Send(server, p.recipient, p.subject, p.body)
		took := time.Since(started).Milliseconds()

		if sendErr != nil {
			m.failed(ctx, server.MailServerId, p.id, p.attempts+1, sendErr)
			continue
		}

		m.sent(ctx, server.MailServerId, p.id, took)
	}
}

func (m *Mailer) defaultServer(ctx context.Context) (*model.MailServer, error) {
	var s model.MailServer
	if err := m.DB.DB.QueryRow(ctx,
		`SELECT mail_server_id, sender_email, sender_name, smtp_server, smtp_login, smtp_password
		   FROM hst.mail_servers
		  WHERE enabled AND is_default LIMIT 1`).
		Scan(&s.MailServerId, &s.SenderEmail, &s.SenderName,
			&s.SmtpServer, &s.SmtpLogin, &s.SmtpPassword); err != nil {
		return nil, ErrNoMailServer
	}

	return &s, nil
}

// serverById resolves an enabled configuration by id; nil when it is missing or disabled.
func (m *Mailer) serverById(ctx context.Context, id int64) *model.MailServer {
	var s model.MailServer
	if err := m.DB.DB.QueryRow(ctx,
		`SELECT mail_server_id, sender_email, sender_name, smtp_server, smtp_login, smtp_password
		   FROM hst.mail_servers
		  WHERE enabled AND mail_server_id = $1`, id).
		Scan(&s.MailServerId, &s.SenderEmail, &s.SenderName,
			&s.SmtpServer, &s.SmtpLogin, &s.SmtpPassword); err != nil {
		return nil
	}

	return &s
}

// serverByName resolves a configured mail server by name; empty name uses the default server.
func (m *Mailer) serverByName(ctx context.Context, name string) (*model.MailServer, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return m.defaultServer(ctx)
	}

	var s model.MailServer
	err := m.DB.DB.QueryRow(ctx,
		`SELECT mail_server_id, sender_email, sender_name, smtp_server, smtp_login, smtp_password
		   FROM hst.mail_servers
		  WHERE enabled AND name = $1 LIMIT 1`, name).
		Scan(&s.MailServerId, &s.SenderEmail, &s.SenderName,
			&s.SmtpServer, &s.SmtpLogin, &s.SmtpPassword)
	if err != nil {
		return m.defaultServer(ctx)
	}

	return &s, nil
}

func (m *Mailer) sent(ctx context.Context, serverId, outboxId, tookMs int64) {
	if _, err := m.DB.DB.Exec(ctx,
		`UPDATE hst.outbox SET state = $2, sent_at = $3, last_error = ''
		  WHERE outbox_id = $1`,
		outboxId, model.OutboxState_sent, time.Now().UnixNano()); err != nil {
		return
	}

	// time_min starts at zero on a fresh row, so the first send has to seed it rather than min with 0
	_, _ = m.DB.DB.Exec(ctx,
		`UPDATE hst.mail_servers SET
		     total_sent  = total_sent + 1,
		     time_sum_ms = time_sum_ms + $2,
		     time_max_ms = GREATEST(time_max_ms, $2),
		     time_min_ms = CASE WHEN total_sent = 0 THEN $2 ELSE LEAST(time_min_ms, $2) END
		   WHERE mail_server_id = $1`, serverId, tookMs)
}

func (m *Mailer) failed(ctx context.Context, serverId, outboxId int64, attempts int32, cause error) {
	state := model.OutboxState_queued
	if attempts >= model.OutboxMaxAttempts {
		state = model.OutboxState_failed
	}

	_, _ = m.DB.DB.Exec(ctx,
		`UPDATE hst.outbox SET attempts = $2, last_error = $3, state = $4 WHERE outbox_id = $1`,
		outboxId, attempts, cause.Error(), state)

	_, _ = m.DB.DB.Exec(ctx,
		`UPDATE hst.mail_servers SET total_errors = total_errors + 1 WHERE mail_server_id = $1`,
		serverId)

	m.Log.Log(logger.TypeSys, logger.CodeErr, "mail send failed",
		"outbox_id", outboxId, "attempts", attempts, "error", cause.Error())
}

// Send delivers one mail synchronously. Port 465 is implicit TLS; 25 and 587 upgrade with
// STARTTLS when the server offers it.
func Send(s *model.MailServer, recipient, subject, body string) error {
	host, port, err := net.SplitHostPort(strings.TrimSpace(s.SmtpServer))
	if err != nil {
		return fmt.Errorf("mail server address %q: %w", s.SmtpServer, err)
	}

	client, err := dial(host, port)
	if err != nil {
		return err
	}
	defer func() { _ = client.Quit() }()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}); err != nil {
			return err
		}
	}

	if s.SmtpLogin != "" {
		if err := client.Auth(smtp.PlainAuth("", s.SmtpLogin, s.SmtpPassword, host)); err != nil {
			return err
		}
	}

	if err := client.Mail(s.SenderEmail); err != nil {
		return err
	}
	if err := client.Rcpt(recipient); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte(message(s, recipient, subject, body))); err != nil {
		return err
	}

	return w.Close()
}

func dial(host, port string) (*smtp.Client, error) {
	if port == "465" {
		conn, err := tls.Dial("tcp", net.JoinHostPort(host, port),
			&tls.Config{ServerName: host, MinVersion: tls.VersionTLS12})
		if err != nil {
			return nil, err
		}
		return smtp.NewClient(conn, host)
	}

	return smtp.Dial(net.JoinHostPort(host, port))
}

// message builds the RFC 5322 envelope. Templates are HTML, so the body is sent as HTML.
func message(s *model.MailServer, recipient, subject, body string) string {
	from := s.SenderEmail
	if s.SenderName != "" {
		from = fmt.Sprintf("%s <%s>", mime.QEncoding.Encode("utf-8", s.SenderName), s.SenderEmail)
	}

	var b strings.Builder
	b.WriteString("From: " + from + "\r\n")
	b.WriteString("To: " + recipient + "\r\n")
	b.WriteString("Subject: " + mime.QEncoding.Encode("utf-8", subject) + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
	b.WriteString(body)

	return b.String()
}
