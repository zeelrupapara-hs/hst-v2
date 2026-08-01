package influxdb

import (
	"context"
	"fmt"
	"strings"

	"hstquote/config"
	"hstquote/pkg/logger"

	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/domain"
)

// Ticks are the raw material, not the archive.
//
// A busy feed writes a few hundred points per symbol per minute, which is far too much to keep
// for the years a chart wants to reach back over. So the raw bucket holds only a short window,
// and a task running on the database itself rolls those ticks into minute bars in a second
// bucket that is never trimmed. A minute bar is the smallest unit any chart asks for, and every
// longer bar is built out of them on the way out.
//
// The task lives on the server rather than in this process because it must keep running when no
// feed is connected, and because it should not run once per pod.

const downsampleTaskName = "hstquote-candles-m1"

// EnsureRetention sets the buckets and the rollup task up, and leaves them alone if a previous
// run already did. Safe to call on every boot.
func (c *Client) EnsureRetention(ctx context.Context, cfg *config.Config) error {
	if c == nil {
		return nil
	}

	raw, err := c.bucketFor(ctx, cfg.Influx.Bucket, int64(cfg.Influx.TickRetention.Seconds()))
	if err != nil {
		return fmt.Errorf("raw bucket: %w", err)
	}

	// the candle bucket is never trimmed: it is the history
	if _, err := c.bucketFor(ctx, cfg.Influx.CandleBucket, 0); err != nil {
		return fmt.Errorf("candle bucket: %w", err)
	}

	c.log.Log(logger.TypeSys, logger.CodeOK, "tick store ready",
		"ticks", raw.Name, "keep_ticks", cfg.Influx.TickRetention.String(),
		"candles", cfg.Influx.CandleBucket)

	return c.ensureTask(ctx, cfg)
}

// bucketFor finds a bucket or makes it, and keeps its retention in step with the config.
func (c *Client) bucketFor(ctx context.Context, name string, keepSeconds int64) (*domain.Bucket, error) {
	buckets := c.client.BucketsAPI()

	found, err := buckets.FindBucketByName(ctx, name)
	if err == nil && found != nil {
		if retentionOf(found) != keepSeconds {
			found.RetentionRules = rulesFor(keepSeconds)
			if _, err := buckets.UpdateBucket(ctx, found); err != nil {
				return nil, err
			}
		}

		return found, nil
	}

	org, err := c.client.OrganizationsAPI().FindOrganizationByName(ctx, c.org)
	if err != nil {
		return nil, err
	}

	return buckets.CreateBucket(ctx, &domain.Bucket{
		Name:           name,
		OrgID:          org.Id,
		RetentionRules: rulesFor(keepSeconds),
	})
}

func rulesFor(keepSeconds int64) domain.RetentionRules {
	// zero is forever, which influx spells as an empty rule set
	if keepSeconds <= 0 {
		return domain.RetentionRules{}
	}

	return domain.RetentionRules{{EverySeconds: keepSeconds}}
}

func retentionOf(b *domain.Bucket) int64 {
	if len(b.RetentionRules) == 0 {
		return 0
	}

	return b.RetentionRules[0].EverySeconds
}

// ensureTask installs the minute rollup, or rewrites it when the query has changed.
func (c *Client) ensureTask(ctx context.Context, cfg *config.Config) error {
	tasks := c.client.TasksAPI()

	// influx stores the task with an options header of its own, so the body is kept apart from
	// it: the body is what we compare on, and what a fresh task is created from
	body := candleBody(cfg.Influx.Bucket, cfg.Influx.CandleBucket)
	full := taskOptions + body

	existing, err := tasks.FindTasks(ctx, &api.TaskFilter{Name: downsampleTaskName})
	if err != nil {
		return err
	}

	if len(existing) > 0 {
		task := existing[0]
		if strings.Contains(task.Flux, body) &&
			task.Status != nil && *task.Status == domain.TaskStatusTypeActive {
			return nil
		}

		task.Flux = full
		active := domain.TaskStatusTypeActive
		task.Status = &active

		if _, err := tasks.UpdateTask(ctx, &task); err != nil {
			return err
		}

		c.log.Log(logger.TypeSys, logger.CodeOK, "candle rollup updated", "task", downsampleTaskName)

		return nil
	}

	org, err := c.client.OrganizationsAPI().FindOrganizationByName(ctx, c.org)
	if err != nil {
		return err
	}

	if _, err := tasks.CreateTaskWithEvery(ctx, downsampleTaskName, body, "1m", *org.Id); err != nil {
		return err
	}

	c.log.Log(logger.TypeSys, logger.CodeOK, "candle rollup installed", "task", downsampleTaskName)

	return nil
}

// taskOptions is the header influx writes for itself when it creates a task.
const taskOptions = "option task = {name: \"" + downsampleTaskName + "\", every: 1m}\n\n"

// candleBody is the rollup itself: one minute of ticks becomes one bar per instrument.
//
// It looks two minutes back rather than one so a bar is never cut short by the task firing a
// moment early, and writing the same bar twice is harmless — a point with the same measurement,
// field and timestamp replaces the one before it.
func candleBody(from, to string) string {
	return fmt.Sprintf(`bid = from(bucket: %q)
  |> range(start: -2m)
  |> filter(fn: (r) => r._field == "bid")

vol = from(bucket: %q)
  |> range(start: -2m)
  |> filter(fn: (r) => r._field == "volume")

o = bid |> aggregateWindow(every: 1m, fn: first, timeSrc: "_start", createEmpty: false) |> set(key: "_field", value: "open")
h = bid |> aggregateWindow(every: 1m, fn: max,   timeSrc: "_start", createEmpty: false) |> set(key: "_field", value: "high")
l = bid |> aggregateWindow(every: 1m, fn: min,   timeSrc: "_start", createEmpty: false) |> set(key: "_field", value: "low")
c = bid |> aggregateWindow(every: 1m, fn: last,  timeSrc: "_start", createEmpty: false) |> set(key: "_field", value: "close")
v = vol |> aggregateWindow(every: 1m, fn: sum,   timeSrc: "_start", createEmpty: false) |> set(key: "_field", value: "volume")

union(tables: [o, h, l, c, v])
  |> to(bucket: %q)
`, from, from, to)
}
