package influxdb

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"hstquote/config"
	"hstquote/model"
	"hstquote/pkg/logger"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

// Client writes translated ticks to a shared InfluxDB bucket.
type Client struct {
	client influxdb2.Client
	writer api.WriteAPI
	org    string
	bucket string
	log    *logger.Logger
}

// NewClient connects to InfluxDB when enabled in config.
func NewClient(cfg *config.Config, log *logger.Logger) (*Client, error) {
	if !cfg.Influx.Enabled {
		return nil, nil
	}
	if cfg.Influx.URL == "" || cfg.Influx.Token == "" {
		return nil, fmt.Errorf("influx enabled but URL or token is empty")
	}

	client := influxdb2.NewClient(cfg.Influx.URL, cfg.Influx.Token)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := client.Ping(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("influx ping: %w", err)
	}

	writer := client.WriteAPI(cfg.Influx.Org, cfg.Influx.Bucket)
	c := &Client{
		client: client,
		writer: writer,
		org:    cfg.Influx.Org,
		bucket: cfg.Influx.Bucket,
		log:    log,
	}

	go c.logWriteErrors()
	log.Log(logger.TypeSys, logger.CodeOK, "influx connected",
		"org", cfg.Influx.Org, "bucket", cfg.Influx.Bucket)
	return c, nil
}

func (c *Client) logWriteErrors() {
	for err := range c.writer.Errors() {
		if c.log != nil {
			c.log.Log(logger.TypeNet, logger.CodeErr, "influx write failed", "error", err.Error())
		}
	}
}

// WriteTick stores one tick; measurement is the platform symbol ID (vfxmarket-compatible).
func (c *Client) WriteTick(tick model.Tick) {
	if c == nil || c.writer == nil {
		return
	}
	if tick.SymbolID <= 0 {
		if c.log != nil {
			c.log.Log(logger.TypeNet, logger.CodeWarn, "influx write skipped, missing symbol_id",
				"symbol", tick.Symbol, "datafeed_id", tick.DatafeedID)
		}
		return
	}

	point := pointFromTick(tick)
	c.writer.WritePoint(point)
}

func pointFromTick(tick model.Tick) *write.Point {
	ts := tick.Time
	if ts.IsZero() {
		ts = time.Now().UTC()
	}

	point := influxdb2.NewPointWithMeasurement(strconv.FormatInt(tick.SymbolID, 10)).
		AddField("bid", tick.Bid).
		AddField("ask", tick.Ask).
		AddField("high", tick.High).
		AddField("low", tick.Low).
		AddField("open", tick.Open).
		AddField("close", tick.Close).
		AddField("volume", tick.Volume).
		SetTime(ts)
	return point
}

// Health pings InfluxDB.
func (c *Client) Health(ctx context.Context) error {
	if c == nil || c.client == nil {
		return nil
	}
	_, err := c.client.Ping(ctx)
	return err
}

// Close flushes pending writes and releases the client.
func (c *Client) Close() {
	if c == nil {
		return
	}
	if c.writer != nil {
		c.writer.Flush()
	}
	if c.client != nil {
		c.client.Close()
	}
}
