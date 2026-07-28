package v1

import (
	"hstserver/pkg/http"

	"github.com/gofiber/fiber/v2"
)

// HealthResponse is the payload returned by the health endpoint
type HealthResponse struct {
	Status  string            `json:"status"`
	Version string            `json:"version"`
	Checks  map[string]string `json:"checks"`
}

// CheckSystemHealth godoc
//
//	@Summary	Check the system health
//	@Tags		System
//	@Produce	json
//	@Success	200	{object}	HealthResponse
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

	if status != "ok" {
		return c.Status(http.StatusInternalServerError).JSON(res)
	}

	return c.Status(http.StatusOK).JSON(res)
}

// CheckSystemLive godoc
//
//	@Summary	Liveness probe, does not touch the database
//	@Tags		System
//	@Produce	json
//	@Success	200	{object}	map[string]string
//	@Router		/api/v1/system/monitor/live [get]
func (s *HttpServer) CheckSystemLive(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(fiber.Map{"status": "alive"})
}
