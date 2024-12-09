package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/TalhaArjumand/ai-blockchain/pkg/ipfs"
)

// TestGenerateTxID_ValidTransaction tests that GenerateTxID produces a valid, non-empty TxID
func TestGenerateTxID_ValidTransaction(t *testing.T) {
	tx := Transaction{
		DataHash:      "dataHashExample",
		AlgorithmHash: "algorithmHashExample",
		Metadata:      "metadataExample",
	}
	tx.GenerateTxID()

	if len(tx.TxID) == 0 {
		t.Errorf("Expected non-empty TxID, got %v", tx.TxID)
	}
}

// TestGenerateTxID_DifferentTransactions tests that two different transactions produce different TxIDs
func TestGenerateTxID_DifferentTransactions(t *testing.T) {
	tx1 := Transaction{
		DataHash:      "dataHashExample1",
		AlgorithmHash: "algorithmHashExample1",
		Metadata:      "metadataExample1",
	}
	tx2 := Transaction{
		DataHash:      "dataHashExample2",
		AlgorithmHash: "algorithmHashExample2",
		Metadata:      "metadataExample2",
	}

	tx1.GenerateTxID()
	tx2.GenerateTxID()

	if bytes.Equal(tx1.TxID, tx2.TxID) {
		t.Errorf("Expected different TxIDs for different transactions, but got identical TxIDs")
	}
}

// TestGenerateTxID_SameTransactionDifferentTimestamps tests that the same transaction produces different TxIDs when generated at different timestamps
func TestGenerateTxID_SameTransactionDifferentTimestamps(t *testing.T) {
	tx := Transaction{
		DataHash:      "dataHashExample",
		AlgorithmHash: "algorithmHashExample",
		Metadata:      "metadataExample",
	}

	tx.GenerateTxID()
	txID1 := make([]byte, len(tx.TxID))
	copy(txID1, tx.TxID)

	time.Sleep(1 * time.Millisecond) // Ensure a different timestamp
	tx.GenerateTxID()

	if bytes.Equal(txID1, tx.TxID) {
		t.Errorf("Expected different TxIDs for same transaction at different timestamps, but got identical TxIDs")
	}
}

// TestGenerateTxID_ExcludeVMOutput tests that VMOutput is excluded from the TxID calculation
func TestGenerateTxID_ExcludeVMOutput(t *testing.T) {
	tx := Transaction{
		DataHash:      "dataHashExample",
		AlgorithmHash: "algorithmHashExample",
		Metadata:      "metadataExample",
		VMOutput:      []byte("this should not affect TxID"),
	}
	tx.GenerateTxID()

	txWithoutVMOutput := Transaction{
		DataHash:      tx.DataHash,
		AlgorithmHash: tx.AlgorithmHash,
		Metadata:      tx.Metadata,
		Timestamp:     tx.Timestamp,
	}
	data, _ := json.Marshal(txWithoutVMOutput)
	expectedHash := sha256.Sum256(data)

	if !bytes.Equal(tx.TxID, expectedHash[:]) {
		t.Errorf("Expected TxID to exclude VMOutput, but it did not")
	}
}

// TestGenerateTxID_EmptyTransaction tests that an empty transaction can still generate a valid TxID
func TestGenerateTxID_EmptyTransaction(t *testing.T) {
	tx := Transaction{}
	tx.GenerateTxID()

	if len(tx.TxID) == 0 {
		t.Errorf("Expected non-empty TxID for empty transaction, but got %v", tx.TxID)
	}
}

// TestGenerateTxID_SerializationIntegrity tests that altering a field changes the TxID
func TestGenerateTxID_SerializationIntegrity(t *testing.T) {
	tx := Transaction{
		DataHash:      "dataHashExample",
		AlgorithmHash: "algorithmHashExample",
		Metadata:      "metadataExample",
	}

	tx.GenerateTxID()
	originalTxID := make([]byte, len(tx.TxID))
	copy(originalTxID, tx.TxID)

	tx.DataHash = "alteredDataHash"
	tx.GenerateTxID()

	if bytes.Equal(originalTxID, tx.TxID) {
		t.Errorf("Expected different TxID after altering DataHash, but got identical TxIDs")
	}
}

