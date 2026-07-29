package jwt

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken   = errors.New("expired or invalid token")
	ErrInvalidKeySize = errors.New("ed25519 private key must be a 32 byte seed")
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
	issuer  string
	ttl     time.Duration
	parser  *jwt.Parser
}

// NewSigner takes the base64 encoded 32 byte ed25519 seed.
func NewSigner(seedB64, issuer string, ttl time.Duration) (*Signer, error) {
	seed, err := base64.StdEncoding.DecodeString(seedB64)
	if err != nil {
		return nil, err
	}
	if len(seed) != ed25519.SeedSize {
		return nil, ErrInvalidKeySize
	}

	priv := ed25519.NewKeyFromSeed(seed)

	return &Signer{
		private: priv,
		public:  priv.Public().(ed25519.PublicKey),
		issuer:  issuer,
		ttl:     ttl,
		// the allowlist stops alg=none and alg=HS256 forgeries
		parser: jwt.NewParser(
			jwt.WithValidMethods([]string{"EdDSA"}),
			jwt.WithIssuer(issuer),
			jwt.WithLeeway(60*time.Second),
			jwt.WithExpirationRequired(),
		),
	}, nil
}

// PublicKey is what other services need to verify, and nothing more.
func (s *Signer) PublicKey() ed25519.PublicKey { return s.public }

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

	return jwt.NewWithClaims(jwt.SigningMethodEdDSA, c).SignedString(s.private)
}

// Parse verifies the signature and the standard claims.
func (s *Signer) Parse(token string) (*Claims, error) {
	claims := &Claims{}

	t, err := s.parser.ParseWithClaims(token, claims, func(*jwt.Token) (any, error) {
		return s.public, nil
	})
	if err != nil || !t.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
