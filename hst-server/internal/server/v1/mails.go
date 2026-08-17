package v1

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"hstserver/model"
	"hstserver/pkg/cache"
	errs "hstserver/pkg/errors"
	"hstserver/pkg/journal"
	"hstserver/pkg/logger"
	"hstserver/pkg/mailer"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Mail folders.
const (
	MailFolderInbox  = 1
	MailFolderOutbox = 2
	MailFolderDraft  = 3
	MailFolderBin    = 4
)

// errNoSupportAccount is returned rather than guessing a recipient, as v1 did by hardcoding one.
var (
	errNoSupportAccount = errors.New("no support account is configured, grant a manager the techsupport right")
	errNotADraft        = errors.New("only a draft can be edited")
	errEmptyMail        = errors.New("a mail needs a subject or a body")
)

// ViewMail is one message as its owner sees it.
type ViewMail struct {
	MailId         int64  `json:"mail_id"`
	TrackingId     string `json:"tracking_id"`
	SenderLogin    int64  `json:"sender_login"`
	RecipientLogin int64  `json:"recipient_login"`
	Subject        string `json:"subject"`
	Body           string `json:"body"`
	Folder         int32  `json:"folder"`
	ReadAt         int64  `json:"read_at"`
	CreatedAt      int64  `json:"created_at"`
	UpdatedAt      int64  `json:"updated_at"`
}

// BodyMail is a mail a trader sends or saves.
type BodyMail struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
	Draft   bool   `json:"draft"`
}

const mailColumns = `m.mail_id, m.tracking_id, m.sender_login, m.recipient_login, m.subject,
	m.body, m.folder, m.read_at, m.created_at, m.updated_at`

// readMails is the one query behind every mail read handler.
func (s *HttpServer) readMails(ctx context.Context, where string, args []any, p pageOpts) ([]ViewMail, error) {
	where, args = p.bound("m.created_at", where, args)

	rows, err := s.DB.DB.Query(ctx,
		`SELECT `+mailColumns+` FROM hst.mails m WHERE `+where+p.tail("m.mail_id"), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []ViewMail{}
	for rows.Next() {
		var v ViewMail
		if err := rows.Scan(&v.MailId, &v.TrackingId, &v.SenderLogin, &v.RecipientLogin,
			&v.Subject, &v.Body, &v.Folder, &v.ReadAt, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}

	return out, rows.Err()
}

// supportLogin is the account a trader's mail is addressed to.
func (s *HttpServer) supportLogin(ctx context.Context) (int64, error) {
	var login int64
	err := s.DB.DB.QueryRow(ctx,
		`SELECT login FROM hst.managers WHERE right_techsupport = 1 ORDER BY login LIMIT 1`).Scan(&login)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, errNoSupportAccount
	}

	return login, err
}

// mailFolderWhere maps a folder name to the rows that belong to the caller there.
func mailFolderWhere(typ string) (string, error) {
	switch strings.ToLower(typ) {
	case "inbox":
		return "m.recipient_login = $1 AND m.folder = 1", nil
	case "outbox":
		return "m.sender_login = $1 AND m.folder = 2", nil
	case "draft":
		return "m.sender_login = $1 AND m.folder = 3", nil
	case "bin":
		return "m.folder = 4 AND (m.sender_login = $1 OR m.recipient_login = $1)", nil
	default:
		return "", errors.New("type must be inbox, outbox, draft or bin")
	}
}

// GetMyMails lists the caller's mail in one folder.
//
//	@Id			GetMyMails
//	@Tags		Trader
//	@Produce	json
//	@Param		type	query		string	true	"folder"	Enums(inbox, outbox, draft, bin)
//	@Param		limit	query		int		false	"how many, newest first"
//	@Param		page	query		int		false	"which page, zero based"
//	@Param		from	query		int		false	"unix seconds, inclusive"
//	@Param		to		query		int		false	"unix seconds, exclusive"
//	@Success	200		{object}	Response{data=[]ViewMail}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/mails [get]
func (s *HttpServer) GetMyMails(c *fiber.Ctx) error {
	where, err := mailFolderWhere(c.Query("type"))
	if err != nil {
		return s.App.HttpResponseBadQueryParams(c, err)
	}

	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	out, err := s.readMails(c.UserContext(), where, []any{snap.Login}, readPage(c, 100))
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, out)
}

