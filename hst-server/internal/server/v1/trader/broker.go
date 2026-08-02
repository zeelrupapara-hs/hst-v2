package trader

import (
	"github.com/gofiber/fiber/v2"
)

// ViewBroker is the branding the login screen draws before any session exists.
type ViewBroker struct {
	Name         string `json:"name"`
	Company      string `json:"company"`
	Website      string `json:"website"`
	SupportEmail string `json:"support_email"`
	Version      string `json:"version"`
}

// brokerKeys is the allowlist; hst.settings also holds things no anonymous caller may read.
var brokerKeys = []string{"broker_name", "broker_company", "broker_website", "broker_support_email"}

// PublicBroker serves broker branding, unauthenticated.
//
//	@Id			PublicBroker
//	@Tags		Public
//	@Produce	json
//	@Success	200	{object}	Response{data=ViewBroker}
//	@Failure	500	{object}	Response
//	@Router		/public/broker [get]
func (s *Server) PublicBroker(c *fiber.Ctx) error {
	rows, err := s.DB.DB.Query(c.UserContext(),
		`SELECT key, value FROM hst.settings WHERE key = ANY($1)`, brokerKeys)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	found := map[string]string{}
	for rows.Next() {
		var k, val string
		if err := rows.Scan(&k, &val); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		found[k] = val
	}
	if rows.Err() != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, rows.Err())
	}

	return s.App.HttpResponseOK(c, ViewBroker{
		Name:         found["broker_name"],
		Company:      found["broker_company"],
		Website:      found["broker_website"],
		SupportEmail: found["broker_support_email"],
		Version:      s.Cfg.Setting.Version,
	})
}
