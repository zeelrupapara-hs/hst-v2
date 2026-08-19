package ddequotes

import (
	"bufio"
	"bytes"
	"context"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"hstquote/internal/provider"
	"hstquote/pkg/logger"
)

// maxFrameSize caps the scanner buffer so a delimiter-free stream cannot grow it unbounded.
const maxFrameSize = 1 << 20

// Connector maintains a TCP session to a currency server and emits ticks.
type Connector struct {
	cfg   Settings
	log   *logger.Logger
	ticks chan<- provider.RawTick

	// OnSession, when set, hears every stream start and break; set before Run.
	OnSession func(connected bool, detail string)

	logons   atomic.Uint64
	shutdown chan struct{}

	mu   sync.Mutex
	conn net.Conn
}

// NewConnector builds a DDE quotes connector from resolved settings.
func NewConnector(s Settings, log *logger.Logger, ticks chan<- provider.RawTick) *Connector {
	return &Connector{
		cfg:      s,
		log:      log,
		ticks:    ticks,
		shutdown: make(chan struct{}, 1),
	}
}

func (c *Connector) Type() provider.ConnectorType { return provider.TypeDDE }

// Logons counts established streams; a change means the stream broke and reconnected.
func (c *Connector) Logons() uint64 { return c.logons.Load() }

func (c *Connector) Run(ctx context.Context) error {
	addr := net.JoinHostPort(c.cfg.Host, c.cfg.Port)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.shutdown:
			return nil
		default:
		}

		c.session(ctx, addr)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.shutdown:
			return nil
		case <-time.After(c.cfg.ReconnectInt):
		}
	}
}

func (c *Connector) Close() error {
	select {
	case c.shutdown <- struct{}{}:
	default:
	}
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()
	if conn != nil {
		_ = conn.Close()
	}
	return nil
}

// session dials, authenticates, and streams frames until the connection breaks.
func (c *Connector) session(ctx context.Context, addr string) {
	d := net.Dialer{Timeout: c.cfg.DialTimeout}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		c.log.Log(logger.TypeNet, logger.CodeErr, "dde dial failed",
			"feed", c.cfg.Name, "addr", addr, "error", err.Error())
		return
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		c.conn = nil
		c.mu.Unlock()
		_ = conn.Close()
	}()

	if _, err := conn.Write([]byte(c.cfg.Token)); err != nil {
		c.log.Log(logger.TypeNet, logger.CodeErr, "dde auth write failed",
			"feed", c.cfg.Name, "error", err.Error())
		return
	}

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 64<<10), maxFrameSize)
	scanner.Split(splitFrames)

	// the server never acknowledges the token, so the only proof of auth is data arriving
	_ = conn.SetReadDeadline(time.Now().Add(c.cfg.FirstFrameTimeout))

	streaming := false
	for scanner.Scan() {
		if !streaming {
			streaming = true
			// the link was down, so what follows is a break in the stream and cannot be filtered
			c.logons.Add(1)
			c.notify(true, "dde stream established")
		}
		_ = conn.SetReadDeadline(time.Now().Add(c.cfg.IdleTimeout))

		ticks, err := parseFrame(scanner.Text(), c.cfg.VolumeIndex, time.Now())
		if err != nil {
			c.log.Log(logger.TypeNet, logger.CodeWarn, "dde frame dropped",
				"feed", c.cfg.Name, "error", err.Error())
			continue
		}
		for _, tick := range ticks {
			c.emitTick(tick)
		}
	}

	detail := "dde connection lost"
	if !streaming {
		detail = "no data received after auth, check feed credentials"
	}
	if err := scanner.Err(); err != nil {
		detail += ": " + err.Error()
	}
	c.notify(false, detail)
}

func (c *Connector) notify(connected bool, detail string) {
	if c.OnSession != nil {
		c.OnSession(connected, detail)
	}
}

func (c *Connector) emitTick(tick provider.RawTick) {
	select {
	case c.ticks <- tick:
	default:
		c.log.Log(logger.TypeNet, logger.CodeWarn, "dde tick channel full",
			"feed", c.cfg.Name, "symbol", tick.SourceSymbol)
	}
}

// splitFrames is a bufio.SplitFunc for '!'-terminated frames; an incomplete
// tail at stream end is dropped.
func splitFrames(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if i := bytes.IndexByte(data, '!'); i >= 0 {
		return i + 1, data[:i], nil
	}
	if atEOF {
		return 0, nil, bufio.ErrFinalToken
	}
	return 0, nil, nil
}