// GetMyMail reads one message, and marks it read when the caller is the one it was addressed to.
//
//	@Id			GetMyMail
//	@Tags		Trader
//	@Produce	json
//	@Param		tracking_id	path		string	true	"the mail"
//	@Success	200			{object}	Response{data=ViewMail}
//	@Failure	404			{object}	Response
//	@Failure	500			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/mails/{tracking_id} [get]
func (s *HttpServer) GetMyMail(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	out, err := s.readMails(c.UserContext(),
		"m.tracking_id = $1 AND (m.sender_login = $2 OR m.recipient_login = $2)",
		[]any{c.Params("tracking_id"), snap.Login}, pageOpts{limit: 1})
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if len(out) == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	v := out[0]
	if v.RecipientLogin == snap.Login && v.ReadAt == 0 {
		v.ReadAt = time.Now().UnixNano()
		if _, err := s.DB.DB.Exec(c.UserContext(),
			`UPDATE hst.mails SET read_at = $1, updated_at = $1 WHERE mail_id = $2`,
			v.ReadAt, v.MailId); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		v.UpdatedAt = v.ReadAt
	}

	return s.App.HttpResponseOK(c, v)
}

// SendMyMail sends a message to support, or saves it as a draft.
//
//	@Id			SendMyMail
//	@Tags		Trader
//	@Accept		json
//	@Produce	json
//	@Param		body	body		BodyMail	true	"the message"
//	@Success	200		{object}	Response{data=ViewMail}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/mails [post]
func (s *HttpServer) SendMyMail(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	var in BodyMail
	if err := c.BodyParser(&in); err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}
	if strings.TrimSpace(in.Subject) == "" && strings.TrimSpace(in.Body) == "" {
		return s.App.HttpResponseBadRequest(c, errEmptyMail)
	}

	// the recipient is looked up, never assumed: v1 hardcoded an admin id and broke on every other install
	recipient, err := s.supportLogin(c.UserContext())
	if err != nil {
		if errors.Is(err, errNoSupportAccount) {
			return s.App.HttpResponseBadRequest(c, err)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	now := time.Now().UnixNano()
	folder := MailFolderOutbox
	if in.Draft {
		folder = MailFolderDraft
	}

	v := ViewMail{
		TrackingId:     uuid.NewString(),
		SenderLogin:    snap.Login,
		RecipientLogin: recipient,
		Subject:        in.Subject,
		Body:           in.Body,
		Folder:         int32(folder),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	tx, err := s.DB.DB.Begin(c.UserContext())
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer func() { _ = tx.Rollback(c.UserContext()) }()

	if err := tx.QueryRow(c.UserContext(),
		`INSERT INTO hst.mails (tracking_id, sender_login, recipient_login, subject, body,
		        folder, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $7) RETURNING mail_id`,
		v.TrackingId, v.SenderLogin, v.RecipientLogin, v.Subject, v.Body, v.Folder, now).
		Scan(&v.MailId); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// a sent mail is two rows, so each side bins its own copy without touching the other's
	inbox := v
	if !in.Draft {
		inbox.TrackingId = uuid.NewString()
		inbox.Folder = MailFolderInbox
		if err := tx.QueryRow(c.UserContext(),
			`INSERT INTO hst.mails (tracking_id, sender_login, recipient_login, subject, body,
			        folder, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $7) RETURNING mail_id`,
			inbox.TrackingId, inbox.SenderLogin, inbox.RecipientLogin, inbox.Subject, inbox.Body,
			inbox.Folder, now).Scan(&inbox.MailId); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
	}

	if err := tx.Commit(c.UserContext()); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if !in.Draft {
		s.notifyMailInbox(&inbox)
	}

	return s.App.HttpResponseOK(c, v)
}

// UpdateMyDraft edits an unsent message.
//
//	@Id			UpdateMyDraft
//	@Tags		Trader
//	@Accept		json
//	@Produce	json
//	@Param		tracking_id	path		string		true	"the draft"
//	@Param		body		body		BodyMail	true	"the message"
//	@Success	200			{object}	Response{data=ViewMail}
//	@Failure	400			{object}	Response
//	@Failure	404			{object}	Response
//	@Failure	500			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/mails/{tracking_id} [put]
func (s *HttpServer) UpdateMyDraft(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	var in BodyMail
	if err := c.BodyParser(&in); err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}

	out, err := s.readMails(c.UserContext(), "m.tracking_id = $1 AND m.sender_login = $2",
		[]any{c.Params("tracking_id"), snap.Login}, pageOpts{limit: 1})
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if len(out) == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if out[0].Folder != MailFolderDraft {
		return s.App.HttpResponseBadRequest(c, errNotADraft)
	}

	v := out[0]
	v.Subject, v.Body, v.UpdatedAt = in.Subject, in.Body, time.Now().UnixNano()

	if _, err := s.DB.DB.Exec(c.UserContext(),
		`UPDATE hst.mails SET subject = $1, body = $2, updated_at = $3 WHERE mail_id = $4`,
		v.Subject, v.Body, v.UpdatedAt, v.MailId); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, v)
}

