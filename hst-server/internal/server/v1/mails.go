package v1

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"hstserver/model"
	"hstserver/pkg/cache"
	errs "hstserver/pkg/errors"
	"hstserver/pkg/journal"
	"hstserver/pkg/logger"
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
	if !in.Draft {
		if _, err := tx.Exec(c.UserContext(),
			`INSERT INTO hst.mails (tracking_id, sender_login, recipient_login, subject, body,
			        folder, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $7)`,
			uuid.NewString(), v.SenderLogin, v.RecipientLogin, v.Subject, v.Body,
			MailFolderInbox, now); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
	}

	if err := tx.Commit(c.UserContext()); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
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

// BodySendMail is a manager's message to a set of accounts, named by login or by group mask.
type BodySendMail struct {
	Logins    []int64 `json:"logins"`
	GroupMask string  `json:"group_mask" validate:"max=128"`
	Subject   string  `json:"subject" validate:"required,max=128"`
	Body      string  `json:"body" validate:"required,max=4000"`
}

var errNoRecipients = errors.New("no accounts match the given logins or group mask")

// sendMailRecipients resolves the request to the logins the caller's group masks cover.
func (s *HttpServer) sendMailRecipients(ctx context.Context, snap *cache.Session, in *BodySendMail) ([]int64, error) {
	where, args := groupWhere(snap.IsManager, snap.ManagerGroups, 1)
	if len(in.Logins) > 0 {
		where += ` AND u.login = ANY($` + strconv.Itoa(len(args)+1) + `)`
		args = append(args, in.Logins)
	} else {
		where += ` AND u."group" LIKE $` + strconv.Itoa(len(args)+1)
		args = append(args, strings.ReplaceAll(in.GroupMask, "*", "%"))
	}

	rows, err := s.DB.DB.Query(ctx, `SELECT u.login FROM hst.users u WHERE `+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []int64{}
	for rows.Next() {
		var login int64
		if err := rows.Scan(&login); err != nil {
			return nil, err
		}
		out = append(out, login)
	}

	return out, rows.Err()
}

// SendMail writes the message into each recipient's inbox; internal mailbox only, no SMTP.
//
//	@Id			SendMail
//	@Tags		Mails
//	@Accept		json
//	@Produce	json
//	@Param		body	body		BodySendMail	true	"the message and who gets it"
//	@Success	200		{object}	Response
//	@Failure	400		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/mails [post]
func (s *HttpServer) SendMail(c *fiber.Ctx) error {
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
	if len(in.Logins) == 0 && in.GroupMask == "" {
		return s.App.HttpResponseBadRequest(c, errors.New("logins or group_mask is required"))
	}

	recipients, err := s.sendMailRecipients(c.UserContext(), snap, &in)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if len(recipients) == 0 {
		return s.App.HttpResponseBadRequest(c, errNoRecipients)
	}

	now := time.Now().UnixNano()
	tx, err := s.DB.DB.Begin(c.UserContext())
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer func() { _ = tx.Rollback(c.UserContext()) }()

	for _, login := range recipients {
		if _, err := tx.Exec(c.UserContext(),
			`INSERT INTO hst.mails (tracking_id, sender_login, recipient_login, subject, body,
			        folder, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $7)`,
			uuid.NewString(), snap.Login, login, in.Subject, in.Body,
			MailFolderInbox, now); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
	}

	if err := tx.Commit(c.UserContext()); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	s.JournalEntry(c, model.JournalType_mail, logger.CodeOK,
		journal.MailBroadcastMsg(in.Subject, len(recipients)), in)

	return s.App.HttpResponseOK(c, fiber.Map{"sent": len(recipients)})
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
