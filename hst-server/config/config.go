package config

// Config will use .ENV for docker-compose and load into config
import (
	"fmt"
	"math"
	"os"
	"strconv"
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
)

type Config struct {
	Setting  Setting
	Logger   Logger
	Postgres Postgres
	HTTP     Http
	Nats     Nats
	Redis    Redis
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
	// Mirrors MT5 RequestLimitLogs: 30, 90, 180, 365, 730, 1095.
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
}

// NewConfig will load the env vars into the config struct
func NewConfig() *Config {
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

	return c
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
