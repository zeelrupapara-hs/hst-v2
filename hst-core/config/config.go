package config

// Config will use .ENV for docker-compose and load into config
import (
	"errors"
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
	LOG_DIR          = "LOG_DIR"
	LOG_MAX_AGE_DAYS = "LOG_MAX_AGE_DAYS"

	GRPC_HOST             = "GRPC_HOST"
	GRPC_PORT             = "GRPC_PORT"
	GRPC_SHUTDOWN_TIMEOUT = "GRPC_SHUTDOWN_TIMEOUT"

	POSTGRES_HOST = "POSTGRES_HOST"
	POSTGRES_PORT = "POSTGRES_PORT"
	POSTGRES_USER = "POSTGRES_USER"
	// #nosec G101 -- env var name, not a credential
	POSTGRES_PASSWORD = "POSTGRES_PASSWORD"
	POSTGRES_DB       = "POSTGRES_DB"
	POSTGRES_SSL_MODE = "POSTGRES_SSL_MODE"
	POSTGRES_MAX_CONN = "POSTGRES_MAX_CONN"
	POSTGRES_MIN_CONN = "POSTGRES_MIN_CONN"

	HEALTH_HOST       = "HEALTH_HOST"
	HEALTH_PORT       = "HEALTH_PORT"
	HEALTH_DRAIN_WAIT = "HEALTH_DRAIN_WAIT"

	NATS_HOST = "NATS_HOST"
	NATS_PORT = "NATS_PORT"
	NATS_NAME = "NATS_NAME"

	REDIS_URL = "REDIS_URL"
	// #nosec G101 -- env var name, not a credential
	REDIS_PASSWORD  = "REDIS_PASSWORD"
	REDIS_DB        = "REDIS_DB"
	REDIS_POOL_SIZE = "REDIS_POOL_SIZE"
	REDIS_TLS       = "REDIS_TLS"
)

// Config hstcore microservice
type Config struct {
	Setting  Setting
	Logger   Logger
	GRPC     GRPC
	Health   Health
	Postgres Postgres
	Nats     Nats
	Redis    Redis
}

type Setting struct {
	Version string
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

// GRPC config
type GRPC struct {
	Host string
	Port string
	// ShutdownTimeout caps how long we wait for in-flight rpcs to drain
	ShutdownTimeout time.Duration
}

// Health config for the probe listener
type Health struct {
	Host string
	Port string
	// DrainWait is the pause between failing readiness and actually shutting
	// down, so the endpoints controller stops routing first. Set it to at
	// least twice the readiness probe period.
	DrainWait time.Duration
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

	// Tls turns on tls to redis, which a managed redis normally requires.
	Tls             bool
	PoolSize        int
	MinIdleConns    int
	ConnMaxIdleTime time.Duration
	ConnMaxLifetime time.Duration
}

// NewConfig will load the env vars into the config struct
func NewConfig() (*Config, error) {
	badEnv = nil

	c := &Config{}

	// Setting
	c.Setting.Version = "1.0.0"

	// Logger
	c.Logger.LogDir = getEnv(LOG_DIR, "logs")
	c.Logger.LogMaxAgeDays = getEnvAsInt(LOG_MAX_AGE_DAYS, 365)

	// GRPC
	c.GRPC.Host = getEnv(GRPC_HOST, "0.0.0.0")
	c.GRPC.Port = getEnv(GRPC_PORT, "3001")
	c.GRPC.ShutdownTimeout = time.Duration(getEnvAsInt(GRPC_SHUTDOWN_TIMEOUT, 15)) * time.Second

	// Health
	c.Health.Host = getEnv(HEALTH_HOST, "0.0.0.0")
	c.Health.Port = getEnv(HEALTH_PORT, "8081")
	c.Health.DrainWait = time.Duration(getEnvAsInt(HEALTH_DRAIN_WAIT, 5)) * time.Second

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
	c.Nats.Name = getEnv(NATS_NAME, "hstcore")
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
	c.Redis.Tls = getEnvAsBool(REDIS_TLS, false)
	c.Redis.MinIdleConns = 2
	c.Redis.ConnMaxIdleTime = 30 * time.Minute
	c.Redis.ConnMaxLifetime = time.Hour

	if err := c.validate(); err != nil {
		return nil, err
	}

	return c, nil
}

// validate refuses to boot on a configuration that would fail silently later.
func (c *Config) validate() error {
	if len(badEnv) > 0 {
		return errors.Join(badEnv...)
	}

	return nil
}

// Dsn will return the postgres connection string
func (p *Postgres) Dsn() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		p.PostgresUser, p.PostgresPassword, p.PostgresHost,
		p.PostgresPort, p.PostgresDB, p.PostgresSSLMode)
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultVal
}

// badEnv collects every malformed value so the service can report them all at once.
var badEnv []error

// getEnvAsInt32 clamps to int32 range so a bad env value cannot overflow.
func getEnvAsInt32(name string, defaultVal int32) int32 {
	v := getEnvAsInt(name, int(defaultVal))
	if v < 0 || v > math.MaxInt32 {
		badEnv = append(badEnv, fmt.Errorf("%s must be between 0 and %d, got %d", name, math.MaxInt32, v))
		return defaultVal
	}
	return int32(v)
}

// getEnvAsInt records a set but unparsable value rather than quietly using the
// default, which is how a setting you think you changed never takes effect.
func getEnvAsInt(name string, defaultVal int) int {
	raw, exists := os.LookupEnv(name)
	raw = strings.TrimSpace(raw)
	if !exists || raw == "" {
		return defaultVal
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		badEnv = append(badEnv, fmt.Errorf("%s must be a whole number, got %q", name, raw))
		return defaultVal
	}

	return value
}

// getEnvAsBool accepts true or false and nothing else.
func getEnvAsBool(name string, defaultVal bool) bool {
	raw, exists := os.LookupEnv(name)
	raw = strings.TrimSpace(raw)
	if !exists || raw == "" {
		return defaultVal
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		badEnv = append(badEnv, fmt.Errorf("%s must be true or false, got %q", name, raw))
		return defaultVal
	}

	return value
}
