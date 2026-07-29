package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"hstserver/config"
	"hstserver/pkg/logger"

	"github.com/redis/go-redis/v9"
)

// Redis holds the client. Safe for concurrent use, share one per process.
type Redis struct {
	Client *redis.Client
	Log    *logger.Logger
}

// NewRedisClient connects and verifies the server is reachable.
func NewRedisClient(cfg *config.Config, log *logger.Logger) (*Redis, error) {
	opt := &redis.Options{
		Addr:     cfg.Redis.RedisUrl,
		Password: cfg.Redis.RedisPassword,
		DB:       cfg.Redis.RedisDB,

		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,

		PoolSize:        cfg.Redis.PoolSize,
		MinIdleConns:    cfg.Redis.MinIdleConns,
		ConnMaxIdleTime: cfg.Redis.ConnMaxIdleTime,
		ConnMaxLifetime: cfg.Redis.ConnMaxLifetime,

		// retry transient failures rather than surfacing them to the caller
		MaxRetries:      3,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,
	}

	// a managed redis normally requires tls; the server name is verified
	if cfg.Redis.Tls {
		host, _, err := net.SplitHostPort(cfg.Redis.RedisUrl)
		if err != nil {
			return nil, fmt.Errorf("invalid redis address %q: %w", cfg.Redis.RedisUrl, err)
		}
		opt.TLSConfig = &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}
	}

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Redis.DialTimeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("unable to ping redis on %s: %w", cfg.Redis.RedisUrl, err)
	}

	return &Redis{Client: client, Log: log}, nil
}

// Health reports whether redis still answers, for the health endpoint.
func (r *Redis) Health(ctx context.Context) error {
	if r.Client == nil {
		return fmt.Errorf("redis client not initialised")
	}
	return r.Client.Ping(ctx).Err()
}

// Close releases the connection pool.
func (r *Redis) Close() error {
	if r.Client == nil {
		return nil
	}
	return r.Client.Close()
}
