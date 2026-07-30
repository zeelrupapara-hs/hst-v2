package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"hstcore/config"
	"hstcore/pkg/logger"
)

func newTestServer(t *testing.T, checks map[string]Check) *Server {
	t.Helper()

	// the logger writes a day file, so keep it inside the test's temp dir
	t.Setenv("LOG_DIR", t.TempDir())
	// port 0 so a test never collides with a real listener
	t.Setenv("HEALTH_PORT", "0")

	cfg, err := config.NewConfig()
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	log, err := logger.NewLogger(cfg)
	if err != nil {
		t.Fatalf("logger: %v", err)
	}
	t.Cleanup(func() { _ = log.Close() })

	return New(cfg, log, checks)
}

func status(t *testing.T, h http.HandlerFunc) int {
	t.Helper()
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	return rec.Code
}

// Liveness must not depend on anything: a database blip that fails every
// liveness probe restarts every pod at once.
func TestLivenessIgnoresDependencies(t *testing.T) {
	s := newTestServer(t, map[string]Check{
		"broken": func(context.Context) error { return errors.New("down") },
	})

	if got := status(t, s.live); got != http.StatusOK {
		t.Fatalf("expected 200, got %d", got)
	}
}

func TestReadinessFailsBeforeReady(t *testing.T) {
	s := newTestServer(t, nil)

	if got := status(t, s.readyz); got != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 while starting, got %d", got)
	}
}

func TestReadinessPassesOnceReady(t *testing.T) {
	s := newTestServer(t, map[string]Check{
		"ok": func(context.Context) error { return nil },
	})
	s.Ready()

	if got := status(t, s.readyz); got != http.StatusOK {
		t.Fatalf("expected 200, got %d", got)
	}
}

func TestReadinessFailsWhenADependencyIsDown(t *testing.T) {
	s := newTestServer(t, map[string]Check{
		"postgres": func(context.Context) error { return errors.New("down") },
	})
	s.Ready()

	if got := status(t, s.readyz); got != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", got)
	}
}

// Draining has to win over Ready, or a pod keeps taking traffic while it
// shuts down.
func TestDrainingFailsReadinessEvenWhenReady(t *testing.T) {
	s := newTestServer(t, nil)
	s.Ready()
	s.Draining()

	if got := status(t, s.readyz); got != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 while draining, got %d", got)
	}
	// liveness still passes, the process is alive and finishing its work
	if got := status(t, s.live); got != http.StatusOK {
		t.Fatalf("expected liveness to stay 200, got %d", got)
	}
}

func TestStartBindsAndStops(t *testing.T) {
	s := newTestServer(t, nil)

	if err := s.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	if err := s.Stop(context.Background()); err != nil {
		t.Fatalf("stop: %v", err)
	}
}
