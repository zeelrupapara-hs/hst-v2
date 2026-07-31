package fixquotes

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"hstquote/internal/fixconfig"
	"hstquote/internal/provider"
	"hstquote/pkg/logger"

	"github.com/quickfixgo/enum"
	"github.com/quickfixgo/field"
	fix43mdir "github.com/quickfixgo/fix43/marketdataincrementalrefresh"
	fix43mdfs "github.com/quickfixgo/fix43/marketdatasnapshotfullrefresh"
	fix44mdir "github.com/quickfixgo/fix44/marketdataincrementalrefresh"
	fix44mdfs "github.com/quickfixgo/fix44/marketdatasnapshotfullrefresh"
	"github.com/quickfixgo/quickfix"
)

type quoteBook struct {
	bid, ask, high, low, open, close, volume float64
}

// Connector runs a FIX 4.3/4.4 initiator and emits ticks.
type Connector struct {
	cfg     fixconfig.Settings
	cfgPath string
	symbols []string
	log     *logger.Logger
	ticks   chan<- provider.RawTick

	shutdown  chan struct{}
	initiator *quickfix.Initiator
	connected bool
	mu        sync.Mutex
	book      map[string]quoteBook

	mdUpdateType enum.MDUpdateType

	*quickfix.MessageRouter
}

// NewConnector builds a FIX quotes connector from resolved settings.
func NewConnector(s fixconfig.Settings, cfgPath string, symbols []string, log *logger.Logger, ticks chan<- provider.RawTick) *Connector {
	c := &Connector{
		cfg:           s,
		cfgPath:       cfgPath,
		symbols:       symbols,
		log:           log,
		ticks:         ticks,
		shutdown:      make(chan struct{}, 1),
		book:          make(map[string]quoteBook),
		mdUpdateType:  toMDUpdateEnum(s.MDUpdateType),
		MessageRouter: quickfix.NewMessageRouter(),
	}
	switch s.Dialect {
	case fixconfig.Dialect43:
		c.AddRoute(fix43mdfs.Route(c.onSnapshot43))
		c.AddRoute(fix43mdir.Route(c.onIncremental43))
	default:
		c.AddRoute(fix44mdfs.Route(c.onSnapshot44))
		c.AddRoute(fix44mdir.Route(c.onIncremental44))
	}
	return c
}

func toMDUpdateEnum(v string) enum.MDUpdateType {
	if fixconfig.NormalizeMDUpdateType(v) == "FULL_REFRESH" {
		return enum.MDUpdateType_FULL_REFRESH
	}
	return enum.MDUpdateType_INCREMENTAL_REFRESH
}

func (c *Connector) Type() provider.ConnectorType { return provider.TypeFIX }

func (c *Connector) Dialect() fixconfig.Dialect { return c.cfg.Dialect }

func (c *Connector) Run(ctx context.Context) error {
	if err := c.startInitiator(); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		c.stopInitiator()
		return ctx.Err()
	case <-c.shutdown:
		c.stopInitiator()
		return nil
	}
}

func (c *Connector) Close() error {
	select {
	case c.shutdown <- struct{}{}:
	default:
	}
	c.stopInitiator()
	return nil
}

