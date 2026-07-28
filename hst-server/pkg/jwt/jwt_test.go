package jwt

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"testing"
	"time"
)

func newTestSigner(t *testing.T, ttl time.Duration) *Signer {
	t.Helper()

	seed := make([]byte, ed25519.SeedSize)
	if _, err := rand.Read(seed); err != nil {
		t.Fatalf("seed: %v", err)
	}

	s, err := NewSigner(base64.StdEncoding.EncodeToString(seed), "hstserver", ttl)
	if err != nil {
		t.Fatalf("signer: %v", err)
	}

	return s
}

func TestSignAndParse(t *testing.T) {
	s := newTestSigner(t, time.Minute)

	token, err := s.Sign(1000000, &Claims{Sid: "abc", Cid: 7, ConnType: 33, Ver: 1})
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	claims, err := s.Parse(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if claims.Sid != "abc" || claims.Cid != 7 || claims.ConnType != 33 {
		t.Fatalf("claims round trip lost data: %+v", claims)
	}
	if claims.Subject != "1000000" {
		t.Fatalf("subject = %q, want 1000000", claims.Subject)
	}
}

func TestParseRejectsTamperedToken(t *testing.T) {
	s := newTestSigner(t, time.Minute)

	token, _ := s.Sign(1, &Claims{Sid: "abc"})

	// change one byte of the signature
	b := []byte(token)
	last := len(b) - 1
	if b[last] == 'A' {
		b[last] = 'B'
	} else {
		b[last] = 'A'
	}
	tampered := string(b)

	if _, err := s.Parse(tampered); err == nil {
		t.Fatal("a tampered token was accepted")
	}
}

// A token minted by one key must not verify under another.
func TestParseRejectsForeignKey(t *testing.T) {
	a := newTestSigner(t, time.Minute)
	b := newTestSigner(t, time.Minute)

	token, _ := a.Sign(1, &Claims{Sid: "abc"})

	if _, err := b.Parse(token); err == nil {
		t.Fatal("a token signed by another key was accepted")
	}
}

// alg=none is the classic JWT bypass; the allowlist must stop it.
func TestParseRejectsAlgNone(t *testing.T) {
	s := newTestSigner(t, time.Minute)

	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString(
		[]byte(`{"sid":"abc","iss":"hstserver","exp":99999999999}`))

	if _, err := s.Parse(header + "." + payload + "."); err == nil {
		t.Fatal("alg=none was accepted")
	}
}

func TestParseRejectsExpired(t *testing.T) {
	// negative TTL puts exp in the past, well beyond the 60s leeway
	s := newTestSigner(t, -5*time.Minute)

	token, _ := s.Sign(1, &Claims{Sid: "abc"})

	if _, err := s.Parse(token); err == nil {
		t.Fatal("an expired token was accepted")
	}
}

func TestNewSignerRejectsBadKey(t *testing.T) {
	if _, err := NewSigner("not-base64!!", "hstserver", time.Minute); err == nil {
		t.Fatal("accepted a non base64 key")
	}
	if _, err := NewSigner(base64.StdEncoding.EncodeToString([]byte("short")), "hstserver", time.Minute); err == nil {
		t.Fatal("accepted a key of the wrong size")
	}
}
