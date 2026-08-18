package mailer

import (
	"sync"

	"github.com/microcosm-cc/bluemonday"
)

// htmlPolicy is what a composed mail may keep. The trader terminal renders mail bodies as raw
// HTML, so everything stored in hst.mails or hst.outbox must pass through here first; the
// editor is not the safety boundary, this policy is.
var htmlPolicy = sync.OnceValue(func() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()

	// the compose toolbar drives execCommand, which still emits <u> and <font>
	p.AllowElements("u", "font")
	p.AllowAttrs("color", "size", "face").OnElements("font")

	p.AllowAttrs("style").OnElements("span", "p", "div", "li", "b", "i", "u", "font")
	p.AllowStyles("color", "font-size", "text-align").Globally()

	return p
})

// SanitizeHTML strips a mail body down to the markup the compose toolbar can produce.
func SanitizeHTML(body string) string {
	return htmlPolicy().Sanitize(body)
}
