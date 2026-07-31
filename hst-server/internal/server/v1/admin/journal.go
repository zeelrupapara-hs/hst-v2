package admin

import (
	"fmt"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
)

// journalSortBy are the real columns of journal.
var journalSortBy = utils.NewSortable("journal_id", "created_at", "type", "code")

// host() drops the /32 that inet carries into text, and pgx has no string plan
// for inet either way
const journalColumns = `journal_id, created_at, type, code, login,
	coalesce(host(ip), ''), message, coalesce(detail, '{}'::jsonb)`

// MyJournal returns a page of the caller's own journal entries, newest first.
// The login comes from the session, never from the request, so one manager
// cannot read another's trail.
//
//	@Id			MyJournal
//	@Tags		Journal
//	@Produce	json
//	@Param		page	query		int		false	"page number, from 1"
//	@Param		limit	query		int		false	"rows per page, max 500"
//	@Param		from	query		int		false	"created_at lower bound, unix nanoseconds"
//	@Param		to		query		int		false	"created_at upper bound, unix nanoseconds"
//	@Param		type	query		int		false	"event type, 0 for all"		Enums(0, 1, 2, 3, 4, 5, 6, 7, 8)
//	@Param		mode	query		string	false	"which entries to return"	Enums(full, without_logins, errors_only)
//	@Param		search	query		string	false	"matches message, case sensitive"
//	@Param		sort_by	query		string	false	"journal_id, created_at, type, code"	Enums(journal_id, created_at, type, code)
//	@Param		order	query		string	false	"asc or desc"							Enums(asc, desc)
//	@Success	200		{object}	Response{data=[]model.Journal}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/journal [get]
func (s *Server) MyJournal(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseUnauthorized(c, errs.ErrInvalidSession)
	}

	q, err := utils.QueryFilter(c, journalSortBy, "created_at")
	if err != nil {
		return s.App.HttpResponseBadQueryParams(c, err)
	}

	typ := c.QueryInt("type", 0)
	if typ < 0 || typ > 8 {
		return s.App.HttpResponseBadQueryParams(c, fmt.Errorf("type %d is not an event type", typ))
	}

	mode, ok := model.JournalMode_value[c.Query("mode", "full")]
	if !ok {
		return s.App.HttpResponseBadQueryParams(c,
			fmt.Errorf("mode %q is not full, without_logins or errors_only", c.Query("mode")))
	}

	// sort_by is validated against an allowlist in QueryFilter; a bind
	// parameter cannot carry an ORDER BY clause
	rows, err := s.DB.DB.Query(c.UserContext(),
		`SELECT `+journalColumns+`
		   FROM hst.journal
		  WHERE ($1 = 0 OR created_at >= $1)
		    AND ($2 = 0 OR created_at <= $2)
		    AND ($3 = 0 OR type = $3)
		    AND login = $4
		    AND ($5 = '' OR message LIKE '%'||$5||'%')
		    AND ($6 = 0 OR ($6 = 1 AND code <> 4) OR ($6 = 2 AND code IN (2, 3)))
		  ORDER BY `+q.SortBy+`
		  LIMIT $7 OFFSET $8`,
		c.QueryInt("from", 0), c.QueryInt("to", 0), typ, snap.Login,
		q.Search, mode, q.Limit, q.Offset)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	out := []model.Journal{}
	for rows.Next() {
		var v model.Journal
		if err := rows.Scan(&v.JournalId, &v.CreatedAt, &v.Type, &v.Code, &v.Login,
			&v.Ip, &v.Message, &v.Detail); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		out = append(out, v)
	}
	if rows.Err() != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, rows.Err())
	}

	return s.App.HttpResponseOK(c, out)
}
