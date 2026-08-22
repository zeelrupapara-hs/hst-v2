//go:build e2e

package trader

import (
	"os"
	"testing"

	"hste2e/harness"
)

var env *harness.Env

func TestMain(m *testing.M) {
	env = harness.Boot()
	code := m.Run()
	env.Close()
	os.Exit(code)
}
