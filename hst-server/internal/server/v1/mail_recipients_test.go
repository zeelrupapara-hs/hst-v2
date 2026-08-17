package v1

import (
	"reflect"
	"strings"
	"testing"

	"hstserver/pkg/mailer"
)

func TestParseMailTo(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want [][]mailTerm
	}{
		{
			name: "single login",
			in:   "1000",
			want: [][]mailTerm{{{kind: mailTermLogin, lo: 1000}}},
		},
		{
			name: "login list with spaces",
			in:   " 1000 , 1001 ",
			want: [][]mailTerm{{{kind: mailTermLogin, lo: 1000}, {kind: mailTermLogin, lo: 1001}}},
		},
		{
			name: "range",
			in:   "1000-2000",
			want: [][]mailTerm{{{kind: mailTermRange, lo: 1000, hi: 2000}}},
		},
		{
			name: "reversed range is normalised",
			in:   "2000-1000",
			want: [][]mailTerm{{{kind: mailTermRange, lo: 1000, hi: 2000}}},
		},
		{
			name: "group mask",
			in:   `group:demo\*`,
			want: [][]mailTerm{{{kind: mailTermGroup, text: `demo\*`}}},
		},
		{
			name: "country and city",
			in:   "country:Germany,city:Hamburg",
			want: [][]mailTerm{{
				{kind: mailTermCountry, text: "Germany"},
				{kind: mailTermCity, text: "Hamburg"},
			}},
		},
		{
			name: "negation",
			in:   "group:!managers",
			want: [][]mailTerm{{{kind: mailTermGroup, text: "!managers"}}},
		},
		{
			name: "negated term",
			in:   "!group:managers",
			want: [][]mailTerm{{{neg: true, kind: mailTermGroup, text: "managers"}}},
		},
		{
			name: "or groups",
			in:   `group:demo\*;1000-2000;country:India,!1005`,
			want: [][]mailTerm{
				{{kind: mailTermGroup, text: `demo\*`}},
				{{kind: mailTermRange, lo: 1000, hi: 2000}},
				{{kind: mailTermCountry, text: "India"}, {neg: true, kind: mailTermLogin, lo: 1005}},
			},
		},
		{
			name: "empty tokens are skipped",
			in:   "1000,,;;",
			want: [][]mailTerm{{{kind: mailTermLogin, lo: 1000}}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseMailTo(tc.in)
			if err != nil {
				t.Fatalf("parseMailTo(%q): %v", tc.in, err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("parseMailTo(%q) = %#v, want %#v", tc.in, got, tc.want)
			}
		})
	}
}

func TestParseMailToRejects(t *testing.T) {
	for _, in := range []string{"", " ; , ", "abc", "10a0", "x-100", "100-y", "group:", "country:", "!"} {
		if _, err := parseMailTo(in); err == nil {
			t.Errorf("parseMailTo(%q) accepted, want error", in)
		}
	}
}

func TestMailToWhere(t *testing.T) {
	groups, err := parseMailTo(`1000-2000,3000;country:India,!1005`)
	if err != nil {
		t.Fatal(err)
	}

	where, args := mailToWhere(groups, 3)

	want := "(((u.login BETWEEN $3 AND $4 OR u.login = $5))" +
		" OR ((u.country ILIKE $6) AND NOT (u.login = $7)))"
	if where != want {
		t.Fatalf("where = %s, want %s", where, want)
	}
	if !reflect.DeepEqual(args, []any{int64(1000), int64(2000), int64(3000), "India", int64(1005)}) {
		t.Fatalf("args = %#v", args)
	}
}

func TestMailToWhereKindsIntersect(t *testing.T) {
	groups, err := parseMailTo(`group:demo\*,country:Germany`)
	if err != nil {
		t.Fatal(err)
	}

	where, args := mailToWhere(groups, 1)
	if len(args) != 3 {
		t.Fatalf("args = %#v, want the mask pair and the country", args)
	}
	// the subtree mask keeps the group matcher's exact-or-prefix shape
	if want := `(((((u."group" = $1 OR starts_with(u."group", $2)))) AND (u.country ILIKE $3)))`; where != want {
		t.Fatalf("where = %s, want %s", where, want)
	}
}

func TestMailToWhereOnlyNegations(t *testing.T) {
	groups, err := parseMailTo(`!group:managers`)
	if err != nil {
		t.Fatal(err)
	}

	where, _ := mailToWhere(groups, 1)
	if want := `((NOT ((u."group" = $1))))`; where != want {
		t.Fatalf("where = %s, want %s", where, want)
	}
}

func TestLikePattern(t *testing.T) {
	if got := likePattern(`Ger*`); got != `Ger%` {
		t.Fatalf("likePattern = %q", got)
	}
	// LIKE syntax in the input matches itself literally rather than acting as a wildcard
	if got := likePattern(`a%b_c\d`); got != `a\%b\_c\\d` {
		t.Fatalf("likePattern = %q", got)
	}
}

func TestExpandMailMacros(t *testing.T) {
	r := &mailRecipient{
		Login: 1001, Name: "Ann & Co", Currency: "USD", CurrencyDigits: 2,
		Leverage: 100, Balance: 1234.5, Credit: 0, Equity: 1300.25,
		Margin: 10, MarginFree: 1290.25, MarginLevel: 13002.5,
	}

	body := expandMailMacros(
		"Hi #USERNAME#, #LOGIN# holds #USER_BALANCE# #USER_CURRENCY# at #USER_MARGIN_LEVEL#%", r, true)
	if want := "Hi Ann &amp; Co, 1001 holds 1234.50 USD at 13002.50%"; body != want {
		t.Fatalf("body = %q, want %q", body, want)
	}

	subject := expandMailMacros("For #USERNAME#", r, false)
	if want := "For Ann & Co"; subject != want {
		t.Fatalf("subject = %q, want %q", subject, want)
	}

	if got := expandMailMacros("no macros here", r, true); got != "no macros here" {
		t.Fatalf("plain text changed: %q", got)
	}
	if got := expandMailMacros("#TYPO#", r, true); got != "#TYPO#" {
		t.Fatalf("unknown macro rewritten: %q", got)
	}
}

func TestSanitizeSurvivesToolbarMarkup(t *testing.T) {
	in := `<p style="text-align:center"><b>Hi</b> <font color="#ff0000">there</font>` +
		`<script>alert(1)</script><img src="https://x/y.png"><a href="javascript:alert(1)">x</a></p>`
	out := mailer.SanitizeHTML(in)

	for _, keep := range []string{"<b>Hi</b>", "<font", "text-align", `<img src="https://x/y.png"`} {
		if !strings.Contains(out, keep) {
			t.Errorf("sanitizer stripped %q from %q", keep, out)
		}
	}
	for _, drop := range []string{"<script", "javascript:"} {
		if strings.Contains(out, drop) {
			t.Errorf("sanitizer kept %q in %q", drop, out)
		}
	}
}
