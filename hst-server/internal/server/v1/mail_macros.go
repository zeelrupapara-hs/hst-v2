package v1

import (
	"html"
	"strconv"
	"strings"
)

// mailRecipient is one resolved recipient with everything the macros can name. The money
// figures come from the hst.accounts snapshot, so an account with open positions reads as of
// its last trade event rather than live.
type mailRecipient struct {
	Login          int64
	Email          string
	Name           string
	Group          string
	Leverage       int32
	Currency       string
	CurrencyDigits int32
	Balance        float64
	Credit         float64
	Equity         float64
	Margin         float64
	MarginFree     float64
	MarginLevel    float64
}

// expandMailMacros substitutes the #MACRO# tokens for one recipient. escape guards the HTML
// body against markup smuggled into a profile name; the subject is plain text and skips it.
// An unknown #TOKEN# is left in place where the sender can see the typo, matching how
// template expansion behaves.
func expandMailMacros(text string, r *mailRecipient, escape bool) string {
	if !strings.Contains(text, "#") {
		return text
	}

	quote := func(s string) string {
		if escape {
			return html.EscapeString(s)
		}
		return s
	}
	money := func(v float64) string {
		return strconv.FormatFloat(v, 'f', int(r.CurrencyDigits), 64)
	}

	return strings.NewReplacer(
		"#LOGIN#", strconv.FormatInt(r.Login, 10),
		"#USERNAME#", quote(r.Name),
		"#USER_CURRENCY#", quote(r.Currency),
		"#USER_BALANCE#", money(r.Balance),
		"#USER_CREDIT#", money(r.Credit),
		"#USER_EQUITY#", money(r.Equity),
		"#USER_LEVERAGE#", strconv.FormatInt(int64(r.Leverage), 10),
		"#USER_MARGIN#", money(r.Margin),
		"#USER_MARGIN_FREE#", money(r.MarginFree),
		"#USER_MARGIN_LEVEL#", strconv.FormatFloat(r.MarginLevel, 'f', 2, 64),
	).Replace(text)
}
