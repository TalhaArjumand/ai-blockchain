package miner

import (
	"log"
	"sync"
	"time"

	"github.com/TalhaArjumand/ai-blockchain/pkg/blockchain"
	"github.com/TalhaArjumand/ai-blockchain/pkg/config"
	"github.com/TalhaArjumand/ai-blockchain/pkg/ipfs"
	"github.com/TalhaArjumand/ai-blockchain/pkg/network"
	"github.com/TalhaArjumand/ai-blockchain/pkg/pow"
	"github.com/TalhaArjumand/ai-blockchain/pkg/vm"
)

type Miner struct {
	Mempool              *blockchain.Mempool
	Blockchain           *blockchain.Blockchain
	MaxBlockTransactions int // Add this field
	Peers                []string
	Mutex                sync.Mutex
	Broadcaster          Broadcaster
	IPFSClient           ipfs.IPFSInterface
	Config               *config.Config // Add this field to hold configuration
	DifficultyTarget     string         // New field for the difficulty target
}

type Broadcaster interface {
	BroadcastBlock(block network.BlockMessage, peers []string)
}

// SetBroadcaster sets a custom broadcaster for the miner.
func (m *Miner) SetBroadcaster(b Broadcaster) {
	m.Broadcaster = b
}

type DefaultBroadcaster struct{}

func (b *DefaultBroadcaster) BroadcastBlock(block network.BlockMessage, peers []string) {
	network.BroadcastBlock(block, peers)
}

// NewMiner initializes the miner
// NewMiner initializes the miner
func NewMiner(mempool *blockchain.Mempool, blockchain *blockchain.Blockchain, maxTx int, peers []string, difficultyTarget string) *Miner {
	return &Miner{
		Mempool:              mempool,
		Blockchain:           blockchain,
		MaxBlockTransactions: maxTx,
		Peers:                peers,
		Broadcaster:          &DefaultBroadcaster{},
		DifficultyTarget:     difficultyTarget, // Use the passed difficulty target
	}
}

func (miner *Miner) InitializeBlockchain() {
	if len(miner.Blockchain.Blocks) == 0 {
		genesisBlock := &blockchain.Block{
			Header: blockchain.BlockHeader{
				PreviousHash: []byte("GENESIS"),
				Timestamp:    time.Now().UnixNano(),
			},
			Transactions: nil, // No transactions in the genesis block
		}
		genesisBlock.ComputeMerkleRoot()

		err := miner.Blockchain.AddBlock(genesisBlock)
		if err != nil {
			log.Fatalf("Failed to add genesis block: %v", err)
		}
		log.Println("Genesis block added successfully")
	}
}

