package blockchain

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"sync"
	"testing"
)

// TestNewBlockchain tests the creation of a new blockchain
func TestNewBlockchain(t *testing.T) {
	bc := NewBlockchain()

	if len(bc.Blocks) != 0 {
		t.Errorf("Expected no blocks in a new blockchain, got %d blocks", len(bc.Blocks))
	}
}

// TestAddBlock tests adding a block to the blockchain
func TestAddBlock(t *testing.T) {
	bc := NewBlockchain()
	block := &Block{
		Header: BlockHeader{
			PreviousHash: []byte("prevHash"),
		},
	}
	bc.AddBlock(block)

	if len(bc.Blocks) != 1 {
		t.Errorf("Expected 1 block in the blockchain, got %d", len(bc.Blocks))
	}

	if !reflect.DeepEqual(bc.Blocks[0], block) {
		t.Errorf("Added block does not match the expected block")
	}
}

// TestAddMultipleBlocks tests adding multiple blocks to the blockchain

func TestAddMultipleBlocks(t *testing.T) {
	bc := NewBlockchain()

	for i := 0; i < 5; i++ {
		block := &Block{
			Header: BlockHeader{
				MerkleRoot: []byte(fmt.Sprintf("merkleRoot%d", i)),
			},
		}
		err := bc.AddBlock(block)
		if err != nil {
			t.Fatalf("Failed to add block %d: %v", i, err)
		}
	}

	if len(bc.Blocks) != 5 {
		t.Errorf("Expected 5 blocks in the blockchain, got %d", len(bc.Blocks))
	}
}

// TestPersist tests the persistence of the blockchain to disk
func TestPersist(t *testing.T) {
	bc := NewBlockchain()
	block := &Block{
		Header: BlockHeader{
			PreviousHash: []byte("prevHash"),
		},
	}
	bc.AddBlock(block)

	err := bc.Persist()
	if err != nil {
		t.Fatalf("Failed to persist blockchain: %v", err)
	}

	// Check if the file was created
	if _, err := os.Stat("blockchain.json"); os.IsNotExist(err) {
		t.Errorf("Expected blockchain.json file to exist, but it does not")
	}

	// Clean up
	os.Remove("blockchain.json")
}

// TestLoad tests loading the blockchain from disk
func TestLoad(t *testing.T) {
	// Setup a blockchain and persist it
	bc := NewBlockchain()
	block := &Block{
		Header: BlockHeader{
			PreviousHash: []byte("prevHash"),
		},
	}
	bc.AddBlock(block)
	bc.Persist()

	// Create a new blockchain instance and load the data
	bcLoaded := NewBlockchain()
	err := bcLoaded.Load()
	if err != nil {
		t.Fatalf("Failed to load blockchain: %v", err)
	}

	// Check if the loaded blockchain matches the persisted one
	if len(bcLoaded.Blocks) != len(bc.Blocks) {
		t.Errorf("Expected %d blocks after loading, got %d", len(bc.Blocks), len(bcLoaded.Blocks))
	}

	if !reflect.DeepEqual(bcLoaded.Blocks, bc.Blocks) {
		t.Errorf("Loaded blockchain does not match the persisted blockchain")
	}

	// Clean up
	os.Remove("blockchain.json")
}

// TestLoadNonExistentFile tests loading a blockchain from a non-existent file
func TestLoadNonExistentFile(t *testing.T) {
	bc := NewBlockchain()
	err := bc.Load()

	if err == nil {
		t.Errorf("Expected an error when loading from a non-existent file, got nil")
	}
}

// TestConcurrency tests thread-safe access to the blockchain
func TestConcurrency(t *testing.T) {
	bc := NewBlockchain()
	numBlocks := 100
	var wg sync.WaitGroup

	for i := 0; i < numBlocks; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			block := &Block{
				Header: BlockHeader{
					MerkleRoot: []byte(fmt.Sprintf("merkleRoot%d", i)),
				},
			}
			_ = bc.AddBlock(block) // Add block without causing errors
		}(i)
	}

	wg.Wait()

	if len(bc.Blocks) != numBlocks {
		t.Errorf("Expected %d blocks after concurrent adds, got %d", numBlocks, len(bc.Blocks))
	}
}

// TestGenesisBlockValidation ensures the first block uses "GENESIS" as the previous hash
func TestGenesisBlockValidation(t *testing.T) {
	bc := NewBlockchain()
	block := &Block{}
	bc.AddBlock(block)

	if !bytes.Equal(bc.Blocks[0].Header.PreviousHash, []byte("GENESIS")) {
		t.Errorf("Expected Genesis block to have 'GENESIS' as the previous hash, got %s", bc.Blocks[0].Header.PreviousHash)
	}
}

// TestDuplicateBlockAddition tests adding the same block multiple times
func TestDuplicateBlockAddition(t *testing.T) {
	bc := NewBlockchain()

	// Create the first block
	block1 := &Block{
		Header: BlockHeader{
			MerkleRoot: []byte("merkle123"),
		},
	}

	// Add the first block
	err := bc.AddBlock(block1)
	if err != nil {
		t.Fatalf("Failed to add block: %v", err)
	}

	// Create a second block with the same MerkleRoot
	block2 := &Block{
		Header: BlockHeader{
			MerkleRoot: []byte("merkle123"),
		},
	}

	// Attempt to add the duplicate block
	err = bc.AddBlock(block2)
	if err == nil {
		t.Errorf("Expected error when adding duplicate block, but got none")
	}

	// Ensure the blockchain contains only one block
	if len(bc.Blocks) != 1 {
		t.Errorf("Expected 1 block after adding duplicate blocks, got %d", len(bc.Blocks))
	}
}

