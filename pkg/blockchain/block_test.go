package blockchain

import (
	"bytes"
	"testing"
	"time"

	"github.com/TalhaArjumand/ai-blockchain/pkg/ipfs"
)

func TestBlockWithTransactions(t *testing.T) {
	block := Block{
		Transactions: []Transaction{
			{TxID: []byte("tx1"), VMOutput: []byte("output1")},
			{TxID: []byte("tx2"), VMOutput: []byte("output2")},
		},
	}

	block.ComputeMerkleRoot()
	block.ComputeVMOutputsHash()

	if len(block.Header.MerkleRoot) == 0 {
		t.Errorf("Expected non-zero Merkle Root")
	}

	if len(block.Header.VMOutputsHash) == 0 {
		t.Errorf("Expected non-zero VM Outputs Hash")
	}
}

func TestBlockWithNoTransactions(t *testing.T) {
	block := Block{}

	block.ComputeMerkleRoot()
	block.ComputeVMOutputsHash()

	if len(block.Header.MerkleRoot) != 0 {
		t.Errorf("Expected empty Merkle Root for empty block")
	}

	if len(block.Header.VMOutputsHash) != 0 {
		t.Errorf("Expected empty VM Outputs Hash for empty block")
	}
}

func TestGenerateTxID(t *testing.T) {
	tx := Transaction{
		DataHash:      "dataHashExample",
		AlgorithmHash: "algorithmHashExample",
		Metadata:      "metadataExample",
	}
	tx.GenerateTxID()

	if len(tx.TxID) == 0 {
		t.Errorf("Expected non-empty TxID")
	}
}

func TestGenerateTxIDUniqueness(t *testing.T) {
	tx1 := Transaction{
		DataHash:      "dataHashExample",
		AlgorithmHash: "algorithmHashExample",
		Metadata:      "metadataExample",
	}
	tx2 := Transaction{
		DataHash:      "dataHashExample",
		AlgorithmHash: "algorithmHashExample",
		Metadata:      "metadataExample",
	}

	tx1.GenerateTxID()
	time.Sleep(10 * time.Millisecond) // Ensure timestamps are distinct
	tx2.GenerateTxID()

	if bytes.Equal(tx1.TxID, tx2.TxID) {
		t.Errorf("Expected different TxIDs for transactions with different timestamps")
	}
}

func TestComputeVMOutputsHashEmpty(t *testing.T) {
	block := Block{
		Transactions: []Transaction{},
	}

	block.ComputeVMOutputsHash()

	if len(block.Header.VMOutputsHash) != 0 {
		t.Errorf("Expected empty VM Outputs Hash for block with no transactions")
	}
}

func TestMerkleRootWithDuplicateTransactions(t *testing.T) {
	block := Block{
		Transactions: []Transaction{
			{TxID: []byte("tx1")},
			{TxID: []byte("tx1")},
		},
	}

	block.ComputeMerkleRoot()

	if len(block.Header.MerkleRoot) == 0 {
		t.Errorf("Expected non-zero Merkle Root for duplicate transactions")
	}
}

func TestMerkleRootWithOddTransactions(t *testing.T) {
	block := Block{
		Transactions: []Transaction{
			{TxID: []byte("tx1")},
			{TxID: []byte("tx2")},
			{TxID: []byte("tx3")},
		},
	}

	block.ComputeMerkleRoot()

	if len(block.Header.MerkleRoot) == 0 {
		t.Errorf("Expected non-zero Merkle Root for odd number of transactions")
	}
}

func TestValidateTransactionsValid(t *testing.T) {
	mockClient := &ipfs.MockIPFSClient{
		Valid: true,
	}

	block := Block{
		Transactions: []Transaction{
			{TxID: []byte("tx1")},
		},
	}

	if !block.ValidateTransactions(mockClient) {
		t.Errorf("Expected transactions to be valid")
	}
}

func TestValidateTransactionsInvalid(t *testing.T) {
	mockClient := &ipfs.MockIPFSClient{
		Valid: false,
	}

	block := Block{
		Transactions: []Transaction{
			{TxID: []byte("tx1")},
		},
	}

	if block.ValidateTransactions(mockClient) {
		t.Errorf("Expected transactions to be invalid")
	}
}
