package mailer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// The template kinds, each a directory under TemplatesDir. Names follow MT5 so a broker's
// existing template tree drops in unchanged.
const (
	KindGreeting     = "greeting"
	KindPreliminary  = "greeting/preliminary"
	KindVerifyEmail  = "verify_email"
	KindCertificate  = "certificate"
	KindConfirmation = "confirmation"
	KindStatement    = "statement"
)

// The macros a template may carry, in MT5's comment form.
const (
	MacroLogin            = "LOGIN"
	MacroPasswordMain     = "PASSWORD"
	MacroPasswordInvestor = "PASSWORD_INVESTOR"
	MacroName             = "NAME"
	MacroGroup            = "GROUP"
	MacroCompany          = "COMPANY"
	MacroConfirmationCode = "CONFIRMATION_CODE"
)

// Render reads a template and substitutes its macros. catalog is the group's own template
// folder; empty, or missing on disk, falls back to default as MT5 does.
func (m *Mailer) Render(kind, catalog string, macros map[string]string) (string, error) {
	raw, err := m.read(kind, catalog)
	if err != nil {
		return "", err
	}

	return Expand(raw, macros), nil
}

// read prefers the group's catalog and falls back to default.
func (m *Mailer) read(kind, catalog string) (string, error) {
	var tried []string

	for _, folder := range candidates(catalog) {
		path := filepath.Join(m.TemplatesDir, filepath.FromSlash(kind), folder)

		names, err := filepath.Glob(filepath.Join(path, "*.htm"))
		if err != nil || len(names) == 0 {
			tried = append(tried, path)
			continue
		}

		// one template per folder in practice; the first by name keeps the choice predictable
		b, err := os.ReadFile(names[0])
		if err != nil {
			tried = append(tried, path)
			continue
		}

		return string(b), nil
	}

	return "", fmt.Errorf("no %s template found in %s", kind, strings.Join(tried, ", "))
}

func candidates(catalog string) []string {
	catalog = strings.TrimSpace(catalog)

	// a catalog naming a path would read templates outside the tree
	if catalog == "" || strings.ContainsAny(catalog, `/\`) || catalog == ".." {
		return []string{"default"}
	}

	return []string{catalog, "default"}
}

// Expand replaces every <!--KEY--> with its value. An unknown macro is left in place rather
// than blanked, so a typo in a template is visible instead of silent.
func Expand(body string, macros map[string]string) string {
	if len(macros) == 0 {
		return body
	}

	pairs := make([]string, 0, len(macros)*2)
	for k, v := range macros {
		pairs = append(pairs, "<!--"+k+"-->", v)
	}

	return strings.NewReplacer(pairs...).Replace(body)
}
