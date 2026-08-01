package v1

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

// RegisterV1 registers the surface both panels share, and hands back the groups they mount
// under: root for the public routes, api for everything behind a token.
//
// A manager and a trading account authenticate the same way and listen on the same socket;
// what differs is what each may reach afterwards, which is why the panels register separately.
func (s *HttpServer) RegisterV1() (root, api fiber.Router) {

	// Root group with the requests logger and the header reader
	root = s.App.Group("/", s.Middleware.SecurityHeaders, s.Middleware.CORS(s.Cfg.HTTP.CorsOrigins),
		s.Middleware.RequestsLogger, s.Middleware.HeaderReader)

	if s.Cfg.HTTP.SwaggerEnabled {
		s.registerSwagger(root)
	}

	// websocket.
	root.Get("/ws", s.Middleware.UpgradeWS, s.Middleware.Protect, websocket.New(s.ServeWS, websocket.Config{
		// the handshake must echo the subprotocol the client offered, or the browser fails the connection
		Subprotocols: []string{"bearer"},
	}))

	api = root.Group("/api")
	v1 := api.Group("/v1")

	// system
	system := v1.Group("/system")
	system.Get("/monitor/health", s.CheckSystemHealth)
	system.Get("/monitor/live", s.CheckSystemLive)
	system.Get("/monitor/cache", s.Middleware.Protect, s.Middleware.RequireManager, s.CacheStats)
	system.Get("/monitor/ws", s.Middleware.Protect, s.Middleware.RequireManager, s.WSStats)

	return root, api
}

// registerSwagger serves one spec per panel, so a manager reading the docs is not shown the
// trader API and the other way round.
func (s *HttpServer) registerSwagger(root fiber.Router) {
	for _, spec := range []struct{ path, instance string }{
		{"/swagger/admin/*", "admin"},
		{"/swagger/trader/*", "trader"},
	} {
		root.Get(spec.path, swagger.New(swagger.Config{
			InstanceName: spec.instance,
			// swagger 2.0 has no bearer scheme, so the token is sent verbatim.
			// Add the prefix here when the pasted value is missing it.
			RequestInterceptor: `(req) => {
				const a = req.headers.Authorization;
				if (a && !/^(Bearer|Basic) /i.test(a)) {
					req.headers.Authorization = "Bearer " + a.trim();
				}
				return req;
			}`,
			PersistAuthorization: true,
		}))
	}
}

// RegisterInternal registers service-to-service routes for ingestion workers.
func (s *HttpServer) RegisterInternal() {
	internal := s.App.Group("/internal/v1", s.Middleware.ServiceAuth)
	internal.Get("/datafeeds", s.ListInternalDatafeeds)
	internal.Get("/datafeeds/:id", s.GetInternalDatafeed)
}
