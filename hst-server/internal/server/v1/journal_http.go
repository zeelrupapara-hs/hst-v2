package v1

import (
	errs "hstserver/pkg/errors"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
)

// ViewJournal is one line of an account's own activity log.
type ViewJournal struct {
	JournalId int64  `json:"journal_id"`
	CreatedAt int64  `json:"created_at"`
	Channel   string `json:"channel"`
	Ip        string `json:"ip"`
	Message   string `json:"message"`
}

// GetMyJournal returns the calling account's own journal entries, newest first.
//
//	@Id			GetMyJournal
//	@Tags		Journal
//	@Produce	json
//	@Param		limit	query		int	false	"how many, newest first"
//	@Param		page	query		int	false	"which page, zero based"
//	@Param		from	query		int	false	"unix seconds, inclusive"
//	@Param		to		query		int	false	"unix seconds, exclusive"
//	@Success	200		{object}	Response{data=[]ViewJournal}
//	@Failure	403		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/journal [get]
func (s *HttpServer) GetMyJournal(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	// the login comes from the session, so no account can read another's trail
	p := readPage(c, 100)
	where, args := p.bound("created_at", "login = $1", []any{snap.Login})

	// host() drops the /32 that inet carries; the channel is whatever the writer put in the detail
	rows, err := s.DB.DB.Query(c.UserContext(),
		`SELECT journal_id, created_at, COALESCE(detail->>'channel', 'api'),
		        COALESCE(host(ip), ''), message
		   FROM hst.journal
		  WHERE `+where+p.tail("journal_id"), args...)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	out := []ViewJournal{}
	for rows.Next() {
		var v ViewJournal
		if err := rows.Scan(&v.JournalId, &v.CreatedAt, &v.Channel, &v.Ip, &v.Message); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		out = append(out, v)
	}
	if rows.Err() != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, rows.Err())
	}

	return s.App.HttpResponseOK(c, out)
}
