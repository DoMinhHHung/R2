package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	if err := os.MkdirAll("keys", 0o755); err != nil {
		panic(fmt.Errorf("create keys dir: %w", err))
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(fmt.Errorf("generate rsa key: %w", err))
	}

	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err := os.WriteFile(filepath.Join("keys", "private.pem"), privatePEM, 0o600); err != nil {
		panic(fmt.Errorf("write private key: %w", err))
	}

	publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		panic(fmt.Errorf("marshal public key: %w", err))
	}

	publicPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicDER,
	})
	if err := os.WriteFile(filepath.Join("keys", "public.pem"), publicPEM, 0o644); err != nil {
		panic(fmt.Errorf("write public key: %w", err))
	}
}
