package datafeedmodules

import (
	"fmt"
	"strings"

	"hstserver/model"
)

// Option is one selectable feeder module for the admin UI.
type Option struct {
	Module      string `json:"module"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type moduleDef struct {
	canonical   string
	label       string
	description string
	aliases     map[string]struct{}
}

var (
	quoteModules = []moduleDef{
		{
			canonical:   "fix44",
			label:       "FIX 4.4",
			description: "QuickFIX initiator, incremental refresh by default",
			aliases: aliasSet(
				"fix44", "fix", "FIX", "FIX44", "FIXFeeder", "QuickFIXFeeder", "FIX44Feeder", "fix_44",
			),
		},
		{
			canonical:   "fix43",
			label:       "FIX 4.3",
			description: "QuickFIX initiator, full refresh by default",
			aliases:     aliasSet("fix43", "fix_43", "fix4.3"),
		},
	}
	newsModules = []moduleDef{
		{
			canonical:   "RSSNewsFeeder",
			label:       "RSS News",
			description: "Poll an HTTP RSS or Atom feed",
			aliases:     aliasSet("RSSNewsFeeder", "rss", "RSS"),
		},
	}
)

func aliasSet(names ...string) map[string]struct{} {
	out := make(map[string]struct{}, len(names))
	for _, n := range names {
		out[normalizeKey(n)] = struct{}{}
	}
	return out
}

func normalizeKey(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(strings.ToLower(s), ".exe")
	return s
}

func lookup(defs []moduleDef, module string) (moduleDef, bool) {
	key := normalizeKey(module)
	for _, d := range defs {
		if normalizeKey(d.canonical) == key {
			return d, true
		}
		if _, ok := d.aliases[key]; ok {
			return d, true
		}
	}
	return moduleDef{}, false
}

// Canonical normalizes aliases to the stored module name.
func Canonical(module string) (string, bool) {
	module = strings.TrimSpace(module)
	if module == "" {
		return "", false
	}
	if d, ok := lookup(quoteModules, module); ok {
		return d.canonical, true
	}
	if d, ok := lookup(newsModules, module); ok {
		return d.canonical, true
	}
	return "", false
}

// Validate checks module against feeder mode flags.
func Validate(module string, mode model.FeederFlags) error {
	module = strings.TrimSpace(module)
	if module == "" {
		return fmt.Errorf("module is required")
	}

	hasQuotes := mode&model.FeederFlags_quotes != 0
	hasNews := mode&model.FeederFlags_news != 0
	hasRemote := mode&model.FeederFlags_remote != 0

	if hasQuotes && hasNews {
		return fmt.Errorf("combined quotes and news mode is not supported; use separate datafeeds")
	}

	if hasQuotes {
		if _, ok := lookup(quoteModules, module); ok {
			return nil
		}
		return fmt.Errorf("unsupported quotes module %q; use one of: %s",
			module, strings.Join(QuoteModuleNames(), ", "))
	}

	if hasNews {
		if _, ok := lookup(newsModules, module); ok {
			return nil
		}
		return fmt.Errorf("unsupported news module %q; use one of: %s",
			module, strings.Join(NewsModuleNames(), ", "))
	}

	if hasRemote {
		return fmt.Errorf("remote-only datafeeds are not supported yet")
	}

	return fmt.Errorf("mode must include quotes or news")
}

// Options returns UI options for one mode flag (quotes=1, news=2).
func Options(mode model.FeederFlags) []Option {
	switch {
	case mode&model.FeederFlags_quotes != 0 && mode&model.FeederFlags_news == 0:
		return toOptions(quoteModules)
	case mode&model.FeederFlags_news != 0 && mode&model.FeederFlags_quotes == 0:
		return toOptions(newsModules)
	default:
		return nil
	}
}

func toOptions(defs []moduleDef) []Option {
	out := make([]Option, 0, len(defs))
	for _, d := range defs {
		out = append(out, Option{
			Module:      d.canonical,
			Label:       d.label,
			Description: d.description,
		})
	}
	return out
}

// QuoteModuleNames returns canonical quote module ids.
func QuoteModuleNames() []string {
	return canonicalNames(quoteModules)
}

// NewsModuleNames returns canonical news module ids.
func NewsModuleNames() []string {
	return canonicalNames(newsModules)
}

func canonicalNames(defs []moduleDef) []string {
	out := make([]string, len(defs))
	for i, d := range defs {
		out[i] = d.canonical
	}
	return out
}
