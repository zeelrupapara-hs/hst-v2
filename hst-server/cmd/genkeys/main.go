// Command genkeys prints an ed25519 key pair for AUTH_JWT_PRIVATE_KEY.
package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
)

func main() {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to generate key:", err)
		os.Exit(1)
	}

	fmt.Printf("AUTH_JWT_PRIVATE_KEY=%s\n", base64.StdEncoding.EncodeToString(priv.Seed()))
	fmt.Printf("AUTH_JWT_PUBLIC_KEY=%s\n", base64.StdEncoding.EncodeToString(pub))
}
