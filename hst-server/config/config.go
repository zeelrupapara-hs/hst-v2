package config

// Config will use .ENV for docker-compose and load into config
import (
	"errors"
	"fmt"
	"math"
	"net"
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

	// database
	POSTGRES_HOST = "POSTGRES_HOST"
	POSTGRES_PORT = "POSTGRES_PORT"
	POSTGRES_USER = "POSTGRES_USER"
	// #nosec G101 -- env var name, not a credential
	POSTGRES_PASSWORD = "POSTGRES_PASSWORD"
	POSTGRES_DB       = "POSTGRES_DB"
	POSTGRES_SSL_MODE = "POSTGRES_SSL_MODE"
	POSTGRES_MAX_CONN = "POSTGRES_MAX_CONN"
	POSTGRES_MIN_CONN = "POSTGRES_MIN_CONN"

	// nats
	NATS_HOST = "NATS_HOST"
	NATS_PORT = "NATS_PORT"
	NATS_NAME = "NATS_NAME"

	// redis
	REDIS_URL       = "REDIS_URL"
	REDIS_PASSWORD  = "REDIS_PASSWORD"
	REDIS_DB        = "REDIS_DB"
	REDIS_POOL_SIZE = "REDIS_POOL_SIZE"

	// auth
	// #nosec G101 -- env var name, not a credential
	AUTH_LOG_RESET_CODES = "AUTH_LOG_RESET_CODES"
	// #nosec G101 -- env var name, not a credential
	AUTH_LOG_CREDENTIALS   = "AUTH_LOG_CREDENTIALS"
	AUTH_JWT_PRIVATE_KEY   = "AUTH_JWT_PRIVATE_KEY"
	AUTH_ACCESS_TTL        = "AUTH_ACCESS_TTL"
	AUTH_REFRESH_TTL       = "AUTH_REFRESH_TTL"
	AUTH_ARGON2_MEMORY_KIB = "AUTH_ARGON2_MEMORY_KIB"
	AUTH_ARGON2_TIME       = "AUTH_ARGON2_TIME"
	// login throttle, lowered in tests so lockout scenarios finish in seconds
	AUTH_MAX_FAILED_ATTEMPTS = "AUTH_MAX_FAILED_ATTEMPTS"
	AUTH_MAX_FAILED_PER_IP   = "AUTH_MAX_FAILED_PER_IP"
	AUTH_LOCKOUT_SECONDS     = "AUTH_LOCKOUT_SECONDS"
	// #nosec G101 -- env var name, not a credential
	AUTH_PASSWORD_PEPPER = "AUTH_PASSWORD_PEPPER"

	// cache
	SHARD_ID              = "SHARD_ID"
	SHARD_COUNT           = "SHARD_COUNT"
	MAX_ACCOUNT_PER_SHARD = "MAX_ACCOUNT_PER_SHARD"

	// first manager password (seed)
	// #nosec G101 -- env var name, not a credential
	FIRST_MANAGER_PASSWORD = "FIRST_MANAGER_PASSWORD"

	// swagger
	SWAGGER_ENABLED = "SWAGGER_ENABLED"
	HTTP_TLS_CERT   = "HTTP_TLS_CERT"
	HTTP_TLS_KEY    = "HTTP_TLS_KEY"
	REDIS_TLS       = "REDIS_TLS"
	CORS_ORIGINS    = "CORS_ORIGINS"

	// Trusted Proxies are the value which we trust for example the nginx proxy ip address whatever request.
	TRUSTED_PROXIES = "TRUSTED_PROXIES"

	// registration
	MAIL_TEMPLATES_DIR         = "MAIL_TEMPLATES_DIR"
	MAIL_DRAIN_INTERVAL        = "MAIL_DRAIN_INTERVAL"
	REGISTER_DEMO_GROUP        = "REGISTER_DEMO_GROUP"
	REGISTER_PRELIMINARY_GROUP = "REGISTER_PRELIMINARY_GROUP"

	// internal service auth
	INTERNAL_SERVICE_TOKEN = "INTERNAL_SERVICE_TOKEN"

	INFLUX_URL    = "INFLUX_URL"
	INFLUX_TOKEN  = "INFLUX_TOKEN"
	INFLUX_ORG    = "INFLUX_ORG"
	INFLUX_BUCKET = "INFLUX_BUCKET"

	INFLUX_CANDLE_BUCKET       = "INFLUX_CANDLE_BUCKET"
	INFLUX_TICK_RETENTION_DAYS = "INFLUX_TICK_RETENTION_DAYS"
	JOURNAL_RETENTION_DAYS     = "JOURNAL_RETENTION_DAYS"
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
	Internal Internal
	Register Register
	Influx   Influx
	Mail     Mail
}