// BodySendMail is a manager's message to a set of accounts: a To expression, or the older
// logins / group_mask pair. Internal defaults to on so callers predating the email channel
// keep their behavior.
type BodySendMail struct {
	Logins       []int64 `json:"logins"`
	GroupMask    string  `json:"group_mask" validate:"max=128"`
	To           string  `json:"to" validate:"max=4096"`
	Subject      string  `json:"subject" validate:"required,max=128"`
	Body         string  `json:"body" validate:"required,max=65536"`
	Internal     *bool   `json:"internal"`
	Email        bool    `json:"email"`
	MailServerId int64   `json:"mail_server_id"`
}

// BodyPreviewMail is the recipient half of BodySendMail, for the count the dialog shows.
type BodyPreviewMail struct {
	To        string  `json:"to" validate:"max=4096"`
	Logins    []int64 `json:"logins"`
	GroupMask string  `json:"group_mask" validate:"max=128"`
}

var (
	errNoRecipients          = errors.New("no accounts match the given recipients")
	errNoChannel             = errors.New("select internal mail, email, or both")
	errMailServerUnavailable = errors.New("the chosen mail server is missing or disabled")
)

// mailRecipientsWhere resolves who a mail names to a predicate the caller's masks allow. The
// To expression always narrows the caller's own reach, so it cannot widen access.
func mailRecipientsWhere(snap *cache.Session, to string, logins []int64, groupMask string) (string, []any, error) {
	where, args := groupWhere(snap.IsManager, snap.ManagerGroups, 1)

	switch {
	case strings.TrimSpace(to) != "":
		groups, err := parseMailTo(to)
		if err != nil {
			return "", nil, err
		}
		expr, exprArgs := mailToWhere(groups, len(args)+1)
		where += " AND " + expr
		args = append(args, exprArgs...)

	case len(logins) > 0:
		where += ` AND u.login = ANY($` + strconv.Itoa(len(args)+1) + `)`
		args = append(args, logins)

	case groupMask != "":
		where += ` AND u."group" LIKE $` + strconv.Itoa(len(args)+1)
		args = append(args, strings.ReplaceAll(groupMask, "*", "%"))

	default:
		return "", nil, errors.New("to, logins or group_mask is required")
	}

	return where, args, nil
}

