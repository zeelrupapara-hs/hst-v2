package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"hstcore/config"
	"hstcore/handler"
	"hstcore/pkg/db"
	"hstcore/pkg/logger"
	"hstcore/pkg/nats"
	"hstcore/pkg/redis"
)

var (
	// this name is one time only
	service = "hstcore"

	// bumped per git tag: Major.Minor.Patch
	version = "v1.0.0"
)

// Start boots the service and blocks until it has shut down cleanly.
func Start() {
	// Run() owns the resources so its defers always run.
	if code := Run(); code != 0 {
		os.Exit(code)
	}
}

func Run() int {
	// config instant
	cfg, err := config.NewConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid configuration:", err)
		return 1
	}

	// pass to logger handler instant
	log, err := logger.NewLogger(cfg)
	if err != nil {
		// no logger yet, so stderr is all we have
		fmt.Fprintln(os.Stderr, "failed to init logger:", err)
		return 1
	}
	// flush buffered entries and release the day file last of all
	defer func() {
		_ = log.Sync()
		_ = log.Close()
	}()

	log.Logger.Info("Logging started for service: ", service+"@"+version)

	// Postgres DB
	database, err := db.NewPostgresDB(cfg)
	if err != nil {
		log.Logger.Errorf("failed to connect to postgres %v", err)
		return 1
	}
	defer func() {
		log.Logger.Info("closing postgres pool")
		database.DB.Close()
	}()

	log.Logger.Info("postgres connected")

	// Nats
	natsClient, err := nats.NewNatClient(cfg, log)
	if err != nil {
		log.Logger.Errorf("failed to connect to nats %v", err)
		return 1
	}
	defer func() {
		log.Logger.Info("draining nats connection")
		if err := natsClient.Close(); err != nil {
			log.Logger.Errorf("failed to drain nats %v", err)
		}
	}()

	log.Logger.Info("nats connected")

	// Redis
	redisClient, err := redis.NewRedisClient(cfg, log)
	if err != nil {
		log.Logger.Errorf("failed to connect to redis %v", err)
		return 1
	}
	defer func() {
		log.Logger.Info("closing redis pool")
		if err := redisClient.Close(); err != nil {
			log.Logger.Errorf("failed to close redis %v", err)
		}
	}()

	log.Logger.Info("redis connected")

	// the service itself, everything above is plumbing
	h := handler.New(cfg, log, database, natsClient, redisClient)
	if err := h.Start(context.Background()); err != nil {
		log.Logger.Errorf("failed to start the handler %v", err)
		return 1
	}
	// stopped before the connections it uses are closed
	defer func() {
		log.Logger.Info("stopping handler")
		h.Stop()
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	log.Logger.Infow("shutdown signal received", "signal", sig.String())

	log.Logger.Info("server stopped for service: ", service+"@"+version)
	return 0
}
