package blockchain

import (
	"bytes"
	"testing"
)

func TestBasicFork(t *testing.T) {
	chain := NewBlockchain()

	// Step 1: Add Genesis Block
	genesisBlock := &Block{
		Header: BlockHeader{
			Hash:         []byte("GENESIS_TAG"), // Genesis block's own hash
			PreviousHash: []byte("GENESIS_TAG"), // Distinct marker for the genesis block
		},
	}
	if err := chain.AddBlock(genesisBlock); err != nil {
		t.Fatalf("Failed to add GenesisBlock: %v", err)
	}

	// Step 2: Add Block1 to the main chain
	block1 := &Block{
		Header: BlockHeader{
			Hash:         []byte("Block1"),
			PreviousHash: genesisBlock.Header.Hash,
		},
	}
	if err := chain.AddBlock(block1); err != nil {
		t.Fatalf("Failed to add Block1: %v", err)
	}

	// Step 3: Add ForkBlock1 (doesn't extend the main chain, so it should go to orphan pool)
	forkBlock1 := &Block{
		Header: BlockHeader{
			Hash:         []byte("ForkBlock1"),
			PreviousHash: genesisBlock.Header.Hash,
		},
	}
	if err := chain.AddBlock(forkBlock1); err != nil {
		t.Logf("ForkBlock1 stored as orphan: %v", err)
	}

	// Step 4: Add ForkBlock2 (makes the fork longer)
	forkBlock2 := &Block{
		Header: BlockHeader{
			Hash:         []byte("ForkBlock2"),
			PreviousHash: forkBlock1.Header.Hash,
		},
	}
	if err := chain.AddBlock(forkBlock2); err != nil {
		t.Fatalf("Failed to add ForkBlock2: %v", err)
	}

	for height, blk := range chain.Blocks {
		t.Logf("Block at height %d: Hash %x, PreviousHash %x", height, blk.Header.Hash, blk.Header.PreviousHash)
	}

	for hash, orphan := range chain.OrphanBlocks {
		t.Logf("Orphan Block: Hash %x, PreviousHash %x", hash, orphan.Header.PreviousHash)
	}

	// Step 5: Verify chain reorganization
	if chain.Blocks[0] != genesisBlock ||
		chain.Blocks[1] != forkBlock1 ||
		chain.Blocks[2] != forkBlock2 {
		t.Fatalf("Chain did not reorganize to the fork as expected")
	}

	t.Log("TestBasicFork passed: Chain successfully reorganized to longer fork")
}

func TestDisconnectedBlock(t *testing.T) {
	chain := NewBlockchain()

	// Step 1: Add Genesis Block
	genesisBlock := &Block{
		Header: BlockHeader{
			Hash:         []byte("GENESIS_TAG"),
			PreviousHash: []byte("GENESIS_TAG"),
		},
	}
	if err := chain.AddBlock(genesisBlock); err != nil {
		t.Fatalf("Failed to add genesis block: %v", err)
	}

	// Step 2: Add a disconnected block
	disconnectedBlock := &Block{
		Header: BlockHeader{
			Hash:         []byte("Disconnected"),
			PreviousHash: []byte("UnknownHash"),
		},
	}
	if err := chain.AddBlock(disconnectedBlock); err == nil {
		t.Fatalf("Disconnected block should not have been added")
	}
	t.Log("Disconnected Block Test Passed")
}

