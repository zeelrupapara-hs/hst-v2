package ddequotes

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"hstquote/model"
)

// Settings holds resolved DDE socket options for a quote feed.
type Settings struct {
	Name              string
	Host              string
	Port              string
	Token             string
	ReconnectInt      time.Duration
	DialTimeout       time.Duration
	FirstFrameTimeout time.Duration
	IdleTimeout       time.Duration
	VolumeIndex       int
}

// FromFeed builds DDE settings from a datafeed row and params.
func FromFeed(feed model.QuoteFeed) (Settings, error) {
	s := Settings{
		Name:              feed.Datafeed.Name,
		ReconnectInt:      5 * time.Second,
		DialTimeout:       5 * time.Second,
		FirstFrameTimeout: 30 * time.Second,
		IdleTimeout:       90 * time.Second,
		VolumeIndex:       -1,
	}
	if feed.Datafeed.TimeoutReconnect > 0 {
		s.ReconnectInt = time.Duration(feed.Datafeed.TimeoutReconnect) * time.Second
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
	if v := set("ReconnectInterval", "timeout_reconnect"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			s.ReconnectInt = time.Duration(n) * time.Second
		}
	}
	if v := set("FirstFrameTimeout"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			s.FirstFrameTimeout = time.Duration(n) * time.Second
		}
	}
	if v := set("IdleTimeout"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			s.IdleTimeout = time.Duration(n) * time.Second
		}
	}
	if v := set("VolumeIndex"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 && n < chunkSize {
			s.VolumeIndex = n
		}
	}

	if s.Host == "" || s.Port == "" {
		return s, fmt.Errorf("feed %d: dde feed server must be host:port", feed.Datafeed.DatafeedID)
	}

	if v := set("Token", "token"); v != "" {
		s.Token = v
		return s, nil
	}
	login := set("Feed login", "Username", "username")
	if login == "" {
		login = feed.Datafeed.FeedLogin
	}
	password := set("Password", "password")
	if password == "" {
		password = feed.Datafeed.FeedPassword
	}
	token, err := generateToken(login, password)
	if err != nil {
		return s, fmt.Errorf("feed %d: %w", feed.Datafeed.DatafeedID, err)
	}
	s.Token = token
	return s, nil
}

// generateToken builds the currency server auth handshake; the escaping keeps
// credential bytes from colliding with the G)_ field delimiter.
func generateToken(username, password string) (string, error) {
	if username == "" || password == "" {
		return "", errors.New("feed login and password (or a Token param) required")
	}
	escape := func(s string) string {
		s = strings.ReplaceAll(s, ")", "%29")
		s = strings.ReplaceAll(s, "_", "%5F")
		s = strings.ReplaceAll(s, "G", "%47")
		return s
	}
	return fmt.Sprintf("&&fBI++0%%_OR4JbcDpb(X_.FD9s?<G)_%sG)_%sG)_+3!", escape(username), escape(password)), nil
}

func splitHostPort(raw string) (host, port string, ok bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", false
	}
	raw = strings.TrimPrefix(raw, "tcp://")
	i := strings.LastIndex(raw, ":")
	if i <= 0 || i == len(raw)-1 {
		return "", "", false
	}
	return raw[:i], raw[i+1:], true
}
