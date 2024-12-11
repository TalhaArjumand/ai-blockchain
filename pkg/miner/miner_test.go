package miner_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/TalhaArjumand/ai-blockchain/pkg/blockchain"
	"github.com/TalhaArjumand/ai-blockchain/pkg/config"
	"github.com/TalhaArjumand/ai-blockchain/pkg/ipfs"
	"github.com/TalhaArjumand/ai-blockchain/pkg/miner"
	"github.com/TalhaArjumand/ai-blockchain/pkg/network"
	"github.com/TalhaArjumand/ai-blockchain/pkg/pow"
	"github.com/TalhaArjumand/ai-blockchain/pkg/vm"
)

// StartMockPeer creates a mock peer listener on the specified port
func StartMockPeer(port int, shutdown chan bool, wg *sync.WaitGroup) {
	defer wg.Done()
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		log.Fatalf("Failed to start mock peer on port %d: %v", port, err)
	}
	defer listener.Close()

	log.Printf("Mock peer listening on port %d", port)

	go func() {
		<-shutdown
		listener.Close() // Close listener when shutdown signal is received
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Mock peer on port %d shutting down: %v", port, err)
			return
		}
		go func(c net.Conn) {
			defer c.Close()
			buf := make([]byte, 1024)
			n, _ := c.Read(buf)
			log.Printf("Mock peer on port %d received: %s", port, string(buf[:n]))
		}(conn)
	}
}

type MockBroadcaster struct {
	BroadcastLogs []string // Capture logs for validation
}

func (b *MockBroadcaster) BroadcastBlock(block network.BlockMessage, peers []string) {
	var mu sync.Mutex // Protect access to BroadcastLogs

	for _, peer := range peers {
		conn, err := net.Dial("tcp", peer)
		if err != nil {
			log.Printf("Failed to send block to peer %s: %v", peer, err)
			continue
		}

		// Construct block message
		blockMessage := fmt.Sprintf(
			`{"block_id":"%s","merkle_root":"%s"}`,
			block.BlockID,
			block.MerkleRoot,
		)

		_, err = fmt.Fprintln(conn, blockMessage)
		conn.Close()

		if err != nil {
			log.Printf("Error sending block to peer %s: %v", peer, err)
			continue
		}

		// Log success and append to BroadcastLogs
		log.Printf("Block sent to peer %s", peer)
		mu.Lock()
		b.BroadcastLogs = append(b.BroadcastLogs, fmt.Sprintf(
			"Peer: %s, Block ID: %s, Merkle Root: %s",
			peer, block.BlockID, block.MerkleRoot,
		))
		mu.Unlock()
	}
}

func TestBroadcastBlock_InvalidPeers(t *testing.T) {
	mempool := blockchain.NewMempool()
	chain := blockchain.NewBlockchain()

	mockBroadcaster := &MockBroadcaster{BroadcastLogs: []string{}}
	peers := []string{"127.0.0.1:9999", "invalid_peer", "127.0.0.1:6001"}

	// Start a valid mock peer
	shutdown := make(chan bool)
	var wg sync.WaitGroup
	wg.Add(1)
	go StartMockPeer(6001, shutdown, &wg)

	// Allow mock peer to initialize
	time.Sleep(500 * time.Millisecond)

	miner := miner.NewMiner(mempool, chain, 5, peers, "0000")
	miner.SetBroadcaster(mockBroadcaster)

	block := &blockchain.Block{
		Header: blockchain.BlockHeader{
			MerkleRoot: []byte("merkleRoot"),
		},
	}

	miner.BroadcastBlock(block)

	// Shutdown mock peer
	close(shutdown)
	wg.Wait()

	// Validate BroadcastLogs
	t.Logf("Broadcast logs: %v", mockBroadcaster.BroadcastLogs)
	if len(mockBroadcaster.BroadcastLogs) != 1 {
		t.Errorf("Expected 1 valid broadcast, got %d", len(mockBroadcaster.BroadcastLogs))
		return
	}

	expectedLog := "Peer: 127.0.0.1:6001, Block ID: merkleRoot, Merkle Root: merkleRoot"
	if mockBroadcaster.BroadcastLogs[0] != expectedLog {
		t.Errorf("Expected log: %s, but got: %s", expectedLog, mockBroadcaster.BroadcastLogs[0])
	}
}