// TestInvalidBlockData tests adding a block with invalid fields
func TestInvalidBlockData(t *testing.T) {
	bc := NewBlockchain()
	block := &Block{} // Missing fields like PreviousHash

	bc.AddBlock(block)
	if len(bc.Blocks) != 1 {
		t.Errorf("Expected block with invalid data to still be added")
	}
}

// TestFetchBlocksEdgeCases tests fetching blocks with edge cases
func TestFetchBlocksEdgeCases(t *testing.T) {
	bc := NewBlockchain()

	for i := 0; i < 5; i++ {
		block := &Block{}
		bc.AddBlock(block)
	}

	_, err := FetchBlocks(6, 7, bc)
	if err == nil {
		t.Errorf("Expected error when fetching non-existent blocks, got nil")
	}

	_, err = FetchBlocks(3, 2, bc)
	if err == nil {
		t.Errorf("Expected error for invalid range (startHeight > endHeight), got nil")
	}
}

// TestBoundaryConditions ensures correct PreviousHash for consecutive blocks
func TestBoundaryConditions(t *testing.T) {
	bc := NewBlockchain()

	block1 := &Block{
		Header: BlockHeader{
			PreviousHash: []byte("prevHash1"),
		},
	}
	block2 := &Block{}

	bc.AddBlock(block1)
	bc.AddBlock(block2)

	if !bytes.Equal(block2.Header.PreviousHash, block1.Header.MerkleRoot) {
		t.Errorf("Expected second block's PreviousHash to match first block's MerkleRoot")
	}
}

// TestPersistConsistency ensures blockchain persistence does not save unintended changes
func TestPersistConsistency(t *testing.T) {
	bc := NewBlockchain()
	block := &Block{
		Header: BlockHeader{
			PreviousHash: []byte("prevHash"),
		},
	}
	bc.AddBlock(block)
	bc.Persist()

	// Modify the blockchain in memory
	bc.Blocks[0].Header.PreviousHash = []byte("modifiedHash")

	// Load from disk and check for consistency
	bcLoaded := NewBlockchain()
	bcLoaded.Load()

	if reflect.DeepEqual(bc.Blocks, bcLoaded.Blocks) {
		t.Errorf("Loaded blockchain should not reflect in-memory changes")
	}

	os.Remove("blockchain.json") // Clean up
}

// TestConcurrencyForFetchBlocks ensures thread-safe FetchBlocks during concurrent block additions
func TestConcurrencyForFetchBlocks(t *testing.T) {
	bc := NewBlockchain()

	// Add blocks to the blockchain
	for i := 0; i < 10; i++ {
		block := &Block{
			Header: BlockHeader{
				MerkleRoot: []byte(fmt.Sprintf("merkleRoot%d", i)),
			},
		}
		_ = bc.AddBlock(block)
	}

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			blocks, err := FetchBlocks(start, end, bc)
			if err != nil {
				t.Errorf("Failed to fetch blocks concurrently: %v", err)
				return
			}
			if len(blocks) != end-start+1 {
				t.Errorf("Expected %d blocks, got %d", end-start+1, len(blocks))
			}
		}(0, 9) // Fetch full range
	}
	wg.Wait()
}

// TestHasBlock checks for existence of blocks by hash
func TestHasBlock(t *testing.T) {
	bc := NewBlockchain()
	block := &Block{
		Header: BlockHeader{
			MerkleRoot: []byte("hash123"),
		},
	}
	bc.AddBlock(block)

	if !bc.HasBlock(fmt.Sprintf("%x", block.Header.MerkleRoot)) {
		t.Errorf("Expected blockchain to have block with hash '%x'", block.Header.MerkleRoot)
	}
}

// TestGetBlock ensures correct retrieval of blocks by hash
func TestGetBlock(t *testing.T) {
	bc := NewBlockchain()
	block := &Block{
		Header: BlockHeader{
			MerkleRoot: []byte("hash123"),
		},
	}
	bc.AddBlock(block)

	retrieved := bc.GetBlock("hash123")
	if retrieved == nil || !reflect.DeepEqual(retrieved, block) {
		t.Errorf("Expected to retrieve block with hash 'hash123'")
	}

	retrieved = bc.GetBlock("nonexistent")
	if retrieved != nil {
		t.Errorf("Expected to retrieve nil for nonexistent block hash")
	}
}

// TestCorruptBlockchainFile tests loading a blockchain from a corrupted file
func TestCorruptBlockchainFile(t *testing.T) {
	// Write invalid JSON to the file
	_ = os.WriteFile("blockchain.json", []byte("{invalid json"), 0644)
	defer os.Remove("blockchain.json") // Clean up

	bc := NewBlockchain()
	err := bc.Load()

	if err == nil {
		t.Errorf("Expected error when loading from a corrupted file, got nil")
	}
}
