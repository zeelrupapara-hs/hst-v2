package v1

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"path/filepath"
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

var (
	errNotADraft          = errors.New("only a draft can be edited")
	errEmptyMail          = errors.New("a mail needs a subject or a body")
	errNoMailbox          = errors.New("your manager account has no mailbox name, set one to send internal mail")
	errBadMailbox         = errors.New("that mailbox does not exist")
	errBadReplyTo         = errors.New("the message being replied to was not found")
	errBadAttachments     = errors.New("an attachment is missing or already sent")
	errTooManyAttachments = errors.New("up to 5 files can be attached")
)

// ViewMail is one message as its owner sees it. SenderName is the sender's mailbox name when
// the sender is staff, empty for a trader.
type ViewMail struct {
	MailId         int64            `json:"mail_id"`
	TrackingId     string           `json:"tracking_id"`
	ThreadId       string           `json:"thread_id"`
	AttachId       string           `json:"attach_id,omitempty"`
	SenderLogin    int64            `json:"sender_login"`
	SenderName     string           `json:"sender_name"`
	RecipientLogin int64            `json:"recipient_login"`
	RecipientName  string           `json:"recipient_name"`
	Subject        string           `json:"subject"`
	Body           string           `json:"body"`
	Folder         int32            `json:"folder"`
	ReadAt         int64            `json:"read_at"`
	CreatedAt      int64            `json:"created_at"`
	UpdatedAt      int64            `json:"updated_at"`
	Attachments    []ViewAttachment `json:"attachments,omitempty"`
}

// ViewAttachment names a stored file; the bytes come from the download endpoint.
type ViewAttachment struct {
	AttachmentId int64  `json:"attachment_id"`
	Name         string `json:"name"`
	Size         int64  `json:"size"`
}

// BodyMail is a mail a trader sends or saves: the mailbox picked from the broker's list,
// optionally as a reply into an existing thread, with staged attachments.
type BodyMail struct {
	MailboxLogin  int64   `json:"mailbox_login"`
	ReplyTo       string  `json:"reply_to"`
	Subject       string  `json:"subject"`
	Body          string  `json:"body"`
	Draft         bool    `json:"draft"`
	AttachmentIds []int64 `json:"attachment_ids"`
}

const mailColumns = `m.mail_id, m.tracking_id, m.thread_id::text, COALESCE(m.attach_id::text, ''),
	m.sender_login, COALESCE(NULLIF(mg.mailbox, ''), NULLIF(us.name, ''), ''),
	m.recipient_login, COALESCE(NULLIF(mr.mailbox, ''), NULLIF(ur.name, ''), ''), m.subject,
	m.body, m.folder, m.read_at, m.created_at, m.updated_at`

// both parties resolve to something readable: a staff mailbox name first, the account name after
const mailFrom = ` FROM hst.mails m
	LEFT JOIN hst.managers mg ON mg.login = m.sender_login
	LEFT JOIN hst.users us ON us.login = m.sender_login
	LEFT JOIN hst.managers mr ON mr.login = m.recipient_login
	LEFT JOIN hst.users ur ON ur.login = m.recipient_login
	WHERE `

