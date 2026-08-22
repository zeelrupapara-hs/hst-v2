package harness

import (
	"encoding/json"
	"math"
	"testing"
)

// Status fails when an http status is not the one expected.
func Status(t *testing.T, got, want int, what string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: status %d, want %d", what, got, want)
	}
}

// Approx fails when two amounts differ by more than eps.
func Approx(t *testing.T, got, want, eps float64, what string) {
	t.Helper()
	if math.Abs(got-want) > eps {
		t.Fatalf("%s: got %.5f, want %.5f (eps %.5f)", what, got, want, eps)
	}
}

// Retcode fails when an order_rejected payload does not carry the retcode expected.
func Retcode(t *testing.T, payload []byte, want int) {
	t.Helper()
	var res struct {
		RetCode int    `json:"retcode"`
		Message string `json:"message"`
	}
	_ = json.Unmarshal(payload, &res)
	if res.RetCode != want {
		t.Fatalf("retcode %d (%s), want %d", res.RetCode, res.Message, want)
	}
}
