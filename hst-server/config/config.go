package config

// Config will use .ENV for docker-compose and load into config
import (
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Env vars gose here so we don't change names by mistake
const (
	BASE_URL              = "BASE_URL"
	LOG_DIR               = "LOG_DIR"
	LOG_MAX_AGE_DAYS      = "LOG_MAX_AGE_DAYS"
	HTTP_HOST             = "HTTP_HOST"
	HTTP_PORT             = "HTTP_PORT"
	HTTP_SHUTDOWN_TIMEOUT = "HTTP_SHUTDOWN_TIMEOUT"
	HTTP_BODY_LIMIT       = "HTTP_BODY_LIMIT"
	POSTGRES_HOST         = "POSTGRES_HOST"
	POSTGRES_PORT         = "POSTGRES_PORT"
	POSTGRES_USER         = "POSTGRES_USER"
	// #nosec G101 -- env var name, not a credential
	POSTGRES_PASSWORD = "POSTGRES_PASSWORD"
	POSTGRES_DB       = "POSTGRES_DB"
	POSTGRES_SSL_MODE = "POSTGRES_SSL_MODE"
	POSTGRES_MAX_CONN = "POSTGRES_MAX_CONN"
	POSTGRES_MIN_CONN = "POSTGRES_MIN_CONN"
	NATS_HOST         = "NATS_HOST"
	NATS_PORT         = "NATS_PORT"
	NATS_NAME         = "NATS_NAME"
	REDIS_URL         = "REDIS_URL"
	REDIS_PASSWORD    = "REDIS_PASSWORD"
	REDIS_DB          = "REDIS_DB"
	REDIS_POOL_SIZE   = "REDIS_POOL_SIZE"
	// #nosec G101 -- env var name, not a credential
	AUTH_JWT_PRIVATE_KEY   = "AUTH_JWT_PRIVATE_KEY"
	AUTH_ACCESS_TTL        = "AUTH_ACCESS_TTL"
	AUTH_REFRESH_TTL       = "AUTH_REFRESH_TTL"
	AUTH_ARGON2_MEMORY_KIB = "AUTH_ARGON2_MEMORY_KIB"
	AUTH_ARGON2_TIME       = "AUTH_ARGON2_TIME"
	SHARD_ID               = "SHARD_ID"
	SHARD_COUNT            = "SHARD_COUNT"
	MAX_ACCOUNT_PER_SHARD  = "MAX_ACCOUNT_PER_SHARD"
	// #nosec G101 -- env var name, not a credential
	FIRST_MANAGER_PASSWORD = "FIRST_MANAGER_PASSWORD"
	// #nosec G101 -- env var name, not a credential
	AUTH_PASSWORD_PEPPER = "AUTH_PASSWORD_PEPPER"
	SWAGGER_ENABLED      = "SWAGGER_ENABLED"
	CORS_ORIGINS         = "CORS_ORIGINS"
)

type Config struct {
	Setting  Setting
	Logger   Logger
	Postgres Postgres
	HTTP     Http
	Nats     Nats
	Redis    Redis
	Auth     Auth
	Cache    Cache
}

// Auth config
type Auth struct {
	// JwtPrivateKey is a base64 ed25519 seed.
	JwtPrivateKey string
	JwtIssuer     string
	// AccessTTL and RefreshTTL are read from the env in seconds
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	// RefreshAbsoluteTTL caps a whole rotation family however often it rotates.
	RefreshAbsoluteTTL time.Duration

	// Pepper is mixed into every password hash so a database dump alone is not crackable.
	Pepper            string
	Argon2MemoryKiB   int
	Argon2Time        int
	Argon2Parallelism int
	Argon2SaltLength  int
	Argon2KeyLength   int

	MaxFailedAttempts int
	LockoutDuration   time.Duration
	// MaxFailedPerIP throttles credential stuffing across many logins
	MaxFailedPerIP int
	// FirstManagerPassword is the first administrator's password.
	FirstManagerPassword string
}

// Cache config for the sharded in-memory session cache
type Cache struct {
	// ShardId must be unique per instance; k8s supplies the pod ordinal
	ShardId int
	// ShardCount must equal the instance count, or memory is wasted
	ShardCount int
	// MaxAccounts is the memory dial, measured 552 bytes per session
	MaxAccounts int
	// TTL is the staleness contract:
	TTL     time.Duration
	Buckets int
}

type Setting struct {
	Version   string
	LocalPath string
	BaseUrl   string
}

// Logger config
type Logger struct {
	DisableCaller     bool
	DisableStacktrace bool
	Encoding          string
	Level             string
	// LogDir holds one file per day named YYYYMMDD.log, as MT5 does
	LogDir string
	// LogMaxAgeDays prunes day files older than this; 0 keeps them forever.
	LogMaxAgeDays int
}

// Http config
type Http struct {
	Host         string
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	// ShutdownTimeout caps how long we wait for in-flight requests to drain
	ShutdownTimeout time.Duration
	// BodyLimit is the max request body in bytes
	BodyLimit int
	// SwaggerEnabled serves the api docs, keep it off in production.
	SwaggerEnabled bool
	// CorsOrigins is the allowlist, never a wildcard.
	CorsOrigins []string
}

// Postgres config
type Postgres struct {
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresSSLMode  string
	PostgresMaxConn  int32
	PostgresMinConn  int32
}

// Nats config
type Nats struct {
	Host string
	Port string
	// Name identifies this client in nats server monitoring
	Name    string
	Timeout time.Duration
	// ReconnectWait is the pause between reconnect attempts
	ReconnectWait time.Duration
	// ReconnectBufSize buffers publishes made while disconnected
	ReconnectBufSize int
	PingInterval     time.Duration
	MaxPingsOut      int
}

// Redis config
type Redis struct {
	RedisUrl      string
	RedisPassword string
	RedisDB       int

	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	PoolSize        int
	MinIdleConns    int
	ConnMaxIdleTime time.Duration
	ConnMaxLifetime time.Duration
}

// NewConfig will load the env vars into the config struct
func NewConfig() (*Config, error) {
	// init config
	http := Http{}
	setting := Setting{}
	setting.LocalPath = "./locales/*/*"
	setting.Version = "1.0.0"
	logger := Logger{}
	logger.LogDir = "logs"
	postgres := Postgres{}
	nats := Nats{}
	redis := Redis{}

	c := &Config{
		HTTP:     http,
		Logger:   logger,
		Postgres: postgres,
		Nats:     nats,
		Redis:    redis,
		Setting:  setting,
	}

	// Setting
	c.Setting.BaseUrl = getEnv(BASE_URL, "http://localhost:8080")

	// Logger
	c.Logger.LogDir = getEnv(LOG_DIR, "logs")
	c.Logger.LogMaxAgeDays = getEnvAsInt(LOG_MAX_AGE_DAYS, 365)

	// HTTP
	c.HTTP.Host = getEnv(HTTP_HOST, "0.0.0.0")
	c.HTTP.Port = getEnv(HTTP_PORT, "8080")
	c.HTTP.ReadTimeout = 10 * time.Second
	c.HTTP.WriteTimeout = 20 * time.Second
	c.HTTP.IdleTimeout = 120 * time.Second
	c.HTTP.ShutdownTimeout = time.Duration(getEnvAsInt(HTTP_SHUTDOWN_TIMEOUT, 15)) * time.Second
	c.HTTP.BodyLimit = getEnvAsInt(HTTP_BODY_LIMIT, 4*1024*1024)
	c.HTTP.SwaggerEnabled = getEnv(SWAGGER_ENABLED, "false") == "true"
	c.HTTP.CorsOrigins = splitCsv(getEnv(CORS_ORIGINS, "http://localhost:3000"))

	// Postgres
	c.Postgres.PostgresHost = getEnv(POSTGRES_HOST, "localhost")
	c.Postgres.PostgresPort = getEnv(POSTGRES_PORT, "5432")
	c.Postgres.PostgresUser = getEnv(POSTGRES_USER, "hst")
	c.Postgres.PostgresPassword = getEnv(POSTGRES_PASSWORD, "hst")
	c.Postgres.PostgresDB = getEnv(POSTGRES_DB, "hst")
	c.Postgres.PostgresSSLMode = getEnv(POSTGRES_SSL_MODE, "disable")
	c.Postgres.PostgresMaxConn = getEnvAsInt32(POSTGRES_MAX_CONN, 20)
	c.Postgres.PostgresMinConn = getEnvAsInt32(POSTGRES_MIN_CONN, 2)

	// Nats
	c.Nats.Host = getEnv(NATS_HOST, "localhost")
	c.Nats.Port = getEnv(NATS_PORT, "4222")
	c.Nats.Name = getEnv(NATS_NAME, "hstserver")
	c.Nats.Timeout = 5 * time.Second
	c.Nats.ReconnectWait = 2 * time.Second
	c.Nats.ReconnectBufSize = 8 * 1024 * 1024 // 8 MB
	c.Nats.PingInterval = 20 * time.Second
	c.Nats.MaxPingsOut = 3

	// Redis
	c.Redis.RedisUrl = getEnv(REDIS_URL, "localhost:6379")
	c.Redis.RedisPassword = getEnv(REDIS_PASSWORD, "")
	c.Redis.RedisDB = getEnvAsInt(REDIS_DB, 0)
	c.Redis.DialTimeout = 5 * time.Second
	c.Redis.ReadTimeout = 3 * time.Second
	c.Redis.WriteTimeout = 3 * time.Second
	c.Redis.PoolSize = getEnvAsInt(REDIS_POOL_SIZE, 10*runtime.NumCPU())
	c.Redis.MinIdleConns = 2
	c.Redis.ConnMaxIdleTime = 30 * time.Minute
	c.Redis.ConnMaxLifetime = time.Hour

	// Auth
	c.Auth.JwtPrivateKey = getEnv(AUTH_JWT_PRIVATE_KEY, "")
	c.Auth.JwtIssuer = "hstserver"
	c.Auth.AccessTTL = time.Duration(getEnvAsInt(AUTH_ACCESS_TTL, 7200)) * time.Second
	c.Auth.RefreshTTL = time.Duration(getEnvAsInt(AUTH_REFRESH_TTL, 604800)) * time.Second
	c.Auth.RefreshAbsoluteTTL = 30 * 24 * time.Hour
	c.Auth.Pepper = getEnv(AUTH_PASSWORD_PEPPER, "")
	c.Auth.Argon2MemoryKiB = getEnvAsInt(AUTH_ARGON2_MEMORY_KIB, 65536)
	c.Auth.Argon2Time = getEnvAsInt(AUTH_ARGON2_TIME, 3)
	c.Auth.Argon2Parallelism = 2
	c.Auth.Argon2SaltLength = 16
	c.Auth.Argon2KeyLength = 32
	c.Auth.MaxFailedAttempts = 10
	c.Auth.LockoutDuration = 15 * time.Minute
	c.Auth.MaxFailedPerIP = 50
	c.Auth.FirstManagerPassword = getEnv(FIRST_MANAGER_PASSWORD, "")

	// Cache
	c.Cache.ShardId = getEnvAsInt(SHARD_ID, 0)
	c.Cache.ShardCount = getEnvAsInt(SHARD_COUNT, 1)
	c.Cache.MaxAccounts = getEnvAsInt(MAX_ACCOUNT_PER_SHARD, 100000)
	c.Cache.TTL = 30 * time.Second
	c.Cache.Buckets = 16

	if err := c.validate(); err != nil {
		return nil, err
	}

	return c, nil
}

// validate refuses to boot on a configuration that would fail silently later.
func (c *Config) validate() error {
	if c.Auth.JwtPrivateKey == "" {
		return fmt.Errorf("%s is required, run make gen-keys and put the seed in .env", AUTH_JWT_PRIVATE_KEY)
	}
	if c.Auth.RefreshAbsoluteTTL <= c.Auth.RefreshTTL {
		return fmt.Errorf("%s must be shorter than the absolute family cap of %s",
			AUTH_REFRESH_TTL, c.Auth.RefreshAbsoluteTTL)
	}
	if c.Cache.ShardCount < 1 {
		return fmt.Errorf("%s must be at least 1", SHARD_COUNT)
	}
	if c.Cache.ShardId < 0 || c.Cache.ShardId >= c.Cache.ShardCount {
		return fmt.Errorf("%s must be between 0 and %d", SHARD_ID, c.Cache.ShardCount-1)
	}
	// bounds keep the uint32 and uint8 conversions in pkg/crypto safe
	if c.Auth.Argon2MemoryKiB < 8192 || c.Auth.Argon2MemoryKiB > 1<<20 {
		return fmt.Errorf("%s must be between 8192 and %d", AUTH_ARGON2_MEMORY_KIB, 1<<20)
	}
	if c.Auth.Argon2Time < 1 || c.Auth.Argon2Time > 16 {
		return fmt.Errorf("%s must be between 1 and 16", AUTH_ARGON2_TIME)
	}
	if c.Cache.MaxAccounts < 1 {
		return fmt.Errorf("%s must be at least 1", MAX_ACCOUNT_PER_SHARD)
	}
	return nil
}

// Dsn will return the postgres connection string
func (p *Postgres) Dsn() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		p.PostgresUser, p.PostgresPassword, p.PostgresHost,
		p.PostgresPort, p.PostgresDB, p.PostgresSSLMode)
}

// splitCsv turns a comma separated env value into a list.
func splitCsv(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultVal
}

// getEnvAsInt32 clamps to int32 range so a bad env value cannot overflow.
func getEnvAsInt32(name string, defaultVal int32) int32 {
	v := getEnvAsInt(name, int(defaultVal))
	if v < 0 || v > math.MaxInt32 {
		return defaultVal
	}
	return int32(v)
}

func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}