// TestGenerateTxID_LargeMetadata tests that large metadata is handled correctly
func TestGenerateTxID_LargeMetadata(t *testing.T) {
	largeMetadata := make([]byte, 1<<20) // 1 MB of metadata
	for i := range largeMetadata {
		largeMetadata[i] = 'a'
	}

	tx := Transaction{
		DataHash:      "dataHashExample",
		AlgorithmHash: "algorithmHashExample",
		Metadata:      string(largeMetadata),
	}
	tx.GenerateTxID()

	if len(tx.TxID) == 0 {
		t.Errorf("Expected non-empty TxID for large metadata, but got %v", tx.TxID)
	}
}

func TestAddTransactionToMempool(t *testing.T) {
	mempool := NewMempool()
	tx := Transaction{TxID: []byte("tx1")}
	mempool.AddTransaction(tx)

	if !mempool.HasTransaction("tx1") {
		t.Errorf("Expected transaction to be in the mempool, but it was not")
	}
}

func TestRemoveTransactionFromMempool(t *testing.T) {
	mempool := NewMempool()
	tx := Transaction{TxID: []byte("tx1")}
	mempool.AddTransaction(tx)
	mempool.RemoveTransaction("tx1")

	if mempool.HasTransaction("tx1") {
		t.Errorf("Expected transaction to be removed from the mempool, but it was not")
	}
}

func TestGetTransactionFromMempool(t *testing.T) {
	mempool := NewMempool()
	tx := Transaction{TxID: []byte("tx1")}
	mempool.AddTransaction(tx)

	retrievedTx := mempool.GetTransaction("tx1")
	if retrievedTx == nil || string(retrievedTx.TxID) != "tx1" {
		t.Errorf("Expected to retrieve transaction 'tx1', but did not")
	}
}

func TestMempoolConcurrency(t *testing.T) {
	mempool := NewMempool()
	tx1 := Transaction{TxID: []byte("tx1")}
	tx2 := Transaction{TxID: []byte("tx2")}
	var wg sync.WaitGroup

	// Add transactions concurrently
	wg.Add(2)
	go func() {
		defer wg.Done()
		mempool.AddTransaction(tx1)
	}()
	go func() {
		defer wg.Done()
		mempool.AddTransaction(tx2)
	}()
	wg.Wait()

	// Verify both transactions were added
	if !mempool.HasTransaction("tx1") || !mempool.HasTransaction("tx2") {
		t.Errorf("Expected transactions to be present in mempool, but they were not")
	}

	// Remove transactions concurrently
	wg.Add(2)
	go func() {
		defer wg.Done()
		mempool.RemoveTransaction("tx1")
	}()
	go func() {
		defer wg.Done()
		mempool.RemoveTransaction("tx2")
	}()
	wg.Wait()

	// Verify mempool is empty
	if mempool.HasTransaction("tx1") || mempool.HasTransaction("tx2") {
		t.Errorf("Expected mempool to be empty after concurrent operations, but it was not")
	}
}

func TestGenerateTxID_InvalidData(t *testing.T) {
	tx := Transaction{
		DataHash:      "",
		AlgorithmHash: "",
		Metadata:      "",
	}
	tx.GenerateTxID()

	if len(tx.TxID) == 0 {
		t.Errorf("Expected TxID for transaction with missing fields, but got %v", tx.TxID)
	}
}

func TestFetchInputsValid(t *testing.T) {
	mockClient := &ipfs.MockIPFSClient{Valid: true}
	tx := Transaction{DataHash: "validDataHash", AlgorithmHash: "validAlgoHash"}

	data, algo, err := tx.FetchInputs(mockClient)
	if err != nil || data == nil || algo == nil {
		t.Errorf("Expected valid inputs from IPFS, but got error: %v", err)
	}
}

func TestFetchInputsInvalid(t *testing.T) {
	mockClient := &ipfs.MockIPFSClient{Valid: false}
	tx := Transaction{DataHash: "invalidDataHash", AlgorithmHash: "invalidAlgoHash"}

	data, algo, err := tx.FetchInputs(mockClient)
	if err == nil || data != nil || algo != nil {
		t.Errorf("Expected error and nil inputs, but got data: %v, algo: %v", data, algo)
	}
}
