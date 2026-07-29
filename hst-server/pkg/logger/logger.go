package logger

import (
	"os"

	"hstserver/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Code int

const (
	CodeOK    Code = 0 // information message
	CodeWarn  Code = 1 // warning
	CodeErr   Code = 2 // error
	CodeAtt   Code = 3 // critical error
	CodeLogin Code = 4 // system login message
)

// Type mirrors MT5's EnMTLogType — what the entry is about.
type Type int

const (
	TypeAll    Type = 0
	TypeCfg    Type = 1 // configuration changes
	TypeSys    Type = 2 // system events
	TypeNet    Type = 3 // network activity
	TypeHst    Type = 4 // price data
	TypeUser   Type = 5 // user related
	TypeTrade  Type = 6 // trade events
	TypeAPI    Type = 7 // server API
	TypeNotify Type = 8 // notifications
)

type Logger struct {
	Logger *zap.SugaredLogger
	writer *DailyWriter
}

// NewLogger returns a zap logger with two sinks: stdout and a daily file.
func NewLogger(cfg *config.Config) (*Logger, error) {
	writer, err := NewDailyWriter(cfg.Logger.LogDir, cfg.Logger.LogMaxAgeDays)
	if err != nil {
		return nil, err
	}

	encCfg := zap.NewProductionEncoderConfig()
	// full timestamp with timezone, unlike MT5 which writes time only
	encCfg.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	encCfg.TimeKey = "ts"
	encCfg.MessageKey = "msg"
	encCfg.LevelKey = "level"

	level := zap.NewAtomicLevelAt(zap.DebugLevel)

	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encCfg),
		zapcore.AddSync(writer),
		level,
	)

	consoleEnc := encCfg
	consoleEnc.EncodeLevel = zapcore.CapitalColorLevelEncoder
	stdoutCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(consoleEnc),
		zapcore.Lock(os.Stdout),
		level,
	)

	zapLogger := zap.New(
		zapcore.NewTee(stdoutCore, fileCore),
		zap.WithCaller(false),
	)

	return &Logger{Logger: zapLogger.Sugar(), writer: writer}, nil
}

// Log writes an entry tagged with MT5 type and severity codes.
func (l *Logger) Log(t Type, c Code, msg string, kv ...any) {
	fields := append([]any{"type", int(t), "code", int(c)}, kv...)
	switch c {
	case CodeErr, CodeAtt:
		l.Logger.Errorw(msg, fields...)
	case CodeWarn:
		l.Logger.Warnw(msg, fields...)
	default:
		l.Logger.Infow(msg, fields...)
	}
}

// Sync flushes buffered entries. Defer at shutdown, not in NewLogger.
func (l *Logger) Sync() error {
	_ = l.Logger.Sync()
	if l.writer != nil {
		return l.writer.Sync()
	}
	return nil
}

// Close releases the day file.
func (l *Logger) Close() error {
	if l.writer != nil {
		return l.writer.Close()
	}
	return nil
}