func TestRollbackAndReapply(t *testing.T) {
	chain := NewBlockchain()

	// Step 1: Add Genesis Block
	genesisBlock := &Block{
		Header: BlockHeader{
			Hash:         []byte("GENESIS_Tag"),
			PreviousHash: []byte("GENESIS"),
		},
	}
	if err := chain.AddBlock(genesisBlock); err != nil {
		t.Fatalf("Failed to add genesis block: %v", err)
	}

	// Step 2: Add blocks to the current chain
	block1 := &Block{
		Header: BlockHeader{
			Hash:         []byte("Block1"),
			PreviousHash: genesisBlock.Header.Hash,
		},
	}
	if err := chain.AddBlock(block1); err != nil {
		t.Fatalf("Failed to add Block1: %v", err)
	}

	block2 := &Block{
		Header: BlockHeader{
			Hash:         []byte("Block2"),
			PreviousHash: block1.Header.Hash,
		},
	}
	if err := chain.AddBlock(block2); err != nil {
		t.Fatalf("Failed to add Block2: %v", err)
	}

	// Step 3: Create a fork starting from genesis and make it longer
	forkBlock1 := &Block{
		Header: BlockHeader{
			Hash:         []byte("ForkBlock1"),
			PreviousHash: genesisBlock.Header.Hash,
		},
	}
	if err := chain.AddBlock(forkBlock1); err != nil {
		// Initial fork block may not switch the chain
		t.Logf("ForkBlock1 not added: %v", err)
	}

	forkBlock2 := &Block{
		Header: BlockHeader{
			Hash:         []byte("ForkBlock2"),
			PreviousHash: forkBlock1.Header.Hash,
		},
	}
	if err := chain.AddBlock(forkBlock2); err != nil {
		// ForkBlock2 may also not immediately apply
		t.Logf("ForkBlock2 not added: %v", err)
	}

	forkBlock3 := &Block{
		Header: BlockHeader{
			Hash:         []byte("ForkBlock3"),
			PreviousHash: forkBlock2.Header.Hash,
		},
	}
	if err := chain.AddBlock(forkBlock3); err != nil {
		t.Fatalf("Failed to add ForkBlock3: %v", err)
	}

	// Step 4: Validate that the chain has rolled back and applied the longer fork
	if chain.Blocks[1] != forkBlock1 || chain.Blocks[2] != forkBlock2 || chain.Blocks[3] != forkBlock3 {
		t.Fatalf("Blockchain did not reorganize to the longer fork")
	}
	t.Log("Rollback and Reapply Test Passed")
}

func TestMultipleCompetingForks(t *testing.T) {
	chain := NewBlockchain()

	// Step 1: Add Genesis Block
	genesisBlock := &Block{
		Header: BlockHeader{
			Hash:         []byte("GENESIS_TAG"),
			PreviousHash: []byte("GENESIS"),
		},
	}
	if err := chain.AddBlock(genesisBlock); err != nil {
		t.Fatalf("Failed to add GenesisBlock: %v", err)
	}

	// Step 2: Create first competing fork (2 blocks)
	fork1Block1 := &Block{
		Header: BlockHeader{
			Hash:         []byte("Fork1_Block1"),
			PreviousHash: genesisBlock.Header.Hash,
		},
	}
	if err := chain.AddBlock(fork1Block1); err != nil {
		t.Logf("Fork1_Block1 not added: %v", err)
	}

	fork1Block2 := &Block{
		Header: BlockHeader{
			Hash:         []byte("Fork1_Block2"),
			PreviousHash: fork1Block1.Header.Hash,
		},
	}
	if err := chain.AddBlock(fork1Block2); err != nil {
		t.Logf("Fork1_Block2 not added: %v", err)
	}

	// Step 3: Create second competing fork (3 blocks)
	fork2Block1 := &Block{
		Header: BlockHeader{
			Hash:         []byte("Fork2_Block1"),
			PreviousHash: genesisBlock.Header.Hash,
		},
	}
	if err := chain.AddBlock(fork2Block1); err != nil {
		t.Logf("Fork2_Block1 not added: %v", err)
	}

	fork2Block2 := &Block{
		Header: BlockHeader{
			Hash:         []byte("Fork2_Block2"),
			PreviousHash: fork2Block1.Header.Hash,
		},
	}
	if err := chain.AddBlock(fork2Block2); err != nil {
		t.Logf("Fork2_Block2 not added: %v", err)
	}

	fork2Block3 := &Block{
		Header: BlockHeader{
			Hash:         []byte("Fork2_Block3"),
			PreviousHash: fork2Block2.Header.Hash,
		},
	}
	if err := chain.AddBlock(fork2Block3); err != nil {
		t.Logf("Fork2_Block3 not added: %v", err)
	}

	// Step 4: Validate blockchain has adopted the longest fork
	height := len(chain.Blocks)
	if height != 4 {
		t.Fatalf("Blockchain did not adopt the longest fork. Current height: %d", height)
	}

	// Step 5: Validate blocks in the chain
	expectedHashes := []string{"GENESIS_TAG", "Fork2_Block1", "Fork2_Block2", "Fork2_Block3"}
	for i, expected := range expectedHashes {
		if !bytes.Equal(chain.Blocks[i].Header.Hash, []byte(expected)) {
			t.Fatalf("Block at height %d has unexpected hash. Expected: %s, Got: %x",
				i, expected, chain.Blocks[i].Header.Hash)
		}
	}

	t.Log("TestMultipleCompetingForks passed: Blockchain correctly adopted the longest fork")
}
