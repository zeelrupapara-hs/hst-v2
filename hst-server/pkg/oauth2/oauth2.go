package oauth2

import (
	"context"
	"fmt"
	"time"

	"hstserver/config"
	"hstserver/pkg/cache"
	"hstserver/pkg/crypto"
	"hstserver/pkg/db"
	"hstserver/pkg/jwt"
	"hstserver/pkg/logger"
	"hstserver/pkg/redis"
)

// Config is the per-session state handed to a caller after a successful login.
type Config struct {
	Login          int64
	ClientId       int64
	SessionId      string
	Scope          int32
	ConnectionType int32
	IpAddress      string
	UserAgent      string
	Restricted     bool
}

// OAuth2 owns every piece of authentication state that outlives one request.
type OAuth2 struct {
	// Signer mints and verifies access tokens
	Signer *jwt.Signer
	// Cache is the in process snapshot cache, sharded across instances
	Cache *cache.DistributeCache
	// Hasher does argon2id behind a concurrency limit
	Hasher *crypto.Hasher
	// Redis is the source of truth for live sessions
	Redis *redis.Redis
	// DB is read on login and refresh only, never on the request path
	DB *db.PostgresDB
	// Log
	Log *logger.Logger
	// Cfg
	Cfg *config.Config

	// OnInvalidate runs when a session or a login is revoked, on every
	// instance. The websocket hub hangs off this: a revoked session that keeps
	// its socket open would go on receiving events it is no longer entitled
	// to. Optional, and set after construction to keep pkg/oauth2 free of a
	// dependency on the transport.
	OnInvalidate func(sid string, login int64)

	// OnRefresh runs when a login's access was rewritten rather than revoked.
	// The websocket hub rebuilds that login's subscriptions in place, so a
	// manager that gains access keeps its connection instead of being signed
	// out for it.
	OnRefresh func(login int64)

	// touchCh batches last_seen_at writes so the hot path stays free of SQL
	touchCh chan string
	stop    chan struct{}
}

// NewOAuth2 wires the auth stack. It fails only on a bad signing key.
func NewOAuth2(rds *redis.Redis, database *db.PostgresDB, cfg *config.Config, log *logger.Logger) (*OAuth2, error) {
	signer, err := jwt.NewSigner(cfg.Auth.JwtPrivateKey, cfg.Auth.JwtIssuer, cfg.Auth.AccessTTL)
	if err != nil {
		return nil, fmt.Errorf("invalid %s: %w", config.AUTH_JWT_PRIVATE_KEY, err)
	}

	o := &OAuth2{
		Signer: signer,
		Cache: cache.New(
			cfg.Cache.ShardId, cfg.Cache.ShardCount,
			cfg.Cache.MaxAccounts, cfg.Cache.Buckets, cfg.Cache.TTL),
		// ranges are checked in config.validate, so these conversions are safe
		Hasher: crypto.NewHasher(crypto.Params{
			MemoryKiB:   uint32(cfg.Auth.Argon2MemoryKiB),  // #nosec G115
			Time:        uint32(cfg.Auth.Argon2Time),       // #nosec G115
			Parallelism: uint8(cfg.Auth.Argon2Parallelism), // #nosec G115
			SaltLength:  uint32(cfg.Auth.Argon2SaltLength), // #nosec G115
			KeyLength:   uint32(cfg.Auth.Argon2KeyLength),  // #nosec G115
			Pepper:      cfg.Auth.Pepper,
		}),
		Redis:   rds,
		DB:      database,
		Log:     log,
		Cfg:     cfg,
		touchCh: make(chan string, 4096),
		stop:    make(chan struct{}),
	}

	go o.Subscribe()
	go o.flushTouches()

	log.Log(logger.TypeSys, logger.CodeOK, "auth ready",
		"shard_id", cfg.Cache.ShardId,
		"shard_count", cfg.Cache.ShardCount,
		"max_accounts", cfg.Cache.MaxAccounts)

	return o, nil
}

// TouchAsync records activity without blocking.
func (o *OAuth2) TouchAsync(sid string) {
	select {
	case o.touchCh <- sid:
	default:
	}
}

// Close stops the background workers.
func (o *OAuth2) Close() {
	close(o.stop)
	o.Cache.Close()
}

// flushTouches writes the batched last_seen_at values, one statement per tick.
func (o *OAuth2) flushTouches() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	pending := make(map[string]struct{}, 256)

	flush := func() {
		if len(pending) == 0 {
			return
		}

		sids := make([]string, 0, len(pending))
		for sid := range pending {
			sids = append(sids, sid)
			delete(pending, sid)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := o.DB.DB.Exec(ctx,
			`UPDATE hst.sessions SET last_seen_at = $1 WHERE session_id = ANY($2::uuid[])`,
			time.Now().UnixNano(), sids)
		if err != nil {
			o.Log.Log(logger.TypeSys, logger.CodeWarn, "failed to flush session activity",
				"count", len(sids), "error", err.Error())
		}
	}

	for {
		select {
		case <-o.stop:
			flush()
			return
		case sid := <-o.touchCh:
			pending[sid] = struct{}{}
		case <-ticker.C:
			flush()
		}
	}
}
