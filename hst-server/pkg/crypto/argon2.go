package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
)

// maxFieldLen bounds the decoded salt and key against a crafted PHC string.
const maxFieldLen = 1024

var (
	ErrInvalidHash  = errors.New("invalid password hash format")
	ErrHasherBusy   = errors.New("password hasher is saturated")
	ErrUnsupported  = errors.New("unsupported password hash algorithm")
	errWrongVersion = errors.New("unsupported argon2 version")
)

// Params are the argon2id cost settings.
type Params struct {
	MemoryKiB   uint32
	Time        uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// Hasher serialises argon2 work.
type Hasher struct {
	params Params
	sem    chan struct{}
	wait   time.Duration
}

// dummyHash is verified for a missing login, so timing cannot probe for accounts.
var dummyHash string

// NewHasher bounds concurrent hashing to the core count.
func NewHasher(p Params) *Hasher {
	h := &Hasher{
		params: p,
		sem:    make(chan struct{}, runtime.NumCPU()),
		wait:   2 * time.Second,
	}

	if dummyHash == "" {
		if s, err := h.HashPassword("hst-timing-equalisation-placeholder"); err == nil {
			dummyHash = s
		}
	}

	return h
}

// HashPassword returns a PHC string carrying its own salt and parameters.
func (h *Hasher) HashPassword(password string) (string, error) {
	if err := h.acquire(); err != nil {
		return "", err
	}
	defer h.release()

	salt := make([]byte, h.params.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	key := argon2.IDKey([]byte(password), salt,
		h.params.Time, h.params.MemoryKiB, h.params.Parallelism, h.params.KeyLength)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, h.params.MemoryKiB, h.params.Time, h.params.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key)), nil
}

// VerifyPassword reports a match, and whether the hash needs upgrading.
func (h *Hasher) VerifyPassword(encoded, password string) (ok bool, needsRehash bool, err error) {
	p, salt, want, err := decode(encoded)
	if err != nil {
		return false, false, err
	}

	if err := h.acquire(); err != nil {
		return false, false, err
	}
	defer h.release()

	got := argon2.IDKey([]byte(password), salt, p.Time, p.MemoryKiB, p.Parallelism, p.KeyLength)
	if subtle.ConstantTimeCompare(got, want) != 1 {
		return false, false, nil
	}

	weaker := p.MemoryKiB < h.params.MemoryKiB || p.Time < h.params.Time
	return true, weaker, nil
}

// VerifyDummy burns the same work as a real verify.
func (h *Hasher) VerifyDummy(password string) {
	if dummyHash == "" {
		return
	}
	_, _, _ = h.VerifyPassword(dummyHash, password)
}

// acquire takes a slot or gives up, so a login flood becomes 503 not an OOM.
func (h *Hasher) acquire() error {
	timer := time.NewTimer(h.wait)
	defer timer.Stop()

	select {
	case h.sem <- struct{}{}:
		return nil
	case <-timer.C:
		return ErrHasherBusy
	}
}

func (h *Hasher) release() { <-h.sem }

// decode parses a PHC string back into its parameters, salt and key.
func decode(encoded string) (Params, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return Params{}, nil, nil, ErrInvalidHash
	}
	if parts[1] != "argon2id" {
		return Params{}, nil, nil, ErrUnsupported
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return Params{}, nil, nil, ErrInvalidHash
	}
	if version != argon2.Version {
		return Params{}, nil, nil, errWrongVersion
	}

	var p Params
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d",
		&p.MemoryKiB, &p.Time, &p.Parallelism); err != nil {
		return Params{}, nil, nil, ErrInvalidHash
	}

	salt, err := base64.RawStdEncoding.Strict().DecodeString(parts[4])
	if err != nil {
		return Params{}, nil, nil, ErrInvalidHash
	}

	key, err := base64.RawStdEncoding.Strict().DecodeString(parts[5])
	if err != nil {
		return Params{}, nil, nil, ErrInvalidHash
	}

	// a PHC salt and key are tens of bytes; anything larger is malformed
	if len(salt) > maxFieldLen || len(key) > maxFieldLen || len(key) == 0 {
		return Params{}, nil, nil, ErrInvalidHash
	}
	p.SaltLength = uint32(len(salt)) // #nosec G115 -- bounded by maxFieldLen above
	p.KeyLength = uint32(len(key))   // #nosec G115 -- bounded by maxFieldLen above

	return p, salt, key, nil
}
