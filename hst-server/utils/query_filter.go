package utils

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// maxLimit caps one page, so a client cannot ask for the whole table.
const maxLimit = 500

// Query is a parsed list request.
type Query struct {
	Page   int
	Limit  int
	Offset int
	Search string
	SortBy string
}

// sortable lists the columns a caller may order by. An allowlist and not string
// interpolation: sort_by goes into the SQL text, where a bind parameter cannot.
var sortable = map[string]struct{}{
	"created_at": {}, "updated_at": {}, "login": {}, "client_id": {},
	"name": {}, "email": {}, "last_access": {}, "registration": {},
}

// QueryFilter parses page, limit, search and sort_by.
func QueryFilter(c *fiber.Ctx) (*Query, error) {
	page := c.QueryInt("page", 1)
	if page < 1 {
		page = 1
	}

	limit := c.QueryInt("limit", 100)
	if limit < 1 {
		limit = 1
	}
	if limit > maxLimit {
		return nil, fmt.Errorf("you can't request more than %d records per page", maxLimit)
	}

	sortBy := c.Query("sort_by", "created_at")
	desc := c.Query("order", "desc") == "desc"
	if _, ok := sortable[sortBy]; !ok {
		return nil, fmt.Errorf("sort_by %q is not a sortable column", sortBy)
	}
	if desc {
		sortBy += " DESC"
	} else {
		sortBy += " ASC"
	}

	return &Query{
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
		Search: c.Query("search"),
		SortBy: sortBy,
	}, nil
}
