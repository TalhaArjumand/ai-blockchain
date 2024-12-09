package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type Blockchain struct {
	Blocks map[int]*Block // Height -> Block
	Mutex  sync.Mutex     // For thread-safe access
}

// Create a new blockchain
func NewBlockchain() *Blockchain {
	return &Blockchain{
		Blocks: make(map[int]*Block),
		Mutex:  sync.Mutex{},
	}
}

// Add a block to the blockchain
func (bc *Blockchain) AddBlock(block *Block) error {
	bc.Mutex.Lock()
	defer bc.Mutex.Unlock()

	// Check for duplicate blocks based on Merkle Root
	for _, existingBlock := range bc.Blocks {
		if bytes.Equal(existingBlock.Header.MerkleRoot, block.Header.MerkleRoot) {
			return fmt.Errorf("duplicate block with Merkle Root %x", block.Header.MerkleRoot)
		}
	}

	// Set the previous hash for the new block
	height := len(bc.Blocks)
	if height == 0 {
		block.Header.PreviousHash = []byte("GENESIS")
	} else {
		lastBlock := bc.Blocks[height-1]
		block.Header.PreviousHash = lastBlock.Header.MerkleRoot
	}

	// Set the timestamp for the new block
	block.Header.Timestamp = time.Now().UnixNano()

	// Add the block to the blockchain
	bc.Blocks[height] = block
	return nil
}

// Persist the blockchain to disk
func (bc *Blockchain) Persist() error {
	file, err := os.Create("blockchain.json")
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(bc.Blocks)
}

// Load the blockchain from disk
func (bc *Blockchain) Load() error {
	file, err := os.Open("blockchain.json")
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(&bc.Blocks)
}

func FetchBlocks(startHeight, endHeight int, chain *Blockchain) ([]Block, error) {
	chain.Mutex.Lock()
	defer chain.Mutex.Unlock()

	if startHeight > endHeight {
		return nil, fmt.Errorf("startHeight cannot be greater than endHeight")
	}

	var blocks []Block
	for height := startHeight; height <= endHeight; height++ {
		block, exists := chain.Blocks[height]
		if !exists {
			return nil, fmt.Errorf("block at height %d not found", height)
		}
		blocks = append(blocks, *block)
	}
	return blocks, nil
}

func (bc *Blockchain) HasBlock(hash string) bool {
	bc.Mutex.Lock()
	defer bc.Mutex.Unlock()

	for _, block := range bc.Blocks {
		if fmt.Sprintf("%x", block.Header.MerkleRoot) == hash {
			return true
		}
	}
	return false
}

func (bc *Blockchain) GetBlock(hash string) *Block {
	for _, block := range bc.Blocks {
		if string(block.Header.MerkleRoot) == hash {
			return block
		}
	}
	return nil
}

// GetBlockByHeight fetches a block by its height.
func (bc *Blockchain) GetBlockByHeight(height int) *Block {
	block, exists := bc.Blocks[height]
	if !exists {
		return nil // Return nil if the block does not exist
	}
	return block
}
