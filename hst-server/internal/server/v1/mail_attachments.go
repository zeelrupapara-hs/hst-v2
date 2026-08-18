package v1

import (
	"errors"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	errs "hstserver/pkg/errors"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

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
