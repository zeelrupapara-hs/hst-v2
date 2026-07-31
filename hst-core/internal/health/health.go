// Package health serves the probes an orchestrator needs.
//
// Two endpoints, and the difference between them matters. /healthz says the
// process is alive — fail it and kubernetes restarts the pod. /readyz says the
// process can serve — fail it and kubernetes stops routing traffic but leaves
// the pod alone. Wiring a dependency check into the liveness probe is how a
// brief database blip turns into a restart loop across every replica.
//
// net/http on purpose: a probe listener is not worth a dependency.
package health

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"hstcore/config"
	"hstcore/pkg/logger"
)

// Check reports whether one dependency is usable. Keep it cheap: readiness is
// polled every couple of seconds, forever.
type Check func(ctx context.Context) error

// Server is the probe listener. It runs on its own port so a probe never
// queues behind application traffic, and so the port can stay off the public
// service entirely.
type Server struct {
	log    *logger.Logger
	srv    *http.Server
	checks map[string]Check

	// ready flips once boot has finished; until then /readyz fails and no
	// traffic is routed to a pod that is still loading
	ready atomic.Bool
	// shuttingDown makes /readyz fail on SIGTERM while the process keeps
	// serving in-flight work, so the endpoints controller pulls this pod out
	// before it stops answering
	shuttingDown atomic.Bool
}

// New builds the probe server. checks are consulted by /readyz only.
func New(cfg *config.Config, log *logger.Logger, checks map[string]Check) *Server {
	s := &Server{
		log:    log,
		checks: checks,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.live)
	mux.HandleFunc("/readyz", s.readyz)

	s.srv = &http.Server{
		Addr:    net.JoinHostPort(cfg.Health.Host, cfg.Health.Port),
		Handler: mux,
		// a probe that cannot be answered in this long is a failed probe
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return s
}

// Start listens in the background. It returns once the socket is bound, so a
// port already in use is reported at boot rather than swallowed.
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.srv.Addr)
	if err != nil {
		return err
	}

	go func() {
		if err := s.srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.log.Log(logger.TypeNet, logger.CodeErr, "probe server failed",
				"error", err.Error())
		}
	}()

	s.log.Log(logger.TypeNet, logger.CodeOK, "probe server listening",
		"addr", s.srv.Addr)
	return nil
}

// Ready marks boot complete. Call it after everything has loaded.
func (s *Server) Ready() { s.ready.Store(true) }

// Draining fails readiness while the process finishes what it is doing. Call
// it first on SIGTERM, then sleep long enough for the endpoints controller to
// notice, and only then stop the rest. Without that pause, traffic is still
// being routed to a pod that has already begun shutting down.
func (s *Server) Draining() { s.shuttingDown.Store(true) }

// Stop closes the listener.
func (s *Server) Stop(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

// live answers as long as the process is running. It checks nothing on
// purpose: a dependency outage must not restart every pod at once.
func (s *Server) live(w http.ResponseWriter, _ *http.Request) {
	writePlain(w, http.StatusOK, "ok")
}

// readyz answers only when boot has finished, shutdown has not begun, and
// every dependency answers.
func (s *Server) readyz(w http.ResponseWriter, r *http.Request) {
	if s.shuttingDown.Load() {
		writePlain(w, http.StatusServiceUnavailable, "shutting down")
		return
	}
	if !s.ready.Load() {
		writePlain(w, http.StatusServiceUnavailable, "starting")
		return
	}

	// bounded so a hung dependency fails the probe instead of holding the
	// connection open until the kubelet's own timeout
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	for name, check := range s.checks {
		if err := check(ctx); err != nil {
			s.log.Log(logger.TypeNet, logger.CodeWarn, "readiness check failed",
				"check", name, "error", err.Error())
			writePlain(w, http.StatusServiceUnavailable, name+" unavailable")
			return
		}
	}

	writePlain(w, http.StatusOK, "ok")
}

func writePlain(w http.ResponseWriter, code int, body string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(code)
	_, _ = w.Write([]byte(body))
}
