package fixconfig_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"hstquote/internal/fixconfig"
	"hstquote/model"
)

func TestWriteConfigGeneratesFile(t *testing.T) {
	dir := t.TempDir()
	s := fixconfig.Settings{
		HeartBtInt:   "30",
		ResetOnLogon: "Y",
		ReconnectInt: "5",
		BeginString:  "FIX.4.4",
		SenderCompID: "SENDER",
		TargetCompID: "TARGET",
		Host:         "fix.example.com",
		Port:         "9876",
	}

	path, err := fixconfig.WriteConfig(dir, s, 42)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(path, "session.cfg") {
		t.Fatalf("unexpected path: %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	body := string(data)
	if !strings.Contains(body, "SocketConnectHost=fix.example.com") {
		t.Fatalf("missing host in cfg: %s", body)
	}
	if !strings.Contains(body, "SenderCompID=SENDER") {
		t.Fatalf("missing sender in cfg")
	}
	if !strings.Contains(body, "BeginString=FIX.4.4") {
		t.Fatalf("missing begin string in cfg")
	}

	if _, err := os.Stat(filepath.Join(dir, "feed_42")); err != nil {
		t.Fatal(err)
	}
}

func TestFromFeedModuleDefaults(t *testing.T) {
	base := model.QuoteFeed{
		Datafeed: model.Datafeed{FeedServer: "fix.example.com:9876"},
		Params: map[string]string{
			"SenderCompID": "S",
			"TargetCompID": "T",
		},
	}

	fix44 := base
	fix44.Datafeed.Module = "fix44"
	s44, err := fixconfig.FromFeed(fix44)
	if err != nil {
		t.Fatal(err)
	}
	if s44.Dialect != fixconfig.Dialect44 {
		t.Fatalf("dialect: got %s want %s", s44.Dialect, fixconfig.Dialect44)
	}
	if s44.MDUpdateType != "INCREMENTAL_REFRESH" {
		t.Fatalf("md update: got %s", s44.MDUpdateType)
	}
	if s44.BeginString != "FIX.4.4" {
		t.Fatalf("begin: got %s", s44.BeginString)
	}

	fix43 := base
	fix43.Datafeed.Module = "fix43"
	s43, err := fixconfig.FromFeed(fix43)
	if err != nil {
		t.Fatal(err)
	}
	if s43.Dialect != fixconfig.Dialect43 {
		t.Fatalf("dialect: got %s want %s", s43.Dialect, fixconfig.Dialect43)
	}
	if s43.MDUpdateType != "FULL_REFRESH" {
		t.Fatalf("md update: got %s", s43.MDUpdateType)
	}
	if s43.BeginString != "FIX.4.3" {
		t.Fatalf("begin: got %s", s43.BeginString)
	}
}

func TestFromFeedBeginStringOverride(t *testing.T) {
	feed := model.QuoteFeed{
		Datafeed: model.Datafeed{Module: "fix44", FeedServer: "h:1"},
		Params: map[string]string{
			"SenderCompID": "S",
			"TargetCompID": "T",
			"BeginString":  "FIX.4.3",
		},
	}
	s, err := fixconfig.FromFeed(feed)
	if err != nil {
		t.Fatal(err)
	}
	if s.Dialect != fixconfig.Dialect43 {
		t.Fatalf("dialect: got %s", s.Dialect)
	}
	if s.BeginString != "FIX.4.3" {
		t.Fatalf("begin: got %s", s.BeginString)
	}
}

func TestNormalizeMDUpdateType(t *testing.T) {
	cases := map[string]string{
		"0":                    "FULL_REFRESH",
		"1":                    "INCREMENTAL_REFRESH",
		"full_refresh":         "FULL_REFRESH",
		"INCREMENTAL":          "INCREMENTAL_REFRESH",
		"INCREMENTAL_REFRESH":  "INCREMENTAL_REFRESH",
	}
	for in, want := range cases {
		if got := fixconfig.NormalizeMDUpdateType(in); got != want {
			t.Fatalf("%q: got %q want %q", in, got, want)
		}
	}
}

func TestFromFeedMDUpdateTypeParam(t *testing.T) {
	feed := model.QuoteFeed{
		Datafeed: model.Datafeed{Module: "fix43", FeedServer: "h:1"},
		Params: map[string]string{
			"SenderCompID": "S",
			"TargetCompID": "T",
			"MDUpdateType": "1",
			"MarketDepth":  "5",
		},
	}
	s, err := fixconfig.FromFeed(feed)
	if err != nil {
		t.Fatal(err)
	}
	if s.MDUpdateType != "INCREMENTAL_REFRESH" {
		t.Fatalf("md update: got %s", s.MDUpdateType)
	}
	if s.MarketDepth != 5 {
		t.Fatalf("depth: got %d", s.MarketDepth)
	}
}

func TestExternalSymbolsEmpty(t *testing.T) {
	syms := fixconfig.ExternalSymbols(model.QuoteFeed{})
	if len(syms) != 0 {
		t.Fatalf("expected no symbols, got %v", syms)
	}
}
