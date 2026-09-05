package journal

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"hstcore/pkg/logger"
	"hstcore/pkg/nats"
)

// Line is one trade entry as hst-server keeps it in hst.journal.
type Line struct {
	Login   int64  `json:"login"`
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Time    int64  `json:"time"`
}

// Sink forwards every trade log line to hst-server over nats, where the admin journal reads it back.
func Sink(nc *nats.Nats) logger.Sink {
	return func(t logger.Type, c logger.Code, msg string, kv []any) {
		if t != logger.TypeTrade || nc == nil || nc.NC == nil {
			return
		}
		var login int64
		var b strings.Builder
		b.WriteString(msg)
		for i := 0; i+1 < len(kv); i += 2 {
			key := fmt.Sprint(kv[i])
			if key == "login" {
				login, _ = kv[i+1].(int64)
			}
			fmt.Fprintf(&b, " %s=%v", key, kv[i+1])
		}
		data, err := json.Marshal(Line{Login: login, Code: int32(c), Message: b.String(), Time: time.Now().UnixNano()})
		if err != nil {
			return
		}
		// best effort: the file log already has the line, a lost copy is not worth blocking a trade
		_ = nc.NC.Publish(fmt.Sprintf("hstcore.journal.%d", login), data)
	}
}
