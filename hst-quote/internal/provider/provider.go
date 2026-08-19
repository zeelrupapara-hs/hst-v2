package provider

import (
	"context"
	"time"
)

// ConnectorType identifies the quote wire protocol.
type ConnectorType string

const (
	TypeFIX ConnectorType = "fix"
	TypeDDE ConnectorType = "dde"
)

// RawTick is a tick from an external source before translation.
type RawTick struct {
	SourceSymbol string
	Bid          float64
	Ask          float64
	High         float64
	Low          float64
	Open         float64
	Close        float64
	Volume       float64
	Time         time.Time
	BytesRead    int64
}

// FixConfig holds FIX session settings for a datafeed.
type FixConfig struct {
	Name           string
	ConfigFilePath string
	Username       string
	Password       string
	LogPath        string
	Symbols        []string
}

// StreamConnector runs until the context is cancelled.
type StreamConnector interface {
	Type() ConnectorType
	Run(ctx context.Context, ticks chan<- RawTick) error
	Close() error
}

// ForModule returns a connector factory label for an MT5 module name.
func ModuleType(module string) (ConnectorType, bool) {
	switch module {
	case "FIXFeeder", "QuickFIXFeeder", "fix", "FIX", "FIX44", "FIX44Feeder", "fix44", "fix_44":
		return TypeFIX, true
	case "fix43", "fix_43", "fix4.3":
		return TypeFIX, true
	case "dde", "DDE", "DDEFeeder":
		return TypeDDE, true
	default:
		return "", false
	}
}
