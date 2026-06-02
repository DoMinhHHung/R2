package hash_test

import (
	"testing"

	"github.com/DoMinhHHung/auth-service/internal/infrastructure/hash"
)

func TestArgon2id_HashAndVerify(t *testing.T) {
	hasher := hash.NewArgon2idHasher()

	password := "SecurePassword123!"
	hashed, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("Hash() error: %v", err)
	}

	// Valid password
	match, err := hasher.Verify(password, hashed)
	if err != nil || !match {
		t.Error("Verify() should return true for correct password")
	}

	// Wrong password
	match, err = hasher.Verify("WrongPassword", hashed)
	if err != nil || match {
		t.Error("Verify() should return false for wrong password")
	}

	// Different hash each time (due to random salt)
	hashed2, _ := hasher.Hash(password)
	if hashed == hashed2 {
		t.Error("Hash() should produce different hashes for same password")
	}
}

func BenchmarkArgon2id_Hash(b *testing.B) {
	hasher := hash.NewArgon2idHasher()
	for i := 0; i < b.N; i++ {
		hasher.Hash("BenchmarkPassword123!")
	}
}
