package config

// Config will use .ENV for docker-compose and load into config
import (
	"errors"
	"fmt"
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

	// #nosec G101 -- env var name, not a credential

	HEALTH_HOST       = "HEALTH_HOST"
	HEALTH_PORT       = "HEALTH_PORT"
	HEALTH_DRAIN_WAIT = "HEALTH_DRAIN_WAIT"

	NATS_HOST = "NATS_HOST"
	NATS_PORT = "NATS_PORT"
	NATS_NAME = "NATS_NAME"

	NEWS_RELOAD_INTERVAL = "NEWS_RELOAD_INTERVAL"

	HST_SERVER_URL         = "HST_SERVER_URL"
	INTERNAL_SERVICE_TOKEN = "INTERNAL_SERVICE_TOKEN"
	NEWS_INSTANCE_INDEX    = "NEWS_INSTANCE_INDEX"
	NEWS_INSTANCE_COUNT    = "NEWS_INSTANCE_COUNT"

	REDIS_URL = "REDIS_URL"
	// #nosec G101 -- env var name, not a credential
	REDIS_PASSWORD  = "REDIS_PASSWORD"
	REDIS_DB        = "REDIS_DB"
	REDIS_POOL_SIZE = "REDIS_POOL_SIZE"
	REDIS_TLS       = "REDIS_TLS"
)

// Config hstnews microservice
type Config struct {
	Setting Setting
	Logger  Logger
	GRPC    GRPC
	Health  Health
	Nats    Nats
	Redis   Redis
	News    News
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

// News config for ingestion behaviour.
type News struct {
	ReloadInterval time.Duration
	ServerURL      string
	ServiceToken   string
	InstanceIndex  int
	InstanceCount  int
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

	// Nats
	c.Nats.Host = getEnv(NATS_HOST, "localhost")
	c.Nats.Port = getEnv(NATS_PORT, "4222")
	c.Nats.Name = getEnv(NATS_NAME, "hstnews")
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

	c.News.ReloadInterval = time.Duration(getEnvAsInt(NEWS_RELOAD_INTERVAL, 60)) * time.Second
	c.News.ServerURL = getEnv(HST_SERVER_URL, "http://localhost:8080")
	c.News.ServiceToken = getEnv(INTERNAL_SERVICE_TOKEN, "")
	c.News.InstanceIndex = getEnvAsInt(NEWS_INSTANCE_INDEX, 0)
	c.News.InstanceCount = getEnvAsInt(NEWS_INSTANCE_COUNT, 1)

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
	if c.News.ServiceToken == "" {
		return fmt.Errorf("%s is required", INTERNAL_SERVICE_TOKEN)
	}
	if c.News.InstanceCount < 1 {
		return fmt.Errorf("%s must be at least 1", NEWS_INSTANCE_COUNT)
	}
	if c.News.InstanceIndex < 0 || c.News.InstanceIndex >= c.News.InstanceCount {
		return fmt.Errorf("%s must be between 0 and %d", NEWS_INSTANCE_INDEX, c.News.InstanceCount-1)
	}
	return nil
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultVal
}

// badEnv collects every malformed value so the service can report them all at once.
var badEnv []error

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
