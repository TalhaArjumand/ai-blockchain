package pow

import (
	"strings"
	"testing"
)

func TestPerformProofOfWork(t *testing.T) {
	header := []byte("test-block-header")
	difficulty := "0000"

	nonce, hash := PerformProofOfWork(header, difficulty)

	if !strings.HasPrefix(hash, difficulty) {
		t.Errorf("PoW failed. Hash: %s does not meet target: %s", hash, difficulty)
	} else {
		t.Logf("PoW successful. Nonce: %d, Hash: %s", nonce, hash)
	}
}

func TestIsValidHash(t *testing.T) {
	hash := "0000abcd1234"
	if !isValidHash(hash, "0000") {
		t.Errorf("Expected hash %s to be valid for difficulty '0000'", hash)
	}
}

func TestProofOfWorkConsistency(t *testing.T) {
	header := []byte("test block header")
	difficulty := "00"
	nonce, hash := PerformProofOfWork(header, difficulty)

	isValid := ValidateProofOfWork(header, nonce, difficulty)
	if !isValid {
		t.Errorf("PoW validation failed. Nonce: %d, Hash: %s", nonce, hash)
	}
}
