package nats

import (
	"fmt"
	"time"

	"hstserver/config"
	"hstserver/pkg/logger"

	"github.com/nats-io/nats.go"
)

type Nats struct {
	NC  *nats.Conn
	Log *logger.Logger
}

// ErrorHandler logs async errors. ErrSlowConsumer means dropped messages.
func (n *Nats) ErrorHandler(nc *nats.Conn, sub *nats.Subscription, natsErr error) {
	if natsErr == nats.ErrSlowConsumer {
		pendingMsgs, pendingBytes, err := sub.Pending()
		if err != nil {
			n.Log.Log(logger.TypeNet, logger.CodeErr, "nats slow consumer",
				"error", natsErr.Error(), "pending_err", err.Error())
			return
		}
		dropped, _ := sub.Dropped()
		n.Log.Log(logger.TypeNet, logger.CodeErr, "nats slow consumer",
			"subject", sub.Subject,
			"pending_msgs", pendingMsgs,
			"pending_bytes", pendingBytes,
			"dropped", dropped,
		)
		return
	}

	fields := []any{"error", natsErr.Error()}
	if sub != nil {
		fields = append(fields, "subject", sub.Subject)
	}
	n.Log.Log(logger.TypeNet, logger.CodeErr, "nats async error", fields...)
}

// NewNatClient will return a connected NATS client.
func NewNatClient(cfg *config.Config, log *logger.Logger) (*Nats, error) {
	url := fmt.Sprint(cfg.Nats.Host + ":" + cfg.Nats.Port)

	n := &Nats{Log: log}

	opts := []nats.Option{
		nats.Name(cfg.Nats.Name),
		nats.Timeout(cfg.Nats.Timeout),
		// reconnect forever so a broker restart is survivable
		nats.MaxReconnects(-1),
		nats.ReconnectWait(cfg.Nats.ReconnectWait),
		// buffer publishes while disconnected
		nats.ReconnectBufSize(cfg.Nats.ReconnectBufSize),
		nats.PingInterval(cfg.Nats.PingInterval),
		nats.MaxPingsOutstanding(cfg.Nats.MaxPingsOut),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			f := []any{}
			if err != nil {
				f = append(f, "error", err.Error())
			}
			log.Log(logger.TypeNet, logger.CodeWarn, "nats disconnected", f...)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Log(logger.TypeNet, logger.CodeOK, "nats reconnected",
				"url", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(_ *nats.Conn) {
			log.Log(logger.TypeNet, logger.CodeWarn, "nats connection closed")
		}),
		nats.ErrorHandler(n.ErrorHandler),
	}

	nc, err := nats.Connect(url, opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to nats on %s: %w", url, err)
	}
	n.NC = nc

	return n, nil
}

// Close drains before closing so in-flight messages are not discarded.
func (n *Nats) Close() error {
	if n.NC == nil {
		return nil
	}
	if err := n.NC.Drain(); err != nil {
		// fall back to a hard close so shutdown still completes
		n.NC.Close()
		return err
	}
	// drain is async, wait for it
	deadline := time.Now().Add(5 * time.Second)
	for n.NC.IsDraining() && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	return nil
}