func (c *Connector) startInitiator() error {
	f, err := os.Open(c.cfgPath)
	if err != nil {
		return fmt.Errorf("open fix cfg: %w", err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("read fix cfg: %w", err)
	}

	settings, err := quickfix.ParseSettings(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("parse fix cfg: %w", err)
	}

	store := quickfix.NewMemoryStoreFactory()
	logFactory := quickfix.NewNullLogFactory()

	initiator, err := quickfix.NewInitiator(c, store, settings, logFactory)
	if err != nil {
		return fmt.Errorf("new initiator: %w", err)
	}
	if err := initiator.Start(); err != nil {
		return fmt.Errorf("start initiator: %w", err)
	}
	c.initiator = initiator
	return nil
}

func (c *Connector) stopInitiator() {
	if c.initiator != nil {
		c.initiator.Stop()
		c.initiator = nil
	}
}

func (c *Connector) OnCreate(sessionID quickfix.SessionID) {}

func (c *Connector) OnLogon(sessionID quickfix.SessionID) {
	c.mu.Lock()
	c.connected = true
	c.mu.Unlock()

	time.Sleep(500 * time.Millisecond)
	for _, symbol := range c.symbols {
		if err := c.sendMarketRequest(symbol, sessionID); err != nil {
			c.log.Log(logger.TypeNet, logger.CodeErr, "fix subscribe failed",
				"symbol", symbol, "error", err.Error())
		}
	}
	c.log.Log(logger.TypeNet, logger.CodeOK, "fix session logged on",
		"feed", c.cfg.Name, "dialect", c.cfg.Dialect, "md_update", c.cfg.MDUpdateType,
		"symbols", len(c.symbols))
}

func (c *Connector) OnLogout(sessionID quickfix.SessionID) {
	c.mu.Lock()
	c.connected = false
	c.mu.Unlock()
}

func (c *Connector) FromAdmin(msg *quickfix.Message, sessionID quickfix.SessionID) quickfix.MessageRejectError {
	return nil
}

func (c *Connector) ToAdmin(msg *quickfix.Message, sessionID quickfix.SessionID) {
	msgType := &field.MsgTypeField{}
	if err := msg.Header.Get(msgType); err != nil {
		return
	}
	if msgType.Value() == "A" {
		if c.cfg.Username != "" {
			msg.Header.Set(field.NewUsername(c.cfg.Username))
		}
		if c.cfg.Password != "" {
			msg.Header.Set(field.NewPassword(c.cfg.Password))
		}
	}
}

func (c *Connector) ToApp(msg *quickfix.Message, sessionID quickfix.SessionID) error {
	return nil
}

func (c *Connector) FromApp(msg *quickfix.Message, sessionID quickfix.SessionID) quickfix.MessageRejectError {
	return c.Route(msg, sessionID)
}

func (c *Connector) applyMDEntry(symbol string, typ enum.MDEntryType, px, size float64) {
	c.mu.Lock()
	cur := c.book[symbol]
	cur = applyMDEntryToBook(cur, typ, px, size)
	c.book[symbol] = cur
	tick := cur
	c.mu.Unlock()

	if tick.bid > 0 || tick.ask > 0 {
		c.emitTick(symbol, tick)
	}
}

func (c *Connector) publishSnapshot(symbol string, entries quoteBook) {
	c.mu.Lock()
	c.book[symbol] = entries
	c.mu.Unlock()
	c.emitTick(symbol, entries)
}

func (c *Connector) emitTick(symbol string, book quoteBook) {
	select {
	case c.ticks <- provider.RawTick{
		SourceSymbol: symbol,
		Bid:          book.bid,
		Ask:          book.ask,
		High:         book.high,
		Low:          book.low,
		Open:         book.open,
		Close:        book.close,
		Volume:       book.volume,
		Time:         time.Now().UTC(),
	}:
	default:
		c.log.Log(logger.TypeNet, logger.CodeWarn, "fix tick channel full", "symbol", symbol)
	}
}

func applyMDEntryToBook(cur quoteBook, typ enum.MDEntryType, px, size float64) quoteBook {
	switch typ {
	case enum.MDEntryType_BID:
		cur.bid = px
	case enum.MDEntryType_OFFER:
		cur.ask = px
	case enum.MDEntryType_OPENING_PRICE:
		cur.open = px
	case enum.MDEntryType_CLOSING_PRICE, enum.MDEntryType_TRADE:
		cur.close = px
	case enum.MDEntryType_TRADING_SESSION_HIGH_PRICE:
		cur.high = px
	case enum.MDEntryType_TRADING_SESSION_LOW_PRICE:
		cur.low = px
	case enum.MDEntryType_TRADE_VOLUME:
		cur.volume = px
	}
	if size > 0 && (typ == enum.MDEntryType_TRADE || typ == enum.MDEntryType_TRADE_VOLUME) {
		cur.volume = size
	}
	return cur
}

func (c *Connector) onSnapshot44(msg fix44mdfs.MarketDataSnapshotFullRefresh, sessionID quickfix.SessionID) quickfix.MessageRejectError {
	entries, err := msg.GetNoMDEntries()
	if err != nil {
		return nil
	}
	symbol, err := msg.GetSymbol()
	if err != nil {
		return nil
	}
	var book quoteBook
	for i := 0; i < entries.Len(); i++ {
		entry := entries.Get(i)
		price, err := entry.GetMDEntryPx()
		if err != nil {
			continue
		}
		typ, err := entry.GetMDEntryType()
		if err != nil {
			continue
		}
		v, _ := price.Float64()
		size := 0.0
		if qty, err := entry.GetMDEntrySize(); err == nil {
			size, _ = qty.Float64()
		}
		book = applyMDEntryToBook(book, typ, v, size)
	}
	c.publishSnapshot(symbol, book)
	return nil
}

func (c *Connector) onSnapshot43(msg fix43mdfs.MarketDataSnapshotFullRefresh, sessionID quickfix.SessionID) quickfix.MessageRejectError {
	entries, err := msg.GetNoMDEntries()
	if err != nil {
		return nil
	}
	symbol, err := msg.GetSymbol()
	if err != nil {
		return nil
	}
	var book quoteBook
	for i := 0; i < entries.Len(); i++ {
		entry := entries.Get(i)
		price, err := entry.GetMDEntryPx()
		if err != nil {
			continue
		}
		typ, err := entry.GetMDEntryType()
		if err != nil {
			continue
		}
		v, _ := price.Float64()
		size := 0.0
		if qty, err := entry.GetMDEntrySize(); err == nil {
			size, _ = qty.Float64()
		}
		book = applyMDEntryToBook(book, typ, v, size)
	}
	c.publishSnapshot(symbol, book)
	return nil
}

func (c *Connector) onIncremental44(msg fix44mdir.MarketDataIncrementalRefresh, sessionID quickfix.SessionID) quickfix.MessageRejectError {
	entries, err := msg.GetNoMDEntries()
	if err != nil {
		return nil
	}
	for i := 0; i < entries.Len(); i++ {
		entry := entries.Get(i)
		symbol, err := entry.GetSymbol()
		if err != nil {
			continue
		}
		action, _ := entry.GetMDUpdateAction()
		if action == enum.MDUpdateAction_DELETE {
			continue
		}
		typ, err := entry.GetMDEntryType()
		if err != nil {
			continue
		}
		px, err := entry.GetMDEntryPx()
		if err != nil {
			continue
		}
		v, _ := px.Float64()
		size := 0.0
		if qty, err := entry.GetMDEntrySize(); err == nil {
			size, _ = qty.Float64()
		}
		c.applyMDEntry(symbol, typ, v, size)
	}
	return nil
}

func (c *Connector) onIncremental43(msg fix43mdir.MarketDataIncrementalRefresh, sessionID quickfix.SessionID) quickfix.MessageRejectError {
	entries, err := msg.GetNoMDEntries()
	if err != nil {
		return nil
	}
	for i := 0; i < entries.Len(); i++ {
		entry := entries.Get(i)
		symbol, err := entry.GetSymbol()
		if err != nil {
			continue
		}
		action, _ := entry.GetMDUpdateAction()
		if action == enum.MDUpdateAction_DELETE {
			continue
		}
		typ, err := entry.GetMDEntryType()
		if err != nil {
			continue
		}
		px, err := entry.GetMDEntryPx()
		if err != nil {
			continue
		}
		v, _ := px.Float64()
		size := 0.0
		if qty, err := entry.GetMDEntrySize(); err == nil {
			size, _ = qty.Float64()
		}
		c.applyMDEntry(symbol, typ, v, size)
	}
	return nil
}
