package v1

import (
	"hstserver/pkg/http"

	"github.com/gofiber/fiber/v2"
)

// LiveResponse is the payload returned by the liveness probe
type LiveResponse struct {
	Status string `json:"status"`
}

// HealthResponse is the payload returned by the health endpoint
type HealthResponse struct {
	Status  string            `json:"status"`
	Version string            `json:"version"`
	Checks  map[string]string `json:"checks"`
}

// CheckSystemHealth godoc
//
//	@Id			CheckSystemHealth
//	@Summary	Check the system health
//	@Tags		System
//	@Produce	json
//	@Success	200	{object}	Response{data=HealthResponse}
//	@Failure	500	{object}	Response
//	@Router		/api/v1/system/monitor/health [get]
func (s *HttpServer) CheckSystemHealth(c *fiber.Ctx) error {
	checks := make(map[string]string)
	status := "ok"

	if err := s.DB.DB.Ping(c.Context()); err != nil {
		s.Log.Logger.Errorw("postgres health check failed", "error", err)
		checks["postgres"] = "down"
		status = "degraded"
	} else {
		checks["postgres"] = "ok"
	}

	if err := s.Redis.Health(c.Context()); err != nil {
		s.Log.Logger.Errorw("redis health check failed", "error", err)
		checks["redis"] = "down"
		status = "degraded"
	} else {
		checks["redis"] = "ok"
	}

	res := HealthResponse{
		Status:  status,
		Version: s.Cfg.Setting.Version,
		Checks:  checks,
	}

	// the status code is what a kubernetes probe reads, so it stays; the body
	// carries the same envelope as every other endpoint
	if status != "ok" {
		return c.Status(http.StatusInternalServerError).JSON(&http.HttpResponse{
			Success: false,
			Code:    http.RetOK,
			Data:    res,
			Error:   http.ErrInternalServerError,
		})
	}

	return s.App.HttpResponseOK(c, res)
}

// CheckSystemLive godoc
//
//	@Id			CheckSystemLive
//	@Summary	Liveness probe, does not touch the database
//	@Tags		System
//	@Produce	json
//	@Success	200	{object}	Response{data=LiveResponse}
//	@Router		/api/v1/system/monitor/live [get]
func (s *HttpServer) CheckSystemLive(c *fiber.Ctx) error {
	return s.App.HttpResponseOK(c, LiveResponse{Status: "alive"})
}

// CacheStats reports the session cache, so MAX_ACCOUNT_PER_SHARD is tuned from
// data rather than guessed.
//
//	@Id			CacheStats
//	@Tags		System
//	@Produce	json
//	@Success	200	{object}	Response{data=CacheStats}
//	@Failure	401	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/system/monitor/cache [get]
func (s *HttpServer) CacheStats(c *fiber.Ctx) error {
	return s.App.HttpResponseOK(c, s.OAuth2.Cache.Stats())
}
