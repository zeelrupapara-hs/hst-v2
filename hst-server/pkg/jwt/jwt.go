package jwt

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken   = errors.New("expired or invalid token")
	ErrInvalidKeySize = errors.New("ed25519 private key must be a 32 byte seed")
	ErrUnknownKeyID   = errors.New("token signed by an unknown key")
)

// Claims is the access token payload.
type Claims struct {
	Sid        string `json:"sid"`
	Cid        int64  `json:"cid"`
	ConnType   int32  `json:"ct"`
	Restricted bool   `json:"rst"`
	Ver        int64  `json:"ver"`

	jwt.RegisteredClaims
}

// Signer mints and verifies EdDSA access tokens.
type Signer struct {
	private ed25519.PrivateKey
	public  ed25519.PublicKey
	kid     string
	issuer  string
	ttl     time.Duration
	parser  *jwt.Parser
	// verify keys by kid, the only thing a token kid is ever looked up in
	verify map[string]ed25519.PublicKey
}

// NewSigner takes the base64 encoded 32 byte ed25519 seed.
func NewSigner(seedB64, issuer string, ttl time.Duration) (*Signer, error) {
	return NewSignerWithVerifyKeys(seedB64, issuer, ttl)
}

// NewSignerWithVerifyKeys adds old or upcoming seeds that may verify but never sign.
func NewSignerWithVerifyKeys(seedB64, issuer string, ttl time.Duration, verifySeedsB64 ...string) (*Signer, error) {
	priv, err := keyFromSeed(seedB64)
	if err != nil {
		return nil, err
	}

	pub := priv.Public().(ed25519.PublicKey)

	s := &Signer{
		private: priv,
		public:  pub,
		kid:     KeyID(pub),
		issuer:  issuer,
		ttl:     ttl,
		// the allowlist stops alg=none and alg=HS256 forgeries
		parser: jwt.NewParser(
			jwt.WithValidMethods([]string{"EdDSA"}),
			jwt.WithIssuer(issuer),
			jwt.WithLeeway(60*time.Second),
			jwt.WithExpirationRequired(),
		),
		verify: map[string]ed25519.PublicKey{},
	}

	s.verify[s.kid] = pub

	for _, v := range verifySeedsB64 {
		vp, err := keyFromSeed(v)
		if err != nil {
			return nil, err
		}

		vpub := vp.Public().(ed25519.PublicKey)
		s.verify[KeyID(vpub)] = vpub
	}

	return s, nil
}

// keyFromSeed decodes one base64 encoded 32 byte ed25519 seed.
func keyFromSeed(seedB64 string) (ed25519.PrivateKey, error) {
	seed, err := base64.StdEncoding.DecodeString(seedB64)
	if err != nil {
		return nil, err
	}
	if len(seed) != ed25519.SeedSize {
		return nil, ErrInvalidKeySize
	}

	return ed25519.NewKeyFromSeed(seed), nil
}

// KeyID names a key by the first 8 bytes of the public key hash.
func KeyID(pub ed25519.PublicKey) string {
	sum := sha256.Sum256(pub)
	return hex.EncodeToString(sum[:8])
}

// PublicKey is what other services need to verify, and nothing more.
func (s *Signer) PublicKey() ed25519.PublicKey { return s.public }

// KeyID is the kid header the signer puts on new tokens.
func (s *Signer) KeyID() string { return s.kid }

// TTL is the access token lifetime.
func (s *Signer) TTL() time.Duration { return s.ttl }

// Sign mints an access token for a session.
func (s *Signer) Sign(login int64, c *Claims) (string, error) {
	now := time.Now()

	c.RegisteredClaims = jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(login, 10),
		Issuer:    s.issuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
	}

	t := jwt.NewWithClaims(jwt.SigningMethodEdDSA, c)
	t.Header["kid"] = s.kid

	return t.SignedString(s.private)
}

// Parse verifies the signature and the standard claims.
func (s *Signer) Parse(token string) (*Claims, error) {
	claims := &Claims{}

	t, err := s.parser.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		kid, ok := t.Header["kid"].(string)
		if !ok {
			return nil, ErrUnknownKeyID
		}

		// a missing or unknown kid fails here, we never try the other keys
		pub, ok := s.verify[kid]
		if !ok {
			return nil, ErrUnknownKeyID
		}

		return pub, nil
	})
	if err != nil || !t.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
