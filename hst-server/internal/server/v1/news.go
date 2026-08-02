package v1

import (
	"context"
	"errors"

	errs "hstserver/pkg/errors"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

// ViewNews is one ingested headline as a terminal shows it.
type ViewNews struct {
	NewsId      int64    `json:"news_id"`
	ExternalId  string   `json:"external_id"`
	Title       string   `json:"title"`
	Summary     string   `json:"summary"`
	Url         string   `json:"url"`
	Source      string   `json:"source"`
	Symbols     []string `json:"symbols"`
	PublishedAt int64    `json:"published_at"`
	CreatedAt   int64    `json:"created_at"`
}

const newsColumns = `n.news_id, n.external_id, n.title, n.summary, n.url, n.source,
	n.symbols, n.published_at, n.created_at`

// readNews is the one query behind every news read handler.
func (s *HttpServer) readNews(ctx context.Context, where string, args []any, p pageOpts) ([]ViewNews, error) {
	where, args = p.bound("n.published_at", where, args)

	rows, err := s.DB.DB.Query(ctx,
		`SELECT `+newsColumns+` FROM hst.news n WHERE `+where+p.tail("n.published_at"), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []ViewNews{}
	for rows.Next() {
		var v ViewNews
		if err := rows.Scan(&v.NewsId, &v.ExternalId, &v.Title, &v.Summary, &v.Url,
			&v.Source, &v.Symbols, &v.PublishedAt, &v.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}

	return out, rows.Err()
}

// GetMyNews lists ingested news, newest first.
//
//	@Id			GetMyNews
//	@Tags		Trader
//	@Produce	json
//	@Param		symbol	query		string	false	"only items tagged with this symbol"
//	@Param		limit	query		int		false	"how many, newest first"
//	@Param		page	query		int		false	"which page, zero based"
//	@Param		from	query		int		false	"unix seconds, inclusive"
//	@Param		to		query		int		false	"unix seconds, exclusive"
//	@Success	200		{object}	Response{data=[]ViewNews}
//	@Failure	403		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/news [get]
func (s *HttpServer) GetMyNews(c *fiber.Ctx) error {
	if _, ok := utils.GetClient(c); !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	where, args := "TRUE", []any{}
	if symbol := c.Query("symbol"); symbol != "" {
		where, args = "$1 = ANY(n.symbols)", []any{symbol}
	}

	out, err := s.readNews(c.UserContext(), where, args, readPage(c, 100))
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, out)
}

// GetMyNewsItem returns one headline.
//
//	@Id			GetMyNewsItem
//	@Tags		Trader
//	@Produce	json
//	@Param		news_id	path		int	true	"the item"
//	@Success	200		{object}	Response{data=ViewNews}
//	@Failure	400		{object}	Response
//	@Failure	404		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/news/{news_id} [get]
func (s *HttpServer) GetMyNewsItem(c *fiber.Ctx) error {
	if _, ok := utils.GetClient(c); !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	newsId, err := c.ParamsInt("news_id")
	if err != nil || newsId <= 0 {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}

	var v ViewNews
	err = s.DB.DB.QueryRow(c.UserContext(),
		`SELECT `+newsColumns+` FROM hst.news n WHERE n.news_id = $1`, newsId).
		Scan(&v.NewsId, &v.ExternalId, &v.Title, &v.Summary, &v.Url, &v.Source,
			&v.Symbols, &v.PublishedAt, &v.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, v)
}
