package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"sync"
	"time"

	"github.com/TalhaArjumand/ai-blockchain/pkg/ipfs"
)

type Transaction struct {
	TxID          []byte
	DataHash      string // IPFS hash of the data
	AlgorithmHash string // IPFS hash of the algorithm
	Metadata      string // Optional info
	VMOutput      []byte // VM output result
	Timestamp     int64
}

// Mempool represents a pool of unconfirmed transactions
type Mempool struct {
	Transactions map[string]Transaction
	Mutex        sync.Mutex // For thread-safe access
}

// Generate a transaction ID (TxID) based on all fields except VMOutput
func (tx *Transaction) GenerateTxID() {
	tx.VMOutput = nil // Exclude VMOutput
	tx.Timestamp = time.Now().UnixNano()
	data, _ := json.Marshal(tx)
	hash := sha256.Sum256(data)
	tx.TxID = hash[:]
}

func (tx *Transaction) FetchInputs(client ipfs.IPFSInterface) ([]byte, []byte, error) {
	// Fetch dataset
	data, err := client.FetchData(tx.DataHash)
	if err != nil {
		return nil, nil, err
	}

	// Fetch algorithm
	algo, err := client.FetchAlgorithm(tx.AlgorithmHash)
	if err != nil {
		return nil, nil, err
	}

	return data, algo, nil
}

// NewMempool initializes and returns a new Mempool
func NewMempool() *Mempool {
	return &Mempool{
		Transactions: make(map[string]Transaction),
	}
}

// AddTransaction adds a transaction to the mempool
func (m *Mempool) AddTransaction(tx Transaction) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	m.Transactions[string(tx.TxID)] = tx
}

// HasTransaction checks if a transaction exists in the mempool
func (m *Mempool) HasTransaction(txID string) bool {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	_, exists := m.Transactions[txID]
	return exists
}

// RemoveTransaction removes a transaction from the mempool
func (m *Mempool) RemoveTransaction(txID string) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	delete(m.Transactions, txID)
}

func (m *Mempool) GetTransaction(hash string) *Transaction {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	tx, exists := m.Transactions[hash]
	if exists {
		return &tx
	}
	return nil
}