func TestMineBlock(t *testing.T) {
	mempool := blockchain.NewMempool()
	chain := blockchain.NewBlockchain()

	// Add mock transactions to the mempool
	for i := 0; i < 5; i++ {
		tx := blockchain.Transaction{
			DataHash:      fmt.Sprintf("test_data_hash_%d", i),
			AlgorithmHash: "test_algo_hash",
			Metadata:      "test_metadata",
		}
		tx.GenerateTxID()
		mempool.AddTransaction(tx)
	}

	// Mock active peers
	peers := []string{"127.0.0.1:6001", "127.0.0.1:6002"}
	var wg sync.WaitGroup
	shutdown := make(chan bool)
	wg.Add(2)
	go StartMockPeer(6001, shutdown, &wg)
	go StartMockPeer(6002, shutdown, &wg)

	// Allow time for mock peers to start
	time.Sleep(500 * time.Millisecond)

	// Initialize the miner with a mock IPFS client
	mockIPFSClient := &ipfs.MockIPFSClient{Valid: true}
	miner := miner.NewMiner(mempool, chain, 5, peers, "0000")
	miner.IPFSClient = mockIPFSClient

	// Ensure blockchain is initialized with a genesis block
	genesisBlock := &blockchain.Block{
		Header: blockchain.BlockHeader{
			PreviousHash: []byte("GENESIS"),
		},
	}
	err := chain.AddBlock(genesisBlock)
	if err != nil {
		t.Fatalf("Failed to add genesis block: %v", err)
	}

	// Mine a block
	block := miner.MineBlock()
	if block == nil {
		t.Fatal("Failed to mine a block")
	}

	if len(block.Transactions) != 5 {
		t.Errorf("Expected 5 transactions, got %d", len(block.Transactions))
	}

	// Validate the block was added to the blockchain
	chain.Mutex.Lock()
	if len(chain.Blocks) != 2 {
		t.Errorf("Expected 2 blocks in the blockchain, got %d", len(chain.Blocks))
	}
	chain.Mutex.Unlock()

	// Shutdown mock peers
	close(shutdown)
	wg.Wait()
}

func TestMinerInitialization(t *testing.T) {
	mempool := blockchain.NewMempool()
	chain := blockchain.NewBlockchain()
	peers := []string{"127.0.0.1:6001", "127.0.0.1:6002"}
	miner := miner.NewMiner(mempool, chain, 5, peers, "0000")

	if miner.Mempool != mempool {
		t.Errorf("Expected mempool to be correctly initialized")
	}
	if miner.Blockchain != chain {
		t.Errorf("Expected blockchain to be correctly initialized")
	}
	if miner.MaxBlockTransactions != 5 {
		t.Errorf("Expected max transactions to be 5, got %d", miner.MaxBlockTransactions)
	}
}

func TestPickTransactions_ValidMempool(t *testing.T) {
	mempool := blockchain.NewMempool()
	for i := 0; i < 10; i++ {
		tx := blockchain.Transaction{
			TxID: []byte(fmt.Sprintf("tx%d", i)),
		}
		mempool.AddTransaction(tx)
	}

	chain := blockchain.NewBlockchain()
	miner := miner.NewMiner(mempool, chain, 5, []string{}, "0000")
	transactions := miner.PickTransactions()

	if len(transactions) != 5 {
		t.Errorf("Expected 5 transactions, got %d", len(transactions))
	}
}

func TestMineBlock_WithTransactions(t *testing.T) {
	mempool := blockchain.NewMempool()
	for i := 0; i < 5; i++ {
		tx := blockchain.Transaction{
			TxID: []byte(fmt.Sprintf("tx%d", i)),
		}
		mempool.AddTransaction(tx)
	}

	chain := blockchain.NewBlockchain()

	ipfsClient := ipfs.NewMockIPFSClient(true)

	miner := miner.NewMiner(mempool, chain, 5, []string{}, "0000")
	miner.SetIPFSClient(ipfsClient)

	// Initialize the blockchain with a genesis block
	miner.InitializeBlockchain()

	block := miner.MineBlock()
	if block == nil {
		t.Fatalf("Failed to mine a block")
	}
	if len(block.Transactions) != 5 {
		t.Errorf("Expected 5 transactions in the block, got %d", len(block.Transactions))
	}
}