// readMails is the one query behind every mail read handler.
func (s *HttpServer) readMails(ctx context.Context, where string, args []any, p pageOpts) ([]ViewMail, error) {
	where, args = p.bound("m.created_at", where, args)

	rows, err := s.DB.DB.Query(ctx,
		`SELECT `+mailColumns+mailFrom+where+p.tail("m.mail_id"), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []ViewMail{}
	for rows.Next() {
		var v ViewMail
		if err := rows.Scan(&v.MailId, &v.TrackingId, &v.ThreadId, &v.AttachId, &v.SenderLogin,
			&v.SenderName, &v.RecipientLogin, &v.RecipientName,
			&v.Subject, &v.Body, &v.Folder, &v.ReadAt, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}

	return out, rows.Err()
}

// threadOf resolves a reply into the thread it continues; the caller must own the original.
func (s *HttpServer) threadOf(ctx context.Context, trackingId string, login int64) (string, error) {
	var thread string
	err := s.DB.DB.QueryRow(ctx,
		`SELECT thread_id::text FROM hst.mails
		  WHERE tracking_id = $1 AND (sender_login = $2 OR recipient_login = $2)`,
		trackingId, login).Scan(&thread)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", errBadReplyTo
	}

	return thread, err
}

// claimAttachments stamps the caller's staged uploads with the send's attach id, inside its tx.
func claimAttachments(ctx context.Context, tx pgx.Tx, ids []int64, login int64, attachId string) error {
	if len(ids) > 5 {
		return errTooManyAttachments
	}

	tag, err := tx.Exec(ctx,
		`UPDATE hst.mail_attachments SET attach_id = $1
		  WHERE attachment_id = ANY($2) AND owner_login = $3 AND attach_id IS NULL`,
		attachId, ids, login)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != int64(len(ids)) {
		return errBadAttachments
	}

	return nil
}

// loadAttachments lists the files hanging off one logical message.
func (s *HttpServer) loadAttachments(ctx context.Context, attachId string) ([]ViewAttachment, error) {
	rows, err := s.DB.DB.Query(ctx,
		`SELECT attachment_id, name, size FROM hst.mail_attachments
		  WHERE attach_id = $1 ORDER BY attachment_id`, attachId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []ViewAttachment{}
	for rows.Next() {
		var v ViewAttachment
		if err := rows.Scan(&v.AttachmentId, &v.Name, &v.Size); err != nil {
			return nil, err
		}
		out = append(out, v)
	}

	return out, rows.Err()
}

// ViewMailbox is one entry of the broker's mailbox list a trader writes to.
type ViewMailbox struct {
	Login   int64  `json:"login"`
	Mailbox string `json:"mailbox"`
}

// GetMailboxes lists the mailboxes a trader may write to: the staff whose group masks reach
// the caller's own group, so mail goes to whoever is actually responsible for the account.
//
//	@Id			GetMailboxes
//	@Tags		Trader
//	@Produce	json
//	@Success	200	{object}	Response{data=[]ViewMailbox}
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/mailboxes [get]
func (s *HttpServer) GetMailboxes(c *fiber.Ctx) error {
	ctx := c.UserContext()

	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	rows, err := s.DB.DB.Query(ctx,
		`SELECT login, mailbox, groups FROM hst.managers WHERE mailbox <> '' ORDER BY mailbox, login`)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	type candidate struct {
		ViewMailbox
		groups []string
	}
	all := []candidate{}
	for rows.Next() {
		var v candidate
		if err := rows.Scan(&v.Login, &v.Mailbox, &v.groups); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		all = append(all, v)
	}
	if rows.Err() != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, rows.Err())
	}

	out := []ViewMailbox{}
	for _, m := range all {
		ok, err := s.masksReach(ctx, m.groups, snap.Group)
		if err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		if ok {
			out = append(out, m.ViewMailbox)
		}
	}

	return s.App.HttpResponseOK(c, out)
}

// masksReach says whether a manager's group masks cover one group, judged by the same
// GroupAccess predicate every list endpoint filters with.
func (s *HttpServer) masksReach(ctx context.Context, masks []string, group string) (bool, error) {
	where, args := utils.GroupAccess(masks, "$1::text", 2)
	switch where {
	case "TRUE":
		return true, nil
	case "FALSE":
		return false, nil
	}

	var ok bool
	err := s.DB.DB.QueryRow(ctx, `SELECT `+where, append([]any{group}, args...)...).Scan(&ok)
	return ok, err
}