// Mail is how outgoing email is rendered and drained. The servers themselves are configured
// in the panel, not here, because MT5 keeps them as records a manager can edit.
type Mail struct {
	// TemplatesDir holds greeting/, verify_email/ and the rest, in MT5's layout.
	TemplatesDir string
	// DrainInterval is how often the outbox is swept.
	DrainInterval time.Duration
}

// Influx is the tick store the charts read their history out of.
type Influx struct {
	Url    string
	Token  string
	Org    string
	Bucket string
	// CandleBucket holds the minute bars the rollup writes; it is where deep history lives.
	CandleBucket string
	// TickRetention is how far back the raw ticks still reach.
	TickRetention time.Duration
}

// Internal holds service-to-service auth settings.
type Internal struct {
	ServiceToken string
}

// Register config for public signups.
type Register struct {
	// DemoGroup and PreliminaryGroup are where a signup lands; blank disables that type.
	DemoGroup        string
	PreliminaryGroup string
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

	// LogCredentials logs a generated account password. Development only: it is the only way
	// to see one when no mail server is configured.
	LogCredentials bool
	// LogResetCodes prints recovery codes to the log, for development only.
	LogResetCodes bool

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
	// JournalRetention is how long a journal row is kept, zero keeps every row
	JournalRetention time.Duration
}

