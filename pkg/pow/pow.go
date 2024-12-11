package pow

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"strings"
)

func serializeHeader(header []byte, nonce uint64) []byte {
	nonceBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBytes, nonce)
	return append(header, nonceBytes...)
}

func PerformProofOfWork(header []byte, difficulty string) (uint64, string) {
	target := strings.Repeat("0", len(difficulty))
	var nonce uint64
	var hash string

	for {
		serialized := serializeHeader(header, nonce) // Use consistent serialization
		hashBytes := sha256.Sum256(serialized)
		hash = fmt.Sprintf("%x", hashBytes)

		if strings.HasPrefix(hash, target) {
			break
		}

		nonce++
		if nonce == math.MaxUint64 {
			panic("Nonce overflow, PoW failed")
		}
	}
	log.Printf("Mining Header: %x", header)

	return nonce, hash
}

func isValidHash(hash, difficultyTarget string) bool {
	return strings.HasPrefix(hash, difficultyTarget)
}

func ValidateProofOfWork(header []byte, nonce uint64, difficulty string) bool {
	serialized := serializeHeader(header, nonce) // Use consistent serialization
	hashBytes := sha256.Sum256(serialized)
	hash := fmt.Sprintf("%x", hashBytes)
	log.Printf("Validation Header: %x", header)

	return strings.HasPrefix(hash, difficulty)
}
