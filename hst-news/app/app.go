package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hstnews/config"
	"hstnews/handler"
	"hstnews/internal/health"
	"hstnews/pkg/logger"
	"hstnews/pkg/nats"
	"hstnews/pkg/redis"
)

var (
	service = "hstnews"
	version = "v1.0.0"
)

func Start() {
	if code := Run(); code != 0 {
		os.Exit(code)
	}
}

func Run() int {
	cfg, err := config.NewConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid configuration:", err)
		return 1
	}

	log, err := logger.NewLogger(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to init logger:", err)
		return 1
	}
	defer func() {
		_ = log.Sync()
		_ = log.Close()
	}()

	log.Logger.Info("Logging started for service: ", service+"@"+version)

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

	probes := health.New(cfg, log, map[string]health.Check{
		"redis": redisClient.Health,
		"nats": func(context.Context) error {
			if !natsClient.NC.IsConnected() {
				return errors.New("nats is not connected")
			}
			return nil
		},
	})
	if err := probes.Start(); err != nil {
		log.Logger.Errorf("failed to start the probe server %v", err)
		return 1
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := probes.Stop(ctx); err != nil {
			log.Logger.Errorf("failed to stop the probe server %v", err)
		}
	}()

	h := handler.New(cfg, log, natsClient, redisClient)
	if err := h.Start(context.Background()); err != nil {
		log.Logger.Errorf("failed to start the handler %v", err)
		return 1
	}
	defer func() {
		log.Logger.Info("stopping handler")
		h.Stop()
	}()

	probes.Ready()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	log.Logger.Infow("shutdown signal received", "signal", sig.String())

	probes.Draining()
	log.Logger.Infow("draining before shutdown", "wait", cfg.Health.DrainWait.String())
	time.Sleep(cfg.Health.DrainWait)

	log.Logger.Info("server stopped for service: ", service+"@"+version)
	return 0
}
