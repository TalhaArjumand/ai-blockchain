package main

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/TalhaArjumand/ai-blockchain/pkg/blockchain"
	"github.com/TalhaArjumand/ai-blockchain/pkg/network"
)

func TestMessageSerialization(t *testing.T) {
	tx := network.TxMessage{
		TxID:      "tx123",
		DataHash:  "data123",
		AlgoHash:  "algo123",
		Metadata:  "Test Transaction",
		Timestamp: 1234567890,
	}
	serialized, err := network.SerializeMessage(tx)
	if err != nil {
		t.Fatalf("Failed to serialize TxMessage: %v", err)
	}

	var deserialized network.TxMessage
	err = json.Unmarshal(serialized, &deserialized)
	if err != nil || deserialized.TxID != tx.TxID {
		t.Fatalf("Failed to deserialize TxMessage: %v", err)
	}
}

func TestGetBlock(t *testing.T) {
	blockchainInstance := blockchain.NewBlockchain()
	block := &blockchain.Block{
		Header: blockchain.BlockHeader{
			MerkleRoot: []byte("merkle123"),
		},
	}
	blockchainInstance.Blocks[0] = block
	fetchedBlock := blockchainInstance.GetBlock("merkle123")
	if fetchedBlock == nil || string(fetchedBlock.Header.MerkleRoot) != "merkle123" {
		t.Fatalf("GetBlock failed to fetch the correct block")
	}
}

func TestMempoolAddTransaction(t *testing.T) {
	mempool := blockchain.NewMempool()
	tx := blockchain.Transaction{
		TxID: []byte("tx123"),
	}
	mempool.AddTransaction(tx)
	if !mempool.HasTransaction("tx123") {
		t.Fatalf("Transaction was not added to the mempool")
	}
}

func TestAddMultipleBlocks(t *testing.T) {
	blockchainInstance := blockchain.NewBlockchain()

	// Add multiple blocks
	for i := 1; i <= 5; i++ {
		block := &blockchain.Block{
			Header: blockchain.BlockHeader{
				MerkleRoot: []byte(fmt.Sprintf("merkle%d", i)),
			},
		}
		blockchainInstance.Blocks[i] = block
	}

	// Verify blocks were added
	for i := 1; i <= 5; i++ {
		fetchedBlock := blockchainInstance.GetBlock(fmt.Sprintf("merkle%d", i))
		if fetchedBlock == nil || string(fetchedBlock.Header.MerkleRoot) != fmt.Sprintf("merkle%d", i) {
			t.Fatalf("Block %d was not correctly fetched", i)
		}
	}
}

func TestFetchNonExistentBlock(t *testing.T) {
	blockchainInstance := blockchain.NewBlockchain()

	// Attempt to fetch a block that doesn't exist
	block := blockchainInstance.GetBlock("nonexistent")
	if block != nil {
		t.Fatalf("Expected nil, but got a block: %+v", block)
	}
}

func TestDuplicateTransactionInMempool(t *testing.T) {
	mempool := blockchain.NewMempool()

	tx := blockchain.Transaction{
		TxID: []byte("tx123"),
	}
	mempool.AddTransaction(tx)
	mempool.AddTransaction(tx) // Add duplicate transaction

	// Verify only one instance of the transaction exists
	if !mempool.HasTransaction("tx123") {
		t.Fatalf("Transaction was not found in the mempool")
	}

	if len(mempool.Transactions) > 1 {
		t.Fatalf("Duplicate transaction was added to the mempool")
	}
}

func TestSerializeDeserializeError(t *testing.T) {
	// Pass invalid data to serialization
	invalidData := make(chan int) // Channels can't be serialized
	_, err := network.SerializeMessage(invalidData)
	if err == nil {
		t.Fatalf("Expected serialization error for invalid data")
	}

	// Pass invalid JSON for deserialization
	invalidJSON := []byte(`{"type": invalid}`)
	var deserialized network.TxMessage
	err = json.Unmarshal(invalidJSON, &deserialized)
	if err == nil {
		t.Fatalf("Expected deserialization error for invalid JSON")
	}
}

func TestFetchBlockByHeight(t *testing.T) {
	blockchainInstance := blockchain.NewBlockchain()

	block := &blockchain.Block{
		Header: blockchain.BlockHeader{
			MerkleRoot: []byte("merkle123"),
		},
	}
	blockchainInstance.Blocks[1] = block

	fetchedBlock := blockchainInstance.GetBlockByHeight(1)
	if fetchedBlock == nil || string(fetchedBlock.Header.MerkleRoot) != "merkle123" {
		t.Fatalf("Failed to fetch block by height")
	}
}

func TestConcurrentBlockAddition(t *testing.T) {
	blockchainInstance := blockchain.NewBlockchain()

	// Simulate concurrent block addition
	for i := 0; i < 10; i++ {
		go func(index int) {
			block := &blockchain.Block{
				Header: blockchain.BlockHeader{
					MerkleRoot: []byte(fmt.Sprintf("merkle%d", index)),
				},
			}
			blockchainInstance.AddBlock(block)
		}(i)
	}

	time.Sleep(1 * time.Second) // Allow goroutines to complete

	// Verify all blocks are added
	for i := 0; i < 10; i++ {
		block := blockchainInstance.GetBlock(fmt.Sprintf("merkle%d", i))
		if block == nil {
			t.Errorf("Block %d was not added to the blockchain", i)
		}
	}
}

func TestEndToEndIntegration(t *testing.T) {
	port := "6000"
	go network.StartServer(port, func(message []byte) {
		t.Logf("Message received: %s", string(message))
	})

	time.Sleep(1 * time.Second) // Allow the server to start

	// Simulate sending a block
	peerAddress := "localhost:" + port
	block := network.BlockMessage{
		BlockID:      "block123",
		MerkleRoot:   "merkleRoot123",
		PreviousHash: "prevHash123",
		Transactions: []network.TxMessage{},
		Timestamp:    time.Now().Unix(),
	}

	serializedBlock, err := network.SerializeMessage(block)
	if err != nil {
		t.Fatalf("Failed to serialize block: %v", err)
	}

	err = network.SendMessage(peerAddress, serializedBlock)
	if err != nil {
		t.Fatalf("Failed to send block to peer: %v", err)
	}
}
