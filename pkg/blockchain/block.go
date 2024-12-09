package blockchain

import (
	"crypto/sha256"

	"github.com/TalhaArjumand/ai-blockchain/pkg/ipfs"
)

type Block struct {
	Header       BlockHeader
	Transactions []Transaction
}

type BlockHeader struct {
	PreviousHash  []byte
	Timestamp     int64
	Nonce         uint64
	MerkleRoot    []byte
	Difficulty    uint32
	VMOutputsHash []byte
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
	for _, tx := range b.Transactions {
		data, algo, err := client.FetchInputs(string(tx.TxID))
		if err != nil || data == nil || algo == nil {
			return false // Validation fails if inputs are invalid or missing
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
