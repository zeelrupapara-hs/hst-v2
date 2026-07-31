package fixconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"hstquote/model"
)

const (
	paramFixConfigPath = "FixConfigPath"
	paramFeedLogin     = "Feed login"
)

// ResolveConfigPath returns an existing cfg path or generates one from settings.
func ResolveConfigPath(baseDir string, feed model.QuoteFeed, s Settings) (string, error) {
	if path := strings.TrimSpace(feed.Param(paramFixConfigPath)); path != "" {
		return path, nil
	}
	if path := strings.TrimSpace(feed.Datafeed.FeedServer); strings.HasSuffix(strings.ToLower(path), ".cfg") {
		return path, nil
	}
	if s.Host == "" || s.Port == "" {
		return "", fmt.Errorf("feed %d: host and port required", feed.Datafeed.DatafeedID)
	}
	return WriteConfig(baseDir, s, feed.Datafeed.DatafeedID)
}

// WriteConfig writes a QuickFIX initiator cfg file and returns its path.
func WriteConfig(baseDir string, s Settings, datafeedID int64) (string, error) {
	dir := filepath.Join(baseDir, fmt.Sprintf("feed_%d", datafeedID))
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", err
	}

	logPath := filepath.Join(dir, "fixlog")
	storePath := filepath.Join(dir, "store")
	if err := os.MkdirAll(logPath, 0o750); err != nil {
		return "", err
	}
	if err := os.MkdirAll(storePath, 0o750); err != nil {
		return "", err
	}

	cfgPath := filepath.Join(dir, "session.cfg")
	content := fmt.Sprintf(`[DEFAULT]
ConnectionType=initiator
HeartBtInt=%s
EncryptMethod=0
ReconnectInterval=%s
ResetOnLogon=%s
ResetOnLogout=Y
ResetOnDisconnect=Y
StartTime=00:00:00
EndTime=00:00:00
StartDay=Sun
EndDay=Sat
FileLogPath=%s
FileStorePath=%s
UseLocalTime=Y

[SESSION]
BeginString=%s
SenderCompID=%s
TargetCompID=%s
SocketConnectHost=%s
SocketConnectPort=%s
`, s.HeartBtInt, s.ReconnectInt, s.ResetOnLogon,
		logPath, storePath,
		s.BeginString, s.SenderCompID, s.TargetCompID, s.Host, s.Port)

	if err := os.WriteFile(cfgPath, []byte(content), 0o600); err != nil {
		return "", err
	}
	return cfgPath, nil
}

// ExternalSymbols lists LP symbols to subscribe from translates.
func ExternalSymbols(feed model.QuoteFeed) []string {
	if len(feed.Translates) == 0 {
		return nil
	}
	out := make([]string, 0, len(feed.Translates))
	seen := make(map[string]struct{}, len(feed.Translates))
	for _, tr := range feed.Translates {
		ext := tr.ExternalSymbol()
		if ext == "" {
			continue
		}
		if _, ok := seen[ext]; ok {
			continue
		}
		seen[ext] = struct{}{}
		out = append(out, ext)
	}
	return out
}

func splitHostPort(raw string) (host, port string, ok bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", false
	}
	if strings.HasPrefix(raw, "tcp://") {
		raw = strings.TrimPrefix(raw, "tcp://")
	}
	if strings.HasSuffix(strings.ToLower(raw), ".cfg") {
		return "", "", false
	}
	i := strings.LastIndex(raw, ":")
	if i <= 0 || i == len(raw)-1 {
		return "", "", false
	}
	return raw[:i], raw[i+1:], true
}
