package datafeedmodules_test

import (
	"strings"
	"testing"

	"hstserver/model"
	"hstserver/pkg/datafeedmodules"
)

func TestValidateQuoteModules(t *testing.T) {
	for _, mod := range []string{"fix44", "FIXFeeder", "QuoteSimulator", "simulator"} {
		if err := datafeedmodules.Validate(mod, model.FeederFlags_quotes); err != nil {
			t.Fatalf("%q: %v", mod, err)
		}
	}
}

func TestValidateRejectsMT5Feeder(t *testing.T) {
	err := datafeedmodules.Validate("MetaTrader5Feeder.exe", model.FeederFlags_quotes)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unsupported quotes module") {
		t.Fatalf("got %v", err)
	}
}

func TestValidateNewsModule(t *testing.T) {
	if err := datafeedmodules.Validate("RSSNewsFeeder", model.FeederFlags_news); err != nil {
		t.Fatal(err)
	}
	if err := datafeedmodules.Validate("fix44", model.FeederFlags_news); err == nil {
		t.Fatal("expected fix44 rejected for news")
	}
}

func TestCanonical(t *testing.T) {
	got, ok := datafeedmodules.Canonical("FIX44Feeder")
	if !ok || got != "fix44" {
		t.Fatalf("got %q ok=%v", got, ok)
	}
	got, ok = datafeedmodules.Canonical("MetaTrader5Feeder.exe")
	if ok {
		t.Fatalf("unexpected canonical %q", got)
	}
}

func TestOptions(t *testing.T) {
	opts := datafeedmodules.Options(model.FeederFlags_quotes)
	if len(opts) != 3 {
		t.Fatalf("got %d quote options", len(opts))
	}
	opts = datafeedmodules.Options(model.FeederFlags_news)
	if len(opts) != 1 {
		t.Fatalf("got %d news options", len(opts))
	}
}