// sendMailRecipients loads the recipients with everything the macros can name. The money
// figures join the hst.accounts snapshot: exact for a flat account, as of the last trade
// event otherwise.
func (s *HttpServer) sendMailRecipients(ctx context.Context, where string, args []any) ([]mailRecipient, error) {
	rows, err := s.DB.DB.Query(ctx,
		`SELECT u.login, u.email, u.name, u."group", u.leverage,
		        COALESCE(g.currency, ''), COALESCE(g.currency_digits, 2),
		        COALESCE(a.balance, 0), COALESCE(a.credit, 0), COALESCE(a.equity, 0),
		        COALESCE(a.margin, 0), COALESCE(a.margin_free, 0), COALESCE(a.margin_level, 0)
		   FROM hst.users u
		   LEFT JOIN hst.accounts a ON a.login = u.login
		   LEFT JOIN hst.groups g ON g."group" = u."group"
		  WHERE `+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []mailRecipient{}
	for rows.Next() {
		var r mailRecipient
		if err := rows.Scan(&r.Login, &r.Email, &r.Name, &r.Group, &r.Leverage,
			&r.Currency, &r.CurrencyDigits,
			&r.Balance, &r.Credit, &r.Equity,
			&r.Margin, &r.MarginFree, &r.MarginLevel); err != nil {
			return nil, err
		}
		out = append(out, r)
	}

	return out, rows.Err()
}

// notifyMailInbox pushes a delivered mail to the recipient's open terminals, so the inbox
// updates without a refetch. The payload is the same ViewMail the REST fetch returns.
func (s *HttpServer) notifyMailInbox(v *ViewMail) {
	payload, err := json.Marshal(v)
	if err != nil {
		return
	}

	for _, c := range s.Hub.Login(v.RecipientLogin) {
		c.Send(&model.Event{
			Type: model.EventMailInbox, Format: model.FormatJSON, Payload: payload, At: v.CreatedAt,
		})
	}
}

// mailServerFor resolves the picked server to a queueable id: zero means the default.
func (s *HttpServer) mailServerFor(ctx context.Context, id int64) (int64, error) {
	if id == 0 {
		var out int64
		err := s.DB.DB.QueryRow(ctx,
			`SELECT mail_server_id FROM hst.mail_servers WHERE enabled AND is_default LIMIT 1`).
			Scan(&out)
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, mailer.ErrNoMailServer
		}
		return out, err
	}

	var out int64
	err := s.DB.DB.QueryRow(ctx,
		`SELECT mail_server_id FROM hst.mail_servers WHERE enabled AND mail_server_id = $1`, id).
		Scan(&out)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, errMailServerUnavailable
	}

	return out, err
}

