package utils

import (
	"crypto/rand"
	"math/big"
)

// The alphabet a generated password is drawn from. Look-alike characters are left out because
// these passwords are read off an email and typed by hand.
const (
	passwordUpper  = "ABCDEFGHJKLMNPQRSTUVWXYZ"
	passwordLower  = "abcdefghijkmnopqrstuvwxyz"
	passwordDigit  = "23456789"
	passwordSymbol = "!@#$%&*?"
)

// NewPassword returns a password with at least one character from each class. A caller never
// chooses a signup password, so this is the only thing standing between an account and a
// guessable one.
func NewPassword(length int) (string, error) {
	if length < 8 {
		length = 8
	}

	classes := []string{passwordUpper, passwordLower, passwordDigit, passwordSymbol}
	all := passwordUpper + passwordLower + passwordDigit + passwordSymbol

	out := make([]byte, 0, length)

	// one from each class first, so the result always satisfies a mixed-class policy
	for _, class := range classes {
		c, err := pick(class)
		if err != nil {
			return "", err
		}
		out = append(out, c)
	}

	for len(out) < length {
		c, err := pick(all)
		if err != nil {
			return "", err
		}
		out = append(out, c)
	}

	// without the shuffle the first four positions always hold the same classes in the same order
	for i := len(out) - 1; i > 0; i-- {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", err
		}
		j := n.Int64()
		out[i], out[j] = out[j], out[i]
	}

	return string(out), nil
}

func pick(from string) (byte, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(from))))
	if err != nil {
		return 0, err
	}

	return from[n.Int64()], nil
}
