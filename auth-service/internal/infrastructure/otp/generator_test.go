package otp_test

import (
	"testing"

	"github.com/DoMinhHHung/auth-service/internal/infrastructure/otp"
)

func TestOTPGenerator(t *testing.T) {
	gen := otp.NewGenerator()

	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		code, err := gen.Generate()
		if err != nil {
			t.Fatalf("Generate() error: %v", err)
		}

		if len(code) != 6 {
			t.Errorf("OTP length should be 6, got %d", len(code))
		}

		for _, ch := range code {
			if ch < '0' || ch > '9' {
				t.Errorf("OTP should be numeric, got %q", code)
			}
		}

		seen[code] = true
	}

	// Với 100 samples, phải có ít nhất 10 unique values (thực tế sẽ nhiều hơn nhiều)
	if len(seen) < 10 {
		t.Errorf("OTP not sufficiently random: only %d unique values in 100 samples", len(seen))
	}
}