func (miner *Miner) MineBlock() *blockchain.Block {
	// Check if the mempool is empty
	if len(miner.Mempool.Transactions) == 0 {
		log.Println("No transactions in the mempool, skipping mining")
		return nil
	}

	// Check if the blockchain is empty and create a genesis block if needed
	if len(miner.Blockchain.Blocks) == 0 {
		genesisBlock := &blockchain.Block{
			Header: blockchain.BlockHeader{
				PreviousHash: []byte("GENESIS"),
				Timestamp:    time.Now().UnixNano(),
			},
			Transactions: nil, // Genesis block has no transactions
		}
		genesisBlock.ComputeMerkleRoot()

		// Add the genesis block to the blockchain
		err := miner.Blockchain.AddBlock(genesisBlock)
		if err != nil {
			log.Printf("Failed to add genesis block: %v", err)
			return nil
		}
		log.Println("Genesis block mined successfully")
		// Do not return here; continue to mine additional blocks
	}

	// Pick transactions from the mempool
	transactions := miner.PickTransactions()

	if len(transactions) == 0 {
		log.Println("No transactions to mine, skipping")
		return nil
	}

	// Process transactions for VM execution
	for i, tx := range transactions {
		if tx.DataHash == "" || tx.AlgorithmHash == "" {
			log.Printf("Transaction %x has incomplete fields, skipping\n", tx.TxID)
			continue
		}
		data, err := miner.IPFSClient.FetchData(tx.DataHash)
		if err != nil {
			log.Printf("Failed to fetch data for Tx %x, skipping: %v\n", tx.TxID, err)
			continue
		}
		algo, err := miner.IPFSClient.FetchAlgorithm(tx.AlgorithmHash)
		if err != nil {
			log.Printf("Failed to fetch algorithm for Tx %x, skipping: %v\n", tx.TxID, err)
			continue
		}
		vmOutput, err := vm.RunVM(algo, data)
		if err != nil {
			log.Printf("RunVM failed for Tx %x: %v\n", tx.TxID, err)
			continue
		}
		transactions[i].VMOutput = vmOutput
	}

	// Create a new block with the processed transactions
	previousHash := miner.Blockchain.Blocks[len(miner.Blockchain.Blocks)-1].Header.MerkleRoot
	block := &blockchain.Block{
		Header: blockchain.BlockHeader{
			PreviousHash: previousHash,
			Timestamp:    time.Now().UnixNano(),
			Nonce:        0, // Initialize nonce
		},
		Transactions: transactions,
	}

	// Compute Merkle root and other headers
	block.ComputeMerkleRoot()
	block.ComputeVMOutputsHash()

	// Ensure the block has a unique Merkle Root
	if miner.Blockchain.HasDuplicateMerkleRoot(block.Header.MerkleRoot) {
		log.Printf("Duplicate Merkle Root detected: %x, regenerating", block.Header.MerkleRoot)
		block.Header.Timestamp = time.Now().UnixNano()
		block.ComputeMerkleRoot()
	}

	log.Printf("Checking if Mempool is nil: %v", miner.Mempool == nil)
	log.Printf("Checking if Blockchain is nil: %v", miner.Blockchain == nil)
	log.Printf("Checking if Transactions are nil: %v", miner.Mempool.Transactions == nil)

	if miner.DifficultyTarget == "" {
		log.Fatalf("Miner.DifficultyTarget is nil; ensure it is initialized before mining.")
	}

	log.Printf("Before Mining Header: %x", block.Header.Bytes())

	nonce, hash := pow.PerformProofOfWork(block.Header.Bytes(), miner.DifficultyTarget)
	block.Header.Nonce = uint64(nonce)
	block.Header.Hash = []byte(hash)
	log.Printf("Mining Difficulty Target: %s", miner.DifficultyTarget)

	// Add the block to the blockchain
	err := miner.Blockchain.AddBlock(block)
	if err != nil {
		log.Printf("Failed to add mined block: %v", err)
		return nil
	}

	// Log success
	log.Printf("Block mined successfully with Merkle Root: %x", block.Header.MerkleRoot)

	// Broadcast the block to all peers
	log.Printf("Broadcasting the mined block to peers...")
	miner.BroadcastBlock(block)

	// Remove mined transactions from mempool
	for _, tx := range block.Transactions {
		miner.Mempool.RemoveTransaction(string(tx.TxID))
	}

	return block
}

func (miner *Miner) SetIPFSClient(client ipfs.IPFSInterface) {
	miner.IPFSClient = client
}

// pickTransactions selects transactions from the mempool
func (miner *Miner) PickTransactions() []blockchain.Transaction {
	miner.Mempool.Mutex.Lock()
	defer miner.Mempool.Mutex.Unlock()

	var transactions []blockchain.Transaction
	count := 0
	for _, tx := range miner.Mempool.Transactions {
		if count >= miner.MaxBlockTransactions {
			break
		}
		transactions = append(transactions, tx)
		count++
	}
	return transactions
}

// broadcastBlock sends the block to all peers
func (miner *Miner) BroadcastBlock(block *blockchain.Block) {
	// Construct the network.BlockMessage
	blockMsg := network.BlockMessage{
		BlockID:      string(block.Header.MerkleRoot), // Use MerkleRoot as BlockID
		MerkleRoot:   string(block.Header.MerkleRoot),
		PreviousHash: string(block.Header.PreviousHash),
		Timestamp:    block.Header.Timestamp,
		Transactions: []network.TxMessage{},
	}

	// Populate Transactions in the BlockMessage
	for _, tx := range block.Transactions {
		blockMsg.Transactions = append(blockMsg.Transactions, network.TxMessage{
			Type:      "transaction",
			TxID:      string(tx.TxID),
			DataHash:  tx.DataHash,
			AlgoHash:  tx.AlgorithmHash,
			Metadata:  tx.Metadata,
			Timestamp: tx.Timestamp,
		})
	}

	// Ensure a valid broadcaster is set
	if miner.Broadcaster == nil {
		log.Println("Broadcaster is not set for miner. Unable to broadcast block.")
		return
	}

	// Call the broadcaster to send the block message
	miner.Broadcaster.BroadcastBlock(blockMsg, miner.Peers)
}
