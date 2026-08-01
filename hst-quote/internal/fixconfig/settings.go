package fixconfig

import (
	"strconv"
	"strings"

	"hstquote/model"
)

// Dialect selects the quickfixgo message codec (fix43 vs fix44).
type Dialect string

const (
	Dialect44 Dialect = "fix44"
	Dialect43 Dialect = "fix43"
)

// Settings holds resolved FIX session options for a quote feed.
type Settings struct {
	Name         string
	Host         string
	Port         string
	SenderCompID string
	TargetCompID string
	BeginString  string
	Username     string
	Password     string
	HeartBtInt   string
	ResetOnLogon string
	ReconnectInt string
	Dialect      Dialect
	MDUpdateType string
	MarketDepth  int
}

// FromFeed builds FIX settings from a datafeed row, params, and module name.
func FromFeed(feed model.QuoteFeed) (Settings, error) {
	s := Settings{
		Name:         feed.Datafeed.Name,
		HeartBtInt:   "30",
		ResetOnLogon: "Y",
		ReconnectInt: "5",
		MarketDepth:  1,
		Password:     feed.Datafeed.FeedPassword,
	}
	applyModuleDefaults(&s, feed.Datafeed.Module)

	if feed.Datafeed.TimeoutReconnect > 0 {
		s.ReconnectInt = strconv.Itoa(int(feed.Datafeed.TimeoutReconnect))
	}
	if host, port, ok := splitHostPort(feed.Datafeed.FeedServer); ok {
		s.Host, s.Port = host, port
	}

	set := func(keys ...string) string {
		for _, k := range keys {
			if v := strings.TrimSpace(feed.Param(k)); v != "" {
				return v
			}
		}
		return ""
	}

	if v := set("SocketConnectHost", "host"); v != "" {
		s.Host = v
	}
	if v := set("SocketConnectPort", "port"); v != "" {
		s.Port = v
	}
	s.SenderCompID = set("SenderCompID", "sendercompid")
	s.TargetCompID = set("TargetCompID", "targetcompid")
	if v := set("BeginString"); v != "" {
		s.BeginString = v
		s.Dialect = dialectFromBeginString(v)
	}
	s.Username = set("Username", "username", paramFeedLogin)
	if v := set("Password", "password"); v != "" {
		s.Password = v
	}
	if v := set("HeartBtInt"); v != "" {
		s.HeartBtInt = v
	}
	if v := set("ResetOnLogon"); v != "" {
		s.ResetOnLogon = v
	}
	if v := set("ReconnectInterval", "timeout_reconnect"); v != "" {
		s.ReconnectInt = v
	}
	if v := set("MDUpdateType", "md_update_type"); v != "" {
		s.MDUpdateType = NormalizeMDUpdateType(v)
	}
	if v := set("MarketDepth", "market_depth"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			s.MarketDepth = n
		}
	}

	if s.SenderCompID == "" || s.TargetCompID == "" {
		return s, errMissingCompIDs(feed.Datafeed.DatafeedID)
	}
	return s, nil
}

func applyModuleDefaults(s *Settings, module string) {
	switch strings.ToLower(strings.TrimSpace(module)) {
	case "fix43", "fix_43", "fix4.3":
		s.Dialect = Dialect43
		s.BeginString = "FIX.4.3"
		s.MDUpdateType = "FULL_REFRESH"
	default:
		s.Dialect = Dialect44
		s.BeginString = "FIX.4.4"
		s.MDUpdateType = "INCREMENTAL_REFRESH"
	}
}

func dialectFromBeginString(bs string) Dialect {
	if strings.Contains(bs, "4.3") {
		return Dialect43
	}
	return Dialect44
}

// NormalizeMDUpdateType maps param values to FIX MDUpdateType strings.
func NormalizeMDUpdateType(v string) string {
	switch strings.ToUpper(strings.TrimSpace(v)) {
	case "0", "FULL", "FULL_REFRESH", "FULLREFRESH":
		return "FULL_REFRESH"
	case "1", "INCREMENTAL", "INCREMENTAL_REFRESH", "INCREMENTALREFRESH":
		return "INCREMENTAL_REFRESH"
	default:
		return strings.ToUpper(strings.TrimSpace(v))
	}
}

type missingCompIDsError int64

func (e missingCompIDsError) Error() string {
	return "feed " + strconv.FormatInt(int64(e), 10) + ": SenderCompID and TargetCompID params required"
}

func errMissingCompIDs(id int64) error { return missingCompIDsError(id) }