func TestMineBlock_EmptyMempool(t *testing.T) {
	mempool := blockchain.NewMempool()
	chain := blockchain.NewBlockchain()
	miner := miner.NewMiner(mempool, chain, 5, []string{}, "0000")

	block := miner.MineBlock()
	if block != nil {
		t.Errorf("Expected no block to be mined, but got one")
	}
}

func TestMineBlock_ChainIntegration(t *testing.T) {
	mempool := blockchain.NewMempool()
	for i := 0; i < 5; i++ {
		tx := blockchain.Transaction{
			TxID: []byte(fmt.Sprintf("tx%d", i)),
		}
		mempool.AddTransaction(tx)
	}

	chain := blockchain.NewBlockchain()
	chain.Reset() // Ensure a clean blockchain state

	// Initialize a mock IPFS client
	mockIPFSClient := &ipfs.MockIPFSClient{Valid: true}

	// Create a miner and assign the mock IPFS client
	miner := miner.NewMiner(mempool, chain, 5, []string{}, "0000")
	miner.IPFSClient = mockIPFSClient

	// Initialize the blockchain with a genesis block
	miner.InitializeBlockchain()

	// Mine a block
	block := miner.MineBlock()

	// Validate that a block was mined
	if block == nil {
		t.Fatalf("Failed to mine a block")
	}

	// Verify the block was added
	if len(chain.Blocks) != 2 { // 1 genesis block + 1 mined block
		t.Errorf("Expected blockchain to have 2 blocks, got %d", len(chain.Blocks))
	}

	// Validate the block's PreviousHash
	if !bytes.Equal(chain.Blocks[1].Header.PreviousHash, chain.Blocks[0].Header.MerkleRoot) {
		t.Errorf("Expected PreviousHash to be %x, got %x", chain.Blocks[0].Header.MerkleRoot, chain.Blocks[1].Header.PreviousHash)
	}

	log.Printf("TestMineBlock_ChainIntegration passed: Block added with Merkle Root: %x", block.Header.MerkleRoot)
}

