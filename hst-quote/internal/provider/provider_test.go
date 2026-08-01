package provider_test

import (
	"testing"

	"hstquote/internal/provider"
)

func TestModuleTypeSupportsFIXAndSimulator(t *testing.T) {
	cases := map[string]provider.ConnectorType{
		"FIXFeeder":      provider.TypeFIX,
		"fix44":          provider.TypeFIX,
		"fix43":          provider.TypeFIX,
		"fix_43":         provider.TypeFIX,
		"fix4.3":         provider.TypeFIX,
		"QuoteSimulator": provider.TypeSimulator,
		"simulator":      provider.TypeSimulator,
	}
	for mod, want := range cases {
		got, ok := provider.ModuleType(mod)
		if !ok || got != want {
			t.Fatalf("module %s: got %v %v want %v", mod, got, ok, want)
		}
	}
}

func TestModuleTypeRejectsMT5Feeder(t *testing.T) {
	unsupported := []string{
		"MetaTrader5Feeder",
		"RSSNewsFeeder",
		"UnknownFeeder",
	}
	for _, mod := range unsupported {
		if _, ok := provider.ModuleType(mod); ok {
			t.Fatalf("module %s should not be supported", mod)
		}
	}
}