// mailboxExists confirms the picked recipient is a broker mailbox whose masks reach the
// sender's group — the same set the dropdown offered.
func (s *HttpServer) mailboxExists(ctx context.Context, login int64, group string) (bool, error) {
	var masks []string
	err := s.DB.DB.QueryRow(ctx,
		`SELECT groups FROM hst.managers WHERE login = $1 AND mailbox <> ''`, login).Scan(&masks)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return s.masksReach(ctx, masks, group)
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
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	var where string
	args := []any{snap.Login}

	// expanding a conversation asks for one thread across every folder the caller owns
	if thread := c.Query("thread_id"); thread != "" {
		if uuid.Validate(thread) != nil {
			return s.App.HttpResponseBadQueryParams(c, errors.New("thread_id must be a uuid"))
		}
		where = "(m.sender_login = $1 OR m.recipient_login = $1) AND m.thread_id = $2"
		args = append(args, thread)
	} else {
		var err error
		if where, err = mailFolderWhere(c.Query("type")); err != nil {
			return s.App.HttpResponseBadQueryParams(c, err)
		}
	}

	out, err := s.readMails(c.UserContext(), where, args, readPage(c, 100))
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
	if v.AttachId != "" {
		if v.Attachments, err = s.loadAttachments(c.UserContext(), v.AttachId); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
	}
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

	ctx := c.UserContext()

	// To is a broker mailbox picked from the dropdown; a draft may leave it empty for now
	if !in.Draft || in.MailboxLogin != 0 {
		ok, err := s.mailboxExists(ctx, in.MailboxLogin, snap.Group)
		if err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		if !ok {
			return s.App.HttpResponseBadRequest(c, errBadMailbox)
		}
	}

	now := time.Now().UnixNano()
	folder := MailFolderOutbox
	if in.Draft {
		folder = MailFolderDraft
	}

	v := ViewMail{
		TrackingId:     uuid.NewString(),
		SenderLogin:    snap.Login,
		RecipientLogin: in.MailboxLogin,
		Subject:        in.Subject,
		Body:           mailer.SanitizeHTML(in.Body),
		Folder:         int32(folder),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// a reply continues its thread; a fresh mail starts one under its own tracking id
	v.ThreadId = v.TrackingId
	if in.ReplyTo != "" {
		thread, err := s.threadOf(ctx, in.ReplyTo, snap.Login)
		if err != nil {
			if errors.Is(err, errBadReplyTo) {
				return s.App.HttpResponseBadRequest(c, err)
			}
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		v.ThreadId = thread
	}

	if len(in.AttachmentIds) > 0 {
		v.AttachId = uuid.NewString()
	}

	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if v.AttachId != "" {
		if err := claimAttachments(ctx, tx, in.AttachmentIds, snap.Login, v.AttachId); err != nil {
			if errors.Is(err, errBadAttachments) || errors.Is(err, errTooManyAttachments) {
				return s.App.HttpResponseBadRequest(c, err)
			}
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
	}

	if err := tx.QueryRow(ctx,
		`INSERT INTO hst.mails (tracking_id, thread_id, attach_id, sender_login, recipient_login,
		        subject, body, folder, created_at, updated_at)
		 VALUES ($1, $2, NULLIF($3, '')::uuid, $4, $5, $6, $7, $8, $9, $9) RETURNING mail_id`,
		v.TrackingId, v.ThreadId, v.AttachId, v.SenderLogin, v.RecipientLogin,
		v.Subject, v.Body, v.Folder, now).
		Scan(&v.MailId); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// a sent mail is two rows, so each side bins its own copy without touching the other's
	inbox := v
	if !in.Draft {
		inbox.TrackingId = uuid.NewString()
		inbox.Folder = MailFolderInbox
		if err := tx.QueryRow(ctx,
			`INSERT INTO hst.mails (tracking_id, thread_id, attach_id, sender_login, recipient_login,
			        subject, body, folder, created_at, updated_at)
			 VALUES ($1, $2, NULLIF($3, '')::uuid, $4, $5, $6, $7, $8, $9, $9) RETURNING mail_id`,
			inbox.TrackingId, inbox.ThreadId, inbox.AttachId, inbox.SenderLogin, inbox.RecipientLogin,
			inbox.Subject, inbox.Body, inbox.Folder, now).Scan(&inbox.MailId); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if in.Draft {
		s.JournalEntry(c, model.JournalType_mail, logger.CodeOK, journal.MailDraftMsg(in.Subject), nil)
	} else {
		s.notifyMailInbox(&inbox)
		s.JournalEntry(c, model.JournalType_mail, logger.CodeOK, journal.MailSentMsg(in.Subject), nil)
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
	Logins        []int64 `json:"logins"`
	GroupMask     string  `json:"group_mask" validate:"max=128"`
	To            string  `json:"to" validate:"max=4096"`
	Subject       string  `json:"subject" validate:"required,max=128"`
	Body          string  `json:"body" validate:"required,max=65536"`
	Internal      *bool   `json:"internal"`
	Email         bool    `json:"email"`
	MailServerId  int64   `json:"mail_server_id"`
	ReplyTo       string  `json:"reply_to"`
	AttachmentIds []int64 `json:"attachment_ids"`
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

	// MT5 parity: internal mail carries the sender's mailbox name, so no name means no send
	var mailbox string
	if internal {
		if err := s.DB.DB.QueryRow(ctx,
			`SELECT mailbox FROM hst.managers WHERE login = $1`, snap.Login).Scan(&mailbox); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		if mailbox == "" {
			return s.App.HttpResponseBadRequest(c, errNoMailbox)
		}
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

	// a reply keeps its conversation; otherwise every copy of this send starts a fresh thread
	var replyThread string
	if in.ReplyTo != "" {
		replyThread, err = s.threadOf(ctx, in.ReplyTo, snap.Login)
		if err != nil {
			if errors.Is(err, errBadReplyTo) {
				return s.App.HttpResponseBadRequest(c, err)
			}
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
	}

	var attachId string
	if internal && len(in.AttachmentIds) > 0 {
		attachId = uuid.NewString()
	}

	// sanitized once here, then per-recipient macro values are escaped on the way in
	body := mailer.SanitizeHTML(in.Body)

	now := time.Now().UnixNano()
	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if attachId != "" {
		if err := claimAttachments(ctx, tx, in.AttachmentIds, snap.Login, attachId); err != nil {
			if errors.Is(err, errBadAttachments) || errors.Is(err, errTooManyAttachments) {
				return s.App.HttpResponseBadRequest(c, err)
			}
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
	}

	var (
		sent, queued, skipped int
		delivered             []ViewMail
		// refs mirrors the batch: an index into delivered for a mails insert, -1 for outbox
		refs []int
	)

	const insertMail = `INSERT INTO hst.mails (tracking_id, thread_id, attach_id, sender_login,
	        recipient_login, subject, body, folder, created_at, updated_at)
	 VALUES ($1, $2, NULLIF($3, '')::uuid, $4, $5, $6, $7, $8, $9, $9) RETURNING mail_id`

	batch := &pgx.Batch{}
	for i := range recipients {
		r := &recipients[i]
		subject := expandMailMacros(in.Subject, r, false)
		expanded := expandMailMacros(body, r, true)

		if internal {
			v := ViewMail{
				TrackingId:     uuid.NewString(),
				AttachId:       attachId,
				SenderLogin:    snap.Login,
				SenderName:     mailbox,
				RecipientLogin: r.Login,
				RecipientName:  r.Name,
				Subject:        subject,
				Body:           expanded,
				Folder:         MailFolderInbox,
				CreatedAt:      now,
				UpdatedAt:      now,
			}
			v.ThreadId = v.TrackingId
			if replyThread != "" {
				v.ThreadId = replyThread
			}
			batch.Queue(insertMail,
				v.TrackingId, v.ThreadId, v.AttachId, v.SenderLogin, v.RecipientLogin,
				v.Subject, v.Body, v.Folder, now)
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

	// the sender keeps ONE outbox copy per send, template body un-expanded — MT5 behavior
	var outboxRow *ViewMail
	if internal {
		v := ViewMail{
			TrackingId:  uuid.NewString(),
			AttachId:    attachId,
			SenderLogin: snap.Login,
			SenderName:  mailbox,
			Subject:     in.Subject,
			Body:        body,
			Folder:      MailFolderOutbox,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if len(recipients) == 1 {
			v.RecipientName = recipients[0].Name
		}
		v.ThreadId = v.TrackingId
		if replyThread != "" {
			v.ThreadId = replyThread
		}
		// a one-to-one send shares its recipient's thread, so the reply chains on both sides
		if len(recipients) == 1 {
			v.RecipientLogin = recipients[0].Login
			if internal && replyThread == "" && len(delivered) > 0 {
				v.ThreadId = delivered[0].ThreadId
			}
		}
		batch.Queue(insertMail,
			v.TrackingId, v.ThreadId, v.AttachId, v.SenderLogin, v.RecipientLogin,
			v.Subject, v.Body, v.Folder, now)
		delivered = append(delivered, v)
		refs = append(refs, len(delivered)-1)
		outboxRow = &delivered[len(delivered)-1]
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
		if outboxRow == &delivered[i] {
			continue
		}
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

// MT5 trader-side attachment limits.
const (
	mailAttachMaxFiles = 5
	mailAttachMaxFile  = 8 * 1024 * 1024
	mailAttachMaxTotal = 16 * 1024 * 1024
)

var errBadAttachmentType = errors.New("that file type cannot be attached")
var errAttachmentTooBig = errors.New("an attachment is limited to 8MB, 16MB in total")

// mailAttachExts is the MT5 whitelist of attachable file types.
var mailAttachExts = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".bmp": true, ".gif": true,
	".zip": true, ".7z": true, ".doc": true, ".xls": true, ".docx": true,
	".xlsx": true, ".odt": true, ".rtf": true, ".csv": true, ".txt": true, ".log": true,
}

// UploadMailAttachments stages files for a send: rows sit unclaimed under the uploader's
// login until a send stamps them with its attach id.
//
//	@Id			UploadMailAttachments
//	@Tags		Trader
//	@Accept		mpfd
//	@Produce	json
//	@Success	200	{object}	Response{data=[]ViewAttachment}
//	@Failure	400	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/mails/attachments [post]
func (s *HttpServer) UploadMailAttachments(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	form, err := c.MultipartForm()
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}

	files := form.File["files"]
	if len(files) == 0 {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}
	if len(files) > mailAttachMaxFiles {
		return s.App.HttpResponseBadRequest(c, errTooManyAttachments)
	}

	var total int64
	for _, f := range files {
		if !mailAttachExts[strings.ToLower(filepath.Ext(f.Filename))] {
			return s.App.HttpResponseBadRequest(c, errBadAttachmentType)
		}
		if f.Size > mailAttachMaxFile {
			return s.App.HttpResponseBadRequest(c, errAttachmentTooBig)
		}
		total += f.Size
	}
	if total > mailAttachMaxTotal {
		return s.App.HttpResponseBadRequest(c, errAttachmentTooBig)
	}

	now := time.Now().UnixNano()
	out := []ViewAttachment{}
	for _, f := range files {
		src, err := f.Open()
		if err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		data, err := io.ReadAll(src)
		_ = src.Close()
		if err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}

		v := ViewAttachment{Name: filepath.Base(f.Filename), Size: int64(len(data))}
		if err := s.DB.DB.QueryRow(c.UserContext(),
			`INSERT INTO hst.mail_attachments (owner_login, name, size, data, created_at)
			 VALUES ($1, $2, $3, $4, $5) RETURNING attachment_id`,
			snap.Login, v.Name, v.Size, data, now).Scan(&v.AttachmentId); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		out = append(out, v)
	}

	return s.App.HttpResponseOK(c, out)
}

// DownloadMailAttachment streams one file to whoever legitimately holds it: a party to a
// mail carrying its attach id, or its uploader while it is still staged.
//
//	@Id			DownloadMailAttachment
//	@Tags		Trader
//	@Produce	octet-stream
//	@Param		attachment_id	path	int	true	"the file"
//	@Success	200
//	@Failure	404	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/mails/attachments/{attachment_id} [get]
func (s *HttpServer) DownloadMailAttachment(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	id, err := strconv.ParseInt(c.Params("attachment_id"), 10, 64)
	if err != nil {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	var name string
	var data []byte
	err = s.DB.DB.QueryRow(c.UserContext(),
		`SELECT a.name, a.data FROM hst.mail_attachments a
		  WHERE a.attachment_id = $1
		    AND ((a.attach_id IS NULL AND a.owner_login = $2)
		      OR EXISTS (SELECT 1 FROM hst.mails m WHERE m.attach_id = a.attach_id
		                    AND (m.sender_login = $2 OR m.recipient_login = $2)))`,
		id, snap.Login).Scan(&name, &data)
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	c.Set(fiber.HeaderContentDisposition, `attachment; filename="`+strings.ReplaceAll(name, `"`, "")+`"`)
	c.Set(fiber.HeaderContentType, fiber.MIMEOctetStream)
	return c.Send(data)
}
