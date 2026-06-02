package otp

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"

	"github.com/DoMinhHHung/auth-service/internal/domain/port"
)

type secureGenerator struct{}

func NewGenerator() port.OTPGenerator {
	return &secureGenerator{}
}

func (g *secureGenerator) Generate() (string, error) {
	var n uint32
	b := make([]byte, 4)

	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}

	n = binary.BigEndian.Uint32(b) % 1_000_000

	return fmt.Sprintf("%06d", n), nil
}
