package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/TalhaArjumand/ai-blockchain/pkg/ipfs"
)

type Block struct {
	Header       BlockHeader
	Transactions []Transaction
	Nonce        uint64
	Hash         string
}

type BlockHeader struct {
	PreviousHash  []byte
	Timestamp     int64
	Nonce         uint64
	MerkleRoot    []byte
	Difficulty    uint32
	VMOutputsHash []byte
	Hash          []byte // Add this field
}

func (h *BlockHeader) Bytes() []byte {
	// Serialize only PreviousHash, Timestamp, and MerkleRoot
	return append(h.PreviousHash, []byte(fmt.Sprintf("%d|%x", h.Timestamp, h.MerkleRoot))...)
}

func (b *Block) ComputeMerkleRoot() {
	var transactionHashes [][]byte
	for _, tx := range b.Transactions {
		transactionHashes = append(transactionHashes, tx.TxID)
	}
	b.Header.MerkleRoot = computeMerkleRoot(transactionHashes)
}

func (b *Block) ComputeVMOutputsHash() {
	if len(b.Transactions) == 0 {
		b.Header.VMOutputsHash = nil // Explicitly set to nil for empty transactions
		return
	}

	var outputs []byte
	for _, tx := range b.Transactions {
		outputs = append(outputs, tx.VMOutput...)
	}
	hash := sha256.Sum256(outputs)
	b.Header.VMOutputsHash = hash[:]
}

func (b *Block) ValidateTransactions(client ipfs.IPFSInterface) bool {
	if len(b.Transactions) == 0 {
		return false // No transactions in block
	}

	seenTxIDs := make(map[string]bool)
	for _, tx := range b.Transactions {
		// Check for duplicate transactions
		if seenTxIDs[string(tx.TxID)] {
			return false // Duplicate transaction detected
		}
		seenTxIDs[string(tx.TxID)] = true

		// Fetch inputs from IPFS
		data, algo, err := client.FetchInputs(string(tx.TxID))
		if err != nil || data == nil || algo == nil || len(data) == 0 || len(algo) == 0 {
			return false // Validation fails if inputs are invalid or missing
		}

		// Validate metadata (if applicable)
		if len(tx.Metadata) == 0 {
			return false // Invalid metadata
		}
	}

	return true
}

func computeMerkleRoot(hashes [][]byte) []byte {
	if len(hashes) == 0 {
		return nil
	}
	for len(hashes) > 1 {
		var newLevel [][]byte
		for i := 0; i < len(hashes); i += 2 {
			if i+1 < len(hashes) {
				combined := append(hashes[i], hashes[i+1]...)
				newHash := sha256.Sum256(combined)
				newLevel = append(newLevel, newHash[:])
			} else {
				newLevel = append(newLevel, hashes[i])
			}
		}
		hashes = newLevel
	}
	return hashes[0]
}

func (b *Block) ComputeHash() []byte {
	// Serialize the block header
	headerBytes, err := json.Marshal(b.Header)
	if err != nil {
		panic("Failed to serialize block header")
	}

	// Compute the hash
	hash := sha256.Sum256(headerBytes)
	return hash[:]
}