// Logger config
type Logger struct {
	DisableCaller     bool
	DisableStacktrace bool
	Encoding          string
	Level             string
	// LogDir holds one file per day named YYYYMMDD.log, as the platform does
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
	// TrustedProxies are the addresses or CIDR ranges X-Forwarded-For is believed from.
	TrustedProxies []string
	// TlsCert and TlsKey serve https directly.
	TlsCert string
	TlsKey  string
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
	c.Setting.JournalRetention = time.Duration(getEnvAsInt(JOURNAL_RETENTION_DAYS, 0)) * 24 * time.Hour

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
	// 16MB of mail attachments plus multipart overhead must fit
	c.HTTP.BodyLimit = getEnvAsInt(HTTP_BODY_LIMIT, 20*1024*1024)
	c.HTTP.SwaggerEnabled = getEnvAsBool(SWAGGER_ENABLED, false)
	c.HTTP.CorsOrigins = splitCsv(getEnv(CORS_ORIGINS, "http://localhost:3000"))
	c.HTTP.TrustedProxies = splitCsv(getEnv(TRUSTED_PROXIES, ""))
	c.Internal.ServiceToken = getEnv(INTERNAL_SERVICE_TOKEN, "")

	// Influx, where hst-quote writes every tick and the charts read them back
	c.Influx.Url = getEnv(INFLUX_URL, "http://localhost:8086")
	c.Influx.Token = getEnv(INFLUX_TOKEN, "")
	c.Influx.Org = getEnv(INFLUX_ORG, "HybridSolutions")
	c.Influx.Bucket = getEnv(INFLUX_BUCKET, "marketwatch")
	c.Influx.CandleBucket = getEnv(INFLUX_CANDLE_BUCKET, "marketwatch_candles")
	c.Influx.TickRetention = time.Duration(getEnvAsInt(INFLUX_TICK_RETENTION_DAYS, 7)) * 24 * time.Hour
	c.HTTP.TlsCert = getEnv(HTTP_TLS_CERT, "")
	c.HTTP.TlsKey = getEnv(HTTP_TLS_KEY, "")

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
	c.Redis.Tls = getEnvAsBool(REDIS_TLS, false)
	c.Redis.MinIdleConns = 2
	c.Redis.ConnMaxIdleTime = 30 * time.Minute
	c.Redis.ConnMaxLifetime = time.Hour

	// Auth
	c.Auth.JwtPrivateKey = getEnv(AUTH_JWT_PRIVATE_KEY, "")
	c.Auth.JwtIssuer = "hstserver"
	c.Auth.LogResetCodes = getEnvAsBool(AUTH_LOG_RESET_CODES, false)
	c.Auth.LogCredentials = getEnvAsBool(AUTH_LOG_CREDENTIALS, false)
	c.Auth.AccessTTL = time.Duration(getEnvAsInt(AUTH_ACCESS_TTL, 7200)) * time.Second
	c.Auth.RefreshTTL = time.Duration(getEnvAsInt(AUTH_REFRESH_TTL, 604800)) * time.Second
	c.Auth.RefreshAbsoluteTTL = 30 * 24 * time.Hour
	c.Auth.Pepper = getEnv(AUTH_PASSWORD_PEPPER, "")
	c.Auth.Argon2MemoryKiB = getEnvAsInt(AUTH_ARGON2_MEMORY_KIB, 65536)
	c.Auth.Argon2Time = getEnvAsInt(AUTH_ARGON2_TIME, 3)
	c.Auth.Argon2Parallelism = 2
	c.Auth.Argon2SaltLength = 16
	c.Auth.Argon2KeyLength = 32
	c.Auth.MaxFailedAttempts = getEnvAsInt(AUTH_MAX_FAILED_ATTEMPTS, 10)
	c.Auth.LockoutDuration = time.Duration(getEnvAsInt(AUTH_LOCKOUT_SECONDS, 900)) * time.Second
	c.Auth.MaxFailedPerIP = getEnvAsInt(AUTH_MAX_FAILED_PER_IP, 50)
	c.Auth.FirstManagerPassword = getEnv(FIRST_MANAGER_PASSWORD, "")

	// Register
	c.Mail.TemplatesDir = getEnv(MAIL_TEMPLATES_DIR, "templates")
	c.Mail.DrainInterval = time.Duration(getEnvAsInt(MAIL_DRAIN_INTERVAL, 15)) * time.Second

	c.Register.DemoGroup = getEnv(REGISTER_DEMO_GROUP, "demo")
	c.Register.PreliminaryGroup = getEnv(REGISTER_PRELIMINARY_GROUP, "preliminary")

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
	if len(badEnv) > 0 {
		return errors.Join(badEnv...)
	}

	if c.Auth.JwtPrivateKey == "" {
		return fmt.Errorf("%s is required, run make gen-keys and put the seed in .env", AUTH_JWT_PRIVATE_KEY)
	}
	if c.Auth.RefreshAbsoluteTTL <= c.Auth.RefreshTTL {
		return fmt.Errorf("%s must be shorter than the absolute family cap of %s",
			AUTH_REFRESH_TTL, c.Auth.RefreshAbsoluteTTL)
	}
	// a typo here does not error at request time, it silently stops trusting the proxy and every client looks.
	for _, p := range c.HTTP.TrustedProxies {
		if strings.Contains(p, "/") {
			if _, _, err := net.ParseCIDR(p); err != nil {
				return fmt.Errorf("%s entry %q is not a valid CIDR range", TRUSTED_PROXIES, p)
			}
			continue
		}
		if net.ParseIP(p) == nil {
			return fmt.Errorf("%s entry %q is not a valid ip address", TRUSTED_PROXIES, p)
		}
	}

	// one without the other is a misconfiguration, not a fallback to http
	if (c.HTTP.TlsCert == "") != (c.HTTP.TlsKey == "") {
		return fmt.Errorf("%s and %s must be set together", HTTP_TLS_CERT, HTTP_TLS_KEY)
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
	// zero would lock every login on its first failure or never expire the counter
	if c.Auth.MaxFailedAttempts < 1 {
		return fmt.Errorf("%s must be at least 1", AUTH_MAX_FAILED_ATTEMPTS)
	}
	if c.Auth.MaxFailedPerIP < 1 {
		return fmt.Errorf("%s must be at least 1", AUTH_MAX_FAILED_PER_IP)
	}
	if c.Auth.LockoutDuration < time.Second {
		return fmt.Errorf("%s must be at least 1", AUTH_LOCKOUT_SECONDS)
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

// badEnv collects every malformed value so the server can report them all at once.
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

// getEnvAsInt records a set but unparsable value rather than quietly using the default.
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
