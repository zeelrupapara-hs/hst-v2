// Package harness drives the running hst stack over its public HTTP, WS, NATS and DB surfaces.
package harness

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
)

// Env is everything a test package needs; built once by Boot in TestMain.
type Env struct {
	BaseURL       string
	NatsURL       string
	DBURL         string
	RedisAddr     string
	AdminLogin    int64
	AdminPassword string
	FeedPort      int
	// FeedMode is dde (through hst-quote) or nats (tick published straight to hst-core).
	FeedMode string

	NC    *nats.Conn
	DB    *pgxpool.Pool
	Redis *redis.Client
	Feed  *Feed
	Admin *Client
	Seed  Seed
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// Boot reads the env, checks the stack answers, logs the admin in and seeds the baseline.
func Boot() *Env {
	port, _ := strconv.Atoi(envOr("E2E_FEED_PORT", "6100"))
	adminLogin, _ := strconv.ParseInt(envOr("E2E_ADMIN_LOGIN", "1000"), 10, 64)
	e := &Env{
		BaseURL:       envOr("E2E_BASE_URL", "http://localhost:8080"),
		NatsURL:       envOr("E2E_NATS", "nats://localhost:4222"),
		DBURL:         envOr("E2E_DB", "postgres://hst:hst@localhost:5432/hst?sslmode=disable"),
		RedisAddr:     envOr("E2E_REDIS", "localhost:6379"),
		AdminLogin:    adminLogin,
		AdminPassword: envOr("E2E_ADMIN_PASSWORD", "Bootstrap-Admin-2026!"),
		FeedPort:      port,
		FeedMode:      envOr("E2E_FEED", "dde"),
	}

	resp, err := http.Get(e.BaseURL + "/api/v1/system/monitor/health")
	if err != nil || resp.StatusCode != 200 {
		log.Fatalf("hst-server not healthy at %s: %v", e.BaseURL, err)
	}
	_ = resp.Body.Close()

	if e.NC, err = nats.Connect(e.NatsURL); err != nil {
		log.Fatalf("nats: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if e.DB, err = pgxpool.New(ctx, e.DBURL); err != nil {
		log.Fatalf("db: %v", err)
	}
	if err := e.DB.Ping(ctx); err != nil {
		log.Fatalf("db ping: %v", err)
	}
	e.Redis = redis.NewClient(&redis.Options{Addr: e.RedisAddr})
	if err := e.Redis.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis: %v", err)
	}

	e.Admin = NewClient(e.BaseURL)
	if st, res := e.Admin.Login("/auth/v1/login", adminLogin, e.AdminPassword, ConnAdmin); st != 200 {
		log.Fatalf("admin login failed: %d %s", st, res.Error)
	}

	if e.Feed, err = StartFeed(e.FeedPort); err != nil {
		log.Fatalf("feed: %v", err)
	}
	if err := e.seed(); err != nil {
		log.Fatalf("seed: %v", err)
	}
	return e
}

// Close releases the connections; called from TestMain after m.Run.
func (e *Env) Close() {
	e.Feed.Close()
	e.NC.Close()
	e.DB.Close()
	_ = e.Redis.Close()
}

// Now is a unique-enough suffix for names created by one test.
func Now() string { return fmt.Sprintf("%d", time.Now().UnixNano()%1e9) }