// SendMail delivers a manager's message over the internal mailbox, by email, or both.
//
// The To expression is orGroups separated by ';', each a ',' list of terms: a login, a range
// lo-hi, group:mask, country:name or city:name, any of them negated with '!'. Same-kind terms
// widen each other, different kinds narrow, and '!' always narrows. The subject and body may
// carry #LOGIN#, #USERNAME#, #USER_CURRENCY#, #USER_BALANCE#, #USER_CREDIT#, #USER_EQUITY#,
// #USER_LEVERAGE#, #USER_MARGIN#, #USER_MARGIN_FREE# and #USER_MARGIN_LEVEL#, substituted per
// recipient; the money figures read the account snapshot, so an account with open positions
// reads as of its last trade event. The body is sanitized to the markup the compose toolbar
// produces before anything is stored.
//
//	@Id			SendMail
//	@Tags		Mails
//	@Accept		json
//	@Produce	json
//	@Param		body	body		BodySendMail	true	"the message, its channels and who gets it"
//	@Success	200		{object}	Response
//	@Failure	400		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/mails [post]
func (s *HttpServer) SendMail(c *fiber.Ctx) error {
	ctx := c.UserContext()

	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	var in BodySendMail
	if err := c.BodyParser(&in); err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}
	if err := s.Validate.Struct(in); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	internal := in.Internal == nil || *in.Internal
	if !internal && !in.Email {
		return s.App.HttpResponseBadRequest(c, errNoChannel)
	}

	where, args, err := mailRecipientsWhere(snap, in.To, in.Logins, in.GroupMask)
	if err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	recipients, err := s.sendMailRecipients(ctx, where, args)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if len(recipients) == 0 {
		// MT5 parity: an empty delivery is refused and the refusal is on the record
		s.JournalEntry(c, model.JournalType_mail, logger.CodeErr,
			journal.MailNoRecipientsMsg(in.Subject), in)
		return s.App.HttpResponseBadRequest(c, errNoRecipients)
	}

	var mailServerId int64
	if in.Email {
		mailServerId, err = s.mailServerFor(ctx, in.MailServerId)
		if err != nil {
			if errors.Is(err, mailer.ErrNoMailServer) || errors.Is(err, errMailServerUnavailable) {
				return s.App.HttpResponseBadRequest(c, err)
			}
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
	}

	// sanitized once here, then per-recipient macro values are escaped on the way in
	body := mailer.SanitizeHTML(in.Body)

	now := time.Now().UnixNano()
	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var (
		sent, queued, skipped int
		delivered             []ViewMail
		// refs mirrors the batch: an index into delivered for a mails insert, -1 for outbox
		refs []int
	)

	batch := &pgx.Batch{}
	for i := range recipients {
		r := &recipients[i]
		subject := expandMailMacros(in.Subject, r, false)
		expanded := expandMailMacros(body, r, true)

		if internal {
			v := ViewMail{
				TrackingId:     uuid.NewString(),
				SenderLogin:    snap.Login,
				RecipientLogin: r.Login,
				Subject:        subject,
				Body:           expanded,
				Folder:         MailFolderInbox,
				CreatedAt:      now,
				UpdatedAt:      now,
			}
			batch.Queue(
				`INSERT INTO hst.mails (tracking_id, sender_login, recipient_login, subject, body,
				        folder, created_at, updated_at)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $7) RETURNING mail_id`,
				v.TrackingId, v.SenderLogin, v.RecipientLogin, v.Subject, v.Body, v.Folder, now)
			delivered = append(delivered, v)
			refs = append(refs, len(delivered)-1)
			sent++
		}

		if in.Email {
			if r.Email == "" {
				skipped++
				continue
			}
			batch.Queue(
				`INSERT INTO hst.outbox (mail_server_id, recipient, subject, body, state, created_at)
				 VALUES ($1, $2, $3, $4, $5, $6)`,
				mailServerId, r.Email, subject, expanded, model.OutboxState_queued, now)
			refs = append(refs, -1)
			queued++
		}
	}

	br := tx.SendBatch(ctx, batch)
	for _, ref := range refs {
		var err error
		if ref >= 0 {
			err = br.QueryRow().Scan(&delivered[ref].MailId)
		} else {
			_, err = br.Exec()
		}
		if err != nil {
			_ = br.Close()
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
	}
	if err := br.Close(); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// only after the commit: a socket push for a row that rolled back would show ghost mail
	for i := range delivered {
		s.notifyMailInbox(&delivered[i])
	}

	s.JournalEntry(c, model.JournalType_mail, logger.CodeOK,
		journal.MailBroadcastMsg(in.Subject, len(recipients), queued), in)

	return s.App.HttpResponseOK(c, fiber.Map{
		"sent": sent, "emails_queued": queued, "skipped_no_email": skipped,
	})
}

// PreviewMail counts who a To expression reaches, for the badge beside the field.
//
//	@Id			PreviewMail
//	@Tags		Mails
//	@Accept		json
//	@Produce	json
//	@Param		body	body		BodyPreviewMail	true	"the recipients to count"
//	@Success	200		{object}	Response
//	@Failure	400		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/mails/preview [post]
func (s *HttpServer) PreviewMail(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	var in BodyPreviewMail
	if err := c.BodyParser(&in); err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}
	if err := s.Validate.Struct(in); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	where, args, err := mailRecipientsWhere(snap, in.To, in.Logins, in.GroupMask)
	if err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	var count, withEmail int64
	if err := s.DB.DB.QueryRow(c.UserContext(),
		`SELECT COUNT(*), COUNT(*) FILTER (WHERE u.email <> '')
		   FROM hst.users u WHERE `+where, args...).Scan(&count, &withEmail); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, fiber.Map{"count": count, "with_email": withEmail})
}

// DeleteMyMail moves a message to the bin, and purges it when it is already there.
//
//	@Id			DeleteMyMail
//	@Tags		Trader
//	@Produce	json
//	@Param		tracking_id	path		string	true	"the mail"
//	@Success	200			{object}	Response
//	@Failure	404			{object}	Response
//	@Failure	500			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/mails/{tracking_id} [delete]
func (s *HttpServer) DeleteMyMail(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	out, err := s.readMails(c.UserContext(),
		"m.tracking_id = $1 AND (m.sender_login = $2 OR m.recipient_login = $2)",
		[]any{c.Params("tracking_id"), snap.Login}, pageOpts{limit: 1})
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if len(out) == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	v := out[0]
	if v.Folder == MailFolderBin {
		_, err = s.DB.DB.Exec(c.UserContext(), `DELETE FROM hst.mails WHERE mail_id = $1`, v.MailId)
	} else {
		_, err = s.DB.DB.Exec(c.UserContext(),
			`UPDATE hst.mails SET folder = $1, deleted_at = $2, updated_at = $2 WHERE mail_id = $3`,
			MailFolderBin, time.Now().UnixNano(), v.MailId)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, nil)
}