func TestBroadcastBlock_ValidPeers(t *testing.T) {
	peers := []string{"127.0.0.1:6001", "127.0.0.1:6002"}
	mockBroadcaster := &MockBroadcaster{}

	block := network.BlockMessage{
		BlockID:      "block123",
		MerkleRoot:   "merkleRoot123",
		PreviousHash: "prevHash123",
		Transactions: []network.TxMessage{},
	}

	// Start mock peers to simulate valid listeners
	var wg sync.WaitGroup
	shutdown := make(chan bool)
	for _, peer := range peers {
		wg.Add(1)
		port := parsePort(peer)
		go StartMockPeer(port, shutdown, &wg)
	}

	// Allow mock peers to start
	time.Sleep(500 * time.Millisecond)

	// Simulate broadcasting
	mockBroadcaster.BroadcastBlock(block, peers)

	// Shut down mock peers
	close(shutdown)
	wg.Wait()

	// Validate logs
	if len(mockBroadcaster.BroadcastLogs) != len(peers) {
		t.Errorf("Expected %d broadcast attempts, got %d", len(peers), len(mockBroadcaster.BroadcastLogs))
	}

	// Verify content of the logs
	for _, peer := range peers {
		expectedLog := fmt.Sprintf("Peer: %s, Block ID: %s, Merkle Root: %s", peer, block.BlockID, block.MerkleRoot)
		found := false
		for _, log := range mockBroadcaster.BroadcastLogs {
			if log == expectedLog {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected log entry for peer %s not found", peer)
		}
	}
}

// Helper function to parse the port number from a peer address
func parsePort(peer string) int {
	_, portStr, err := net.SplitHostPort(peer)
	if err != nil {
		log.Fatalf("Failed to parse port: %v", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Failed to convert port to integer: %v", err)
	}
	return port
}

func TestBroadcastBlock_NoPeers(t *testing.T) {
	mempool := blockchain.NewMempool()
	chain := blockchain.NewBlockchain()

	// Create a mock broadcaster
	mockBroadcaster := &MockBroadcaster{
		BroadcastLogs: []string{}, // Ensure BroadcastLogs is initialized
	}

	// Create a miner and use the mock broadcaster's BroadcastBlock method
	peers := []string{}
	miner := miner.NewMiner(mempool, chain, 5, peers, "0000")
	miner.SetBroadcaster(mockBroadcaster)

	// Create a block and attempt to broadcast with no peers
	block := &blockchain.Block{}
	miner.BroadcastBlock(block)

	// Assert that no broadcast attempts were made
	if len(mockBroadcaster.BroadcastLogs) != 0 {
		t.Errorf("Expected 0 broadcast attempts, got %d", len(mockBroadcaster.BroadcastLogs))
	}
}

func TestMineBlock_HighThroughput(t *testing.T) {
	mempool := blockchain.NewMempool()
	for i := 0; i < 10000; i++ {
		tx := blockchain.Transaction{
			TxID: []byte(fmt.Sprintf("tx%d", i)),
		}
		mempool.AddTransaction(tx)
	}

	chain := blockchain.NewBlockchain()

	// Initialize the blockchain with a genesis block
	genesisBlock := &blockchain.Block{
		Header: blockchain.BlockHeader{
			PreviousHash: []byte("GENESIS"),
			Timestamp:    time.Now().UnixNano(),
		},
	}
	genesisBlock.ComputeMerkleRoot()
	err := chain.AddBlock(genesisBlock)
	if err != nil {
		t.Fatalf("Failed to add genesis block: %v", err)
	}

	// Initialize a mock IPFS client
	mockIPFSClient := &ipfs.MockIPFSClient{Valid: true}

	// Create a miner and assign the mock IPFS client
	miner := miner.NewMiner(mempool, chain, 100, []string{}, "0000")
	miner.IPFSClient = mockIPFSClient

	// Mine a block
	block := miner.MineBlock()

	// Validate the block was mined successfully
	if block == nil {
		t.Fatalf("Failed to mine a block")
	}

	// Verify that the block contains the expected number of transactions
	if len(block.Transactions) != 100 {
		t.Errorf("Expected 100 transactions in the block, got %d", len(block.Transactions))
	}
}

func TestMineBlock_ConcurrentMining(t *testing.T) {
	mempool := blockchain.NewMempool()
	for i := 0; i < 10; i++ {
		tx := blockchain.Transaction{
			TxID: []byte(fmt.Sprintf("tx%d", i)),
		}
		mempool.AddTransaction(tx)
	}

	chain := blockchain.NewBlockchain()
	chain.Reset() // Ensure a clean blockchain state

	// Add a genesis block to initialize the blockchain
	genesisBlock := &blockchain.Block{
		Header: blockchain.BlockHeader{
			PreviousHash: []byte("GENESIS"),
			Timestamp:    time.Now().UnixNano(),
		},
	}
	genesisBlock.ComputeMerkleRoot()
	err := chain.AddBlock(genesisBlock)
	if err != nil {
		t.Fatalf("Failed to add genesis block: %v", err)
	}

	var wg sync.WaitGroup

	for i := 0; i < 2; i++ { // Two miners
		wg.Add(1)
		go func(minerID int) {
			defer wg.Done()

			// Initialize a miner with a valid mock IPFS client
			mockIPFSClient := &ipfs.MockIPFSClient{Valid: true}
			miner := miner.NewMiner(mempool, chain, 5, []string{}, "0000")
			miner.IPFSClient = mockIPFSClient

			block := miner.MineBlock()
			if block == nil {
				t.Logf("Miner %d failed to mine a block", minerID)
			} else {
				t.Logf("Miner %d successfully mined a block with Merkle Root: %x", minerID, block.Header.MerkleRoot)
			}
		}(i)
	}
	wg.Wait()

	// Verify the number of blocks in the blockchain
	chain.Mutex.Lock()
	defer chain.Mutex.Unlock()

	if len(chain.Blocks) != 3 { // Expect 3 blocks: 1 genesis block + 2 mined blocks
		t.Errorf("Expected 3 blocks in the blockchain, got %d", len(chain.Blocks))
	}
}

func TestMineBlock_ValidTransactions(t *testing.T) {
	mempool := blockchain.NewMempool()
	for i := 0; i < 5; i++ {
		tx := blockchain.Transaction{
			TxID: []byte(fmt.Sprintf("tx%d", i)),
		}
		mempool.AddTransaction(tx)
	}

	chain := blockchain.NewBlockchain()

	// Add a genesis block to initialize the blockchain
	genesisBlock := &blockchain.Block{
		Header: blockchain.BlockHeader{
			PreviousHash: []byte("GENESIS"),
			Timestamp:    time.Now().UnixNano(),
		},
	}
	genesisBlock.ComputeMerkleRoot()
	err := chain.AddBlock(genesisBlock)
	if err != nil {
		t.Fatalf("Failed to add genesis block: %v", err)
	}

	// Mock IPFS client for validating transactions
	ipfsClient := ipfs.NewMockIPFSClient(true)

	// Create and configure the miner
	miner := miner.NewMiner(mempool, chain, 5, []string{}, "0000")
	miner.IPFSClient = ipfsClient

	// Mine a block
	block := miner.MineBlock()

	// Validate the mined block
	if block == nil {
		t.Fatalf("Expected a block to be mined")
	}
	if len(block.Transactions) != 5 {
		t.Errorf("Expected 5 transactions in the block, got %d", len(block.Transactions))
	}
}

func TestRunVMStandalone(t *testing.T) {
	// Mock Dataset
	dataset := [][]float64{
		{1.0, 2.0},
		{2.0, 1.0},
		{3.0, 4.0},
		{5.0, 7.0},
		{3.5, 5.0},
		{4.5, 5.0},
		{3.5, 4.5},
		{6.0, 7.0},
		{7.0, 8.0},
		{8.0, 9.0},
	}

	// Step 1: Serialize dataset to JSON
	dataBytes, err := json.Marshal(dataset)
	if err != nil {
		t.Fatalf("Failed to serialize dataset: %v", err)
	}

	// Step 2: Run KMeans using the vm package
	vmOutput, err := vm.RunKMeans(dataBytes)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}

	// Step 3: Deserialize VM output
	var centroids [][]float64
	err = json.Unmarshal(vmOutput, &centroids)
	if err != nil {
		t.Fatalf("Failed to deserialize VM output: %v", err)
	}

	// Log the final centroids
	t.Logf("VM execution successful. Output: %+v", centroids)
}

func TestRealIntegration(t *testing.T) {
	// Step 1: Load configuration
	cfg, err := config.LoadConfig("./config.json")
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Step 2: Initialize IPFS client
	ipfsClient := ipfs.NewIPFSClient(cfg.IPFSGatewayURL)

	// Step 3: Create a miner
	mempool := blockchain.NewMempool()
	chain := blockchain.NewBlockchain()
	testMiner := miner.NewMiner(mempool, chain, 5, []string{}, "00")

	testMiner.IPFSClient = ipfsClient

	// Step 4: Add transactions to the mempool
	// Step 4: Add transactions to the mempool
	for i := 0; i < cfg.MaxBlockTransactions; i++ {
		tx := blockchain.Transaction{
			TxID:          []byte(fmt.Sprintf("tx%d", i)), // Unique transaction ID
			DataHash:      cfg.DatasetHash,                // Valid dataset hash
			AlgorithmHash: cfg.AlgorithmHash,              // Valid algorithm hash
		}
		mempool.AddTransaction(tx)
	}
	// Add debugging before calling MineBlock
	t.Logf("Miner initialized: %+v", testMiner)
	t.Logf("Mempool transactions: %v", len(mempool.Transactions))
	block := testMiner.MineBlock()
	if block == nil {
		t.Fatalf("Failed to mine a block. Mempool transactions: %v", len(mempool.Transactions))
	}

	// Step 6: Validate VMOutputs for all transactions
	for _, tx := range block.Transactions {
		if len(tx.VMOutput) == 0 {
			t.Errorf("Transaction %x has no VMOutput", tx.TxID)
		} else {
			t.Logf("Transaction %x VMOutput: %s", tx.TxID, string(tx.VMOutput))
		}
	}

	// Validate Proof of Work
	if !pow.ValidateProofOfWork(block.Header.Bytes(), block.Header.Nonce, cfg.MiningDifficultyTarget) {
		t.Errorf("Block's PoW is invalid. Nonce: %d, Hash: %x", block.Header.Nonce, block.Header.Hash)
	} else {
		t.Logf("Block's PoW validated successfully. Nonce: %d, Hash: %x", block.Header.Nonce, block.Header.Hash)
	}

	// Step 7: Verify the mempool is cleared after mining
	if len(mempool.GetAllTransactions()) != 0 {
		t.Errorf("Mempool is not cleared after mining")
	} else {
		t.Logf("Mempool cleared successfully")
	}

}

func TestFetchData(t *testing.T) {
	client := ipfs.NewIPFSClient("http://127.0.0.1:5001")
	data, err := client.FetchData("QmQA9Bcs2A7BsdCeinLRi92mYHRpQbpUt89usVHY25zZdN")
	if err != nil {
		t.Fatalf("Failed to fetch data: %v", err)
	}
	t.Logf("Fetched data: %s", string(data))
}

func TestFetchAlgoirthm(t *testing.T) {
	client := ipfs.NewIPFSClient("http://127.0.0.1:5001")
	data, err := client.FetchAlgorithm("QmX51AYg6FHv3c6WUPVfK4bkdbtLZg9fNY5jQbniKBeeFG")
	if err != nil {
		t.Fatalf("Failed to fetch data: %v", err)
	}
	t.Logf("Fetched data: %s", string(data))
}

func TestSingleTransactionIntegration(t *testing.T) {
	// Step 1: Load configuration
	cfg, err := config.LoadConfig("./config.json")
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Step 2: Initialize IPFS client
	ipfsClient := ipfs.NewIPFSClient(cfg.IPFSGatewayURL)

	// Step 3: Create a miner
	mempool := blockchain.NewMempool()
	chain := blockchain.NewBlockchain()
	testMiner := miner.NewMiner(mempool, chain, 5, []string{}, "00")
	testMiner.IPFSClient = ipfsClient

	// Step 4: Add a single transaction to the mempool
	tx := blockchain.Transaction{
		TxID:          []byte("tx1"),     // Unique transaction ID
		DataHash:      cfg.DatasetHash,   // Valid dataset hash
		AlgorithmHash: cfg.AlgorithmHash, // Valid algorithm hash
	}
	mempool.AddTransaction(tx)

	// Debugging logs before mining
	t.Logf("Miner initialized: %+v", testMiner)
	t.Logf("Mempool transactions: %v", len(mempool.Transactions))

	// Step 5: Mine a block
	block := testMiner.MineBlock()
	if block == nil {
		t.Fatalf("Failed to mine a block. Mempool transactions: %v", len(mempool.Transactions))
	}

	// Step 6: Validate the single transaction's VMOutput
	for _, tx := range block.Transactions {
		if len(tx.VMOutput) == 0 {
			t.Errorf("Transaction %x has no VMOutput", tx.TxID)
		} else {
			t.Logf("Transaction %x VMOutput: %s", tx.TxID, string(tx.VMOutput))
		}
	}

	log.Printf("Before Validation Header: %x", block.Header.Bytes())
	log.Printf("Validation Difficulty Target: %s", cfg.MiningDifficultyTarget)

	// Validate Proof of Work
	if !pow.ValidateProofOfWork(block.Header.Bytes(), block.Header.Nonce, cfg.MiningDifficultyTarget) {

		t.Errorf("Block's PoW is invalid. Nonce: %d, Hash: %x", block.Header.Nonce, block.Header.Hash)
	} else {
		t.Logf("Block's PoW validated successfully. Nonce: %d, Hash: %x", block.Header.Nonce, block.Header.Hash)
	}

	// Step 7: Verify the mempool is cleared after mining
	if len(mempool.GetAllTransactions()) != 0 {
		t.Errorf("Mempool is not cleared after mining")
	} else {
		t.Logf("Mempool cleared successfully")
	}
}
