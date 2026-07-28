package crypto

import (
	"strings"
	"testing"
)

// testParams keep the tests fast; production memory is set from config.
var testParams = Params{MemoryKiB: 8192, Time: 1, Parallelism: 1, SaltLength: 16, KeyLength: 32}

func TestHashAndVerify(t *testing.T) {
	h := NewHasher(testParams)

	encoded, err := h.HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if !strings.HasPrefix(encoded, "$argon2id$") {
		t.Fatalf("not a PHC string: %s", encoded)
	}

	ok, rehash, err := h.VerifyPassword(encoded, "correct horse battery staple")
	if err != nil || !ok {
		t.Fatalf("verify: ok=%v err=%v", ok, err)
	}
	if rehash {
		t.Fatal("same parameters should not ask for a rehash")
	}

	ok, _, err = h.VerifyPassword(encoded, "wrong password")
	if err != nil {
		t.Fatalf("verify wrong: %v", err)
	}
	if ok {
		t.Fatal("wrong password verified")
	}
}

// A salt is random per hash, so the same password must not produce the same string.
func TestHashIsSalted(t *testing.T) {
	h := NewHasher(testParams)

	a, _ := h.HashPassword("same")
	b, _ := h.HashPassword("same")

	if a == b {
		t.Fatal("two hashes of one password are identical, the salt is not random")
	}
}

// A hash made with weaker parameters must be flagged for upgrade on login.
func TestNeedsRehash(t *testing.T) {
	weak := NewHasher(Params{MemoryKiB: 8192, Time: 1, Parallelism: 1, SaltLength: 16, KeyLength: 32})
	strong := NewHasher(Params{MemoryKiB: 16384, Time: 2, Parallelism: 1, SaltLength: 16, KeyLength: 32})

	encoded, _ := weak.HashPassword("pw")

	ok, rehash, err := strong.VerifyPassword(encoded, "pw")
	if err != nil || !ok {
		t.Fatalf("verify: ok=%v err=%v", ok, err)
	}
	if !rehash {
		t.Fatal("a weaker hash should be flagged for rehash")
	}
}

func TestDecodeRejectsMalformed(t *testing.T) {
	h := NewHasher(testParams)

	for _, bad := range []string{
		"",
		"not-a-hash",
		"$argon2i$v=19$m=8192,t=1,p=1$c2FsdA$a2V5",
		"$argon2id$v=19$m=8192,t=1,p=1$c2FsdA",
	} {
		if ok, _, err := h.VerifyPassword(bad, "pw"); ok || err == nil {
			t.Fatalf("accepted malformed hash %q", bad)
		}
	}
}

func TestGenerateTokenIsUnique(t *testing.T) {
	seen := make(map[string]struct{}, 100)

	for i := 0; i < 100; i++ {
		tok, err := GenerateToken()
		if err != nil {
			t.Fatalf("generate: %v", err)
		}
		if _, dup := seen[tok]; dup {
			t.Fatal("duplicate token")
		}
		seen[tok] = struct{}{}
	}
}

func TestHashTokenIsStable(t *testing.T) {
	token := "abc"
	if HashTokenHex(token) != HashTokenHex("a"+"bc") {
		t.Fatal("token hash is not stable")
	}
	if HashTokenHex("abc") == HashTokenHex("abd") {
		t.Fatal("token hash collides")
	}
}
