package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type Blockchain struct {
	Blocks       map[int]*Block    // Height -> Block
	ByHash       map[string]*Block // Hash string -> *Block
	OrphanBlocks map[string]*Block // Hash string -> *Block (blocks that don't yet fit into a longer chain)
	Mutex        sync.Mutex        // For thread-safe access
}

// Reset clears all blocks in the blockchain.
func (bc *Blockchain) Reset() {
	bc.Mutex.Lock()
	defer bc.Mutex.Unlock()
	bc.Blocks = map[int]*Block{}
	bc.ByHash = map[string]*Block{}
	bc.OrphanBlocks = map[string]*Block{}
}

// Create a new blockchain
func NewBlockchain() *Blockchain {
	return &Blockchain{
		Blocks:       make(map[int]*Block),
		ByHash:       make(map[string]*Block),
		OrphanBlocks: make(map[string]*Block),
		Mutex:        sync.Mutex{},
	}
}

// AddBlock attempts to add a block to the blockchain or orphan storage.
func (bc *Blockchain) AddBlock(block *Block) error {
	bc.Mutex.Lock()
	defer bc.Mutex.Unlock()

	// Check for duplicates
	for _, existingBlock := range bc.Blocks {
		if bytes.Equal(existingBlock.Header.Hash, block.Header.Hash) {
			return fmt.Errorf("duplicate block with Hash %x", block.Header.Hash)
		}
	}
	for _, orphan := range bc.OrphanBlocks {
		if bytes.Equal(orphan.Header.Hash, block.Header.Hash) {
			return fmt.Errorf("duplicate orphan block with Hash %x", block.Header.Hash)
		}
	}

	height := len(bc.Blocks)
	if height == 0 {
		block.Header.PreviousHash = []byte("GENESIS") // Ensure a clear distinction for the genesis block
		block.Header.Timestamp = time.Now().UnixNano()
		bc.Blocks[0] = block
		bc.ByHash[string(block.Header.Hash)] = block
		bc.processOrphans() // Re-check orphans after adding the genesis block
		return nil
	}

	// Try to attach to tip
	lastBlock := bc.Blocks[height-1]
	if bytes.Equal(block.Header.PreviousHash, lastBlock.Header.Hash) {
		// Attach to tip normally
		block.Header.Timestamp = time.Now().UnixNano()
		bc.Blocks[height] = block
		bc.ByHash[string(block.Header.Hash)] = block
		// After adding, re-check orphans
		bc.processOrphans()
		return nil
	}

	// Not attaching to the tip, attempt longest-chain logic
	newChain, err := bc.tryFormChain(block)
	if err != nil {
		// Discard block if its ancestor is unknown
		if bc.findBlockByHash(block.Header.PreviousHash) == nil {
			log.Printf("Discarding block %x with unknown ancestor %x", block.Header.Hash, block.Header.PreviousHash)
			return fmt.Errorf("unknown ancestor block with hash %x", block.Header.PreviousHash)
		}

		// Store as orphan if it references a known block
		log.Printf("Storing block %x as orphan: %v", block.Header.Hash, err)
		bc.OrphanBlocks[string(block.Header.Hash)] = block
		return nil
	}

	// If we formed a valid chain, check length
	if len(newChain) > len(bc.Blocks) {
		log.Printf("Reorganizing chain with new longer chain. New length: %d", len(newChain))

		// Longer chain found, reorganize
		bc.reorganizeChain(newChain)
		log.Println("Blockchain reorganized to a longer fork.")
		log.Printf("Blockchain state after reorg: Blocks %d", len(bc.Blocks))

		// After reorganizing, re-check orphans
		bc.processOrphans()
		return nil
	}

	// It's a valid chain but not longer - store as orphan for future
	log.Printf("Valid fork found but not longer. Storing %x as orphan.", block.Header.Hash)
	bc.OrphanBlocks[string(block.Header.Hash)] = block

	// After storing, we can also try to connect other orphans
	bc.processOrphans()
	return nil
}

