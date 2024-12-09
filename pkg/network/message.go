package network

import (
	"encoding/json"

	"github.com/TalhaArjumand/ai-blockchain/pkg/blockchain"
)

// TxMessage encapsulates a new transaction
type TxMessage struct {
	Type      string `json:"type"`      // Message type, e.g., "transaction"
	TxID      string `json:"tx_id"`     // Transaction ID
	DataHash  string `json:"data_hash"` // Hash of the dataset
	AlgoHash  string `json:"algo_hash"` // Hash of the algorithm
	Metadata  string `json:"metadata"`  // Metadata describing the transaction
	Timestamp int64  `json:"timestamp"` // Timestamp of the transaction
}

// BlockMessage encapsulates a new block

type BlockMessage struct {
	BlockID      string      `json:"block_id"`
	MerkleRoot   string      `json:"merkle_root"`
	PreviousHash string      `json:"previous_hash"`
	Transactions []TxMessage `json:"transactions"`
	Timestamp    int64       `json:"timestamp"`
}

type GetBlocksMessage struct {
	RequestingNode string `json:"requesting_node"` // The node making the request
	StartHeight    int    `json:"start_height"`    // Starting block height
	EndHeight      int    `json:"end_height"`      // Ending block height
}

type BlocksMessage struct {
	Blocks []blockchain.Block `json:"blocks"` // List of blocks to send back
}

type InvMessage struct {
	Hashes []string `json:"hashes"`
	Type   string   `json:"type"` // "block" or "transaction"
}

type GetDataMessage struct {
	Type        string `json:"type"`         // "block" or "transaction"
	Hash        string `json:"hash"`         // Hash of the requested data
	PeerAddress string `json:"peer_address"` // Address of the requesting peer
}

// SerializeMessage serializes a message into JSON
func SerializeMessage(message interface{}) ([]byte, error) {
	return json.Marshal(message)
}

// DeserializeMessage deserializes JSON into a generic map
func DeserializeMessage(data []byte) (map[string]interface{}, error) {
	var msg map[string]interface{}
	err := json.Unmarshal(data, &msg)
	return msg, err
}
