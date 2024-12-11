package miner

import (
	"bytes"
	"log"
	"testing"
	"time"

	"github.com/TalhaArjumand/ai-blockchain/pkg/blockchain"
	"github.com/TalhaArjumand/ai-blockchain/pkg/ipfs"
	"github.com/TalhaArjumand/ai-blockchain/pkg/pow"
)

// Standalone test to verify PreviousHash assignment
func TestPreviousHashAssignment(t *testing.T) {
	// Step 1: Initialize blockchain and add a genesis block
	chain := blockchain.NewBlockchain()
	genesisBlock := &blockchain.Block{
		Header: blockchain.BlockHeader{
			PreviousHash: []byte("GENESIS"),
			Timestamp:    time.Now().UnixNano(),
		},
		Transactions: nil,
	}
	genesisBlock.ComputeMerkleRoot()
	genesisBlock.Header.Hash = genesisBlock.ComputeHash()
	// Log the MerkleRoot of the genesis block
	log.Printf("Genesis Block MerkleRoot: %x", genesisBlock.Header.MerkleRoot)
	err := chain.AddBlock(genesisBlock)
	if err != nil {
		t.Fatalf("Failed to add genesis block: %v", err)
	}
	log.Printf("Genesis block added: %+v", genesisBlock)

	// Step 2: Add a new block
	previousBlock := chain.Blocks[len(chain.Blocks)-1]
	newBlock := &blockchain.Block{
		Header: blockchain.BlockHeader{
			PreviousHash: previousBlock.Header.MerkleRoot, // This is the field we're verifying
			Timestamp:    time.Now().UnixNano(),
			Nonce:        0,
		},
		Transactions: nil, // No transactions for simplicity
	}
	newBlock.ComputeMerkleRoot()
	newBlock.Header.Hash = newBlock.ComputeHash()

	// Step 3: Validate PreviousHash
	if string(newBlock.Header.PreviousHash) != string(previousBlock.Header.MerkleRoot) {
		t.Fatalf("PreviousHash mismatch: expected %x, got %x", previousBlock.Header.MerkleRoot, newBlock.Header.PreviousHash)
	}

	log.Printf("PreviousHash is correctly set: %x", newBlock.Header.PreviousHash)
}

func TestIntegration_MineBlock(t *testing.T) {
	// Step 1: Initialize blockchain, mempool, and miner
	mempool := blockchain.NewMempool()
	chain := blockchain.NewBlockchain()
	genesisBlock := &blockchain.Block{
		Header: blockchain.BlockHeader{
			PreviousHash: []byte("GENESIS"),
			Timestamp:    time.Now().UnixNano(),
		},
		Transactions: nil,
	}
	genesisBlock.ComputeMerkleRoot()
	genesisBlock.Header.Hash = []byte("GENESIS")
	if err := chain.AddBlock(genesisBlock); err != nil {
		t.Fatalf("Failed to add genesis block: %v", err)
	}

	miner := NewMiner(mempool, chain, 5, []string{}, "00")
	ipfsClient := ipfs.NewMockIPFSClient(true)
	miner.SetIPFSClient(ipfsClient)

	// Step 2: Add transactions to the mempool
	tx1 := blockchain.Transaction{TxID: []byte("tx1"), DataHash: "dataHash1", AlgorithmHash: "algoHash1"}
	tx2 := blockchain.Transaction{TxID: []byte("tx2"), DataHash: "dataHash2", AlgorithmHash: "algoHash2"}
	mempool.AddTransaction(tx1)
	mempool.AddTransaction(tx2)

	t.Logf("Miner initialized: %+v", miner)
	t.Logf("Mempool transactions: %v", len(mempool.Transactions))

	// Step 3: Mine a block
	block := miner.MineBlock()
	if block == nil {
		t.Fatalf("Failed to mine a block")
	}
	t.Logf("Block mined: %+v", block)

	// Step 6: Validate the PreviousHash
	if len(chain.Blocks) < 2 {
		t.Fatalf("Chain has less than 2 blocks; cannot validate PreviousHash")
	}

	// Step 4: Validate PreviousHash
	lastBlock := chain.Blocks[len(chain.Blocks)-2] // Explicitly refer to the previous block
	t.Logf("Chain Tip Hash: %x", lastBlock.Header.Hash)

	t.Logf("Block PreviousHash: %x", block.Header.PreviousHash)

	if !bytes.Equal(block.Header.PreviousHash, lastBlock.Header.Hash) {
		t.Fatalf("PreviousHash mismatch. Expected: %x, Got: %x",
			lastBlock.Header.Hash, block.Header.PreviousHash)
	}

	t.Logf("PreviousHash validation successful")

	// Step 6: Validate Proof of Work
	if !pow.ValidateProofOfWork(block.Header.Bytes(), block.Header.Nonce, miner.DifficultyTarget) {
		t.Errorf("Block's PoW is invalid. Nonce: %d, Hash: %x", block.Header.Nonce, block.Header.Hash)
	} else {
		t.Logf("Block's PoW validated successfully. Nonce: %d, Hash: %x", block.Header.Nonce, block.Header.Hash)
	}

	//	Step 7: Validate that mempool is cleared
	if len(mempool.GetAllTransactions()) != 0 {
		t.Errorf("Mempool is not cleared after mining")
	} else {
		t.Logf("Mempool cleared successfully")
	}
}
