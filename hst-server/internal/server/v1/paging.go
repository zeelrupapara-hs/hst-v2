package v1

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// How a history read is bounded.
//
// MT5 keeps no closed positions, so a terminal's History tab is built from orders and deals.
// Those two tables only grow, and an account traded for a year is not a single response, so
// every history read takes a date range and a page.

// pageOpts is one caller's slice of a history table. A zero limit means the whole set.
type pageOpts struct {
	limit, offset int
	// from and to are unix seconds, half open, and zero means unbounded on that side
	from, to int64
}

// readPage reads the paging query params, falling back to def when the caller names no limit.
func readPage(c *fiber.Ctx, def int) pageOpts {
	limit := c.QueryInt("limit", def)
	if limit < 0 || limit > 500 {
		limit = def
	}

	page := c.QueryInt("page", 0)
	if page < 0 {
		page = 0
	}

	return pageOpts{
		limit:  limit,
		offset: page * limit,
		from:   int64(c.QueryInt("from", 0)),
		to:     int64(c.QueryInt("to", 0)),
	}
}

// bound narrows a query to the requested range, given in seconds against a nanosecond column.
func (p pageOpts) bound(timeCol, where string, args []any) (string, []any) {
	if p.from > 0 {
		args = append(args, p.from*1e9)
		where += " AND " + timeCol + " >= $" + strconv.Itoa(len(args))
	}

	if p.to > 0 {
		args = append(args, p.to*1e9)
		where += " AND " + timeCol + " < $" + strconv.Itoa(len(args))
	}

	return where, args
}

// tail is the ordering and paging that closes a history query, newest first.
func (p pageOpts) tail(orderCol string) string {
	out := " ORDER BY " + orderCol + " DESC"
	if p.limit > 0 {
		out += " LIMIT " + strconv.Itoa(p.limit) + " OFFSET " + strconv.Itoa(p.offset)
	}

	return out
}
