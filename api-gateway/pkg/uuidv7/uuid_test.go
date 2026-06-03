package uuidv7_test

import (
	"regexp"
	"testing"

	"github.com/DoMinhHHung/R2/pkg/uuidv7"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestNew_Format(t *testing.T) {
	id := uuidv7.New()
	if !uuidPattern.MatchString(id) {
		t.Errorf("New() = %q, does not match UUIDv7 format", id)
	}
}

func TestNew_Version7(t *testing.T) {
	id := uuidv7.New()
	// The 13th character (0-indexed: 14) in the UUID string is the version digit.
	// Format: xxxxxxxx-xxxx-Vxxx-xxxx-xxxxxxxxxxxx
	// Position of version: offset 14 in the raw string
	if len(id) != 36 {
		t.Fatalf("New() length = %d, want 36", len(id))
	}
	if id[14] != '7' {
		t.Errorf("New() version digit = %c, want '7'", id[14])
	}
}

func TestNew_Variant(t *testing.T) {
	id := uuidv7.New()
	// Variant bits: the 19th character (index 19) must be 8, 9, a, or b
	variantChar := id[19]
	if variantChar != '8' && variantChar != '9' && variantChar != 'a' && variantChar != 'b' {
		t.Errorf("New() variant char = %c, want one of [89ab]", variantChar)
	}
}

func TestNew_Uniqueness(t *testing.T) {
	const count = 1000
	seen := make(map[string]struct{}, count)
	for i := 0; i < count; i++ {
		id := uuidv7.New()
		if _, exists := seen[id]; exists {
			t.Errorf("New() produced duplicate UUID: %s", id)
		}
		seen[id] = struct{}{}
	}
}

func TestNew_MonotonicOrder(t *testing.T) {
	// UUIDv7 encodes timestamp in the high bits, so sequential calls should
	// produce lexicographically non-decreasing values within the same millisecond.
	prev := uuidv7.New()
	for i := 0; i < 100; i++ {
		next := uuidv7.New()
		if next < prev {
			t.Errorf("New() produced non-monotonic UUIDs: prev=%s next=%s", prev, next)
		}
		prev = next
	}
}

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		uuidv7.New()
	}
}