// processOrphans tries to connect orphan blocks to the main chain if possible.
// It attempts to build chains from orphans and see if they now form a longer chain.
func (bc *Blockchain) processOrphans() {
	for {
		progressMade := false
		for hash, orphan := range bc.OrphanBlocks {
			// Attempt to form a chain with the orphan
			log.Printf("Processing orphan block: Hash %x, PreviousHash %x", orphan.Header.Hash, orphan.Header.PreviousHash)
			newChain, err := bc.tryFormChain(orphan)
			if err != nil {
				log.Printf("Failed to form chain with orphan %x: %v", orphan.Header.Hash, err)
				// Still can't form a chain, continue
				continue
			}
			// Check if the new chain is longer
			if len(newChain) > len(bc.Blocks) {
				// Reorganize to the longer chain
				bc.reorganizeChain(newChain)
				delete(bc.OrphanBlocks, hash)
				log.Printf("Reorganized chain using orphan block %x", orphan.Header.Hash)
				progressMade = true
			} else {
				log.Printf(" orphan doesn't form a longer chain, it remains in the orphan pool for future re-checks")
				// If the orphan doesn't form a longer chain, it remains
				// in the orphan pool for future re-checks
			}
		}

		if !progressMade {
			break
		}
	}
}

// tryFormChain attempts to walk back from the given block's PreviousHash to genesis.
// If successful, returns the entire chain of blocks from genesis to the given block.
func (bc *Blockchain) tryFormChain(block *Block) ([]*Block, error) {
	chain := []*Block{block}
	current := block

	for {
		if bytes.Equal(current.Header.PreviousHash, []byte("GENESIS")) {
			// Reached genesis
			break
		}

		prevBlock := bc.findBlockByHash(current.Header.PreviousHash)
		if prevBlock == nil {
			return nil, fmt.Errorf("unknown ancestor block with hash %x", current.Header.PreviousHash)
		}
		chain = append(chain, prevBlock)
		current = prevBlock
	}

	// Reverse chain to start from genesis
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}

	return chain, nil
}

// findBlockByHash searches both the main chain and orphan blocks for a block by hash.
func (bc *Blockchain) findBlockByHash(hash []byte) *Block {
	if b, ok := bc.ByHash[string(hash)]; ok {
		return b
	}
	if b, ok := bc.OrphanBlocks[string(hash)]; ok {
		return b
	}
	return nil
}

// reorganizeChain replaces the current chain with a new one.
// Assumes the new chain starts from genesis and is longer.
func (bc *Blockchain) reorganizeChain(newChain []*Block) {
	bc.Blocks = make(map[int]*Block)
	bc.ByHash = make(map[string]*Block)

	for i, blk := range newChain {
		bc.Blocks[i] = blk
		bc.ByHash[string(blk.Header.Hash)] = blk
	}

	// After reorg, some orphans might now be invalid or irrelevant, but we keep them
	// in orphan storage. They might form a different fork in the future. Or we could
	// prune orphan blocks that no longer connect to anything. For simplicity, we leave
	// them as is, as they won't attach without a known ancestor.
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
	err = decoder.Decode(&bc.Blocks)
	if err != nil {
		return err
	}

	// Rebuild ByHash from Blocks
	bc.ByHash = make(map[string]*Block)
	for i, blk := range bc.Blocks {
		bc.ByHash[string(blk.Header.Hash)] = bc.Blocks[i]
	}
	return nil
}

// Additional utility methods remain unchanged
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

	_, exists := bc.ByHash[hash]
	return exists
}

func (bc *Blockchain) GetBlock(hash string) *Block {
	bc.Mutex.Lock()
	defer bc.Mutex.Unlock()

	return bc.ByHash[hash]
}

// GetBlockByHeight fetches a block by its height.
func (bc *Blockchain) GetBlockByHeight(height int) *Block {
	bc.Mutex.Lock()
	defer bc.Mutex.Unlock()

	block, exists := bc.Blocks[height]
	if !exists {
		return nil
	}
	return block
}
