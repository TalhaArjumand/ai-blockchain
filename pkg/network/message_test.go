package network

import (
	"encoding/json"
	"testing"
)

func TestMessageSerialization(t *testing.T) {
	// Create a TxMessage with fields matching the struct definition
	originalMessage := TxMessage{
		Type:      "transaction",
		TxID:      "12345",
		DataHash:  "sample_data_hash",
		AlgoHash:  "sample_algo_hash",
		Metadata:  "Sample Metadata",
		Timestamp: 1234567890,
	}

	// Serialize the message
	serialized, err := SerializeMessage(originalMessage)
	if err != nil {
		t.Fatalf("Error serializing message: %v", err)
	}

	// Deserialize the message
	deserialized := TxMessage{}
	err = json.Unmarshal(serialized, &deserialized)
	if err != nil {
		t.Fatalf("Error deserializing message: %v", err)
	}

	// Check if the deserialized message matches the original
	if deserialized.Type != originalMessage.Type || deserialized.TxID != originalMessage.TxID {
		t.Errorf("Deserialized message does not match original: %+v", deserialized)
	}
}

func TestTxMessageSerialization(t *testing.T) {
	originalMessage := TxMessage{
		Type:      "transaction",
		TxID:      "12345",
		DataHash:  "sample_data_hash",
		AlgoHash:  "sample_algo_hash",
		Metadata:  "Sample Metadata",
		Timestamp: 1234567890,
	}

	serialized, err := SerializeMessage(originalMessage)
	if err != nil {
		t.Fatalf("Error serializing TxMessage: %v", err)
	}

	deserialized := TxMessage{}
	err = json.Unmarshal(serialized, &deserialized)
	if err != nil {
		t.Fatalf("Error deserializing TxMessage: %v", err)
	}

	if deserialized != originalMessage {
		t.Errorf("Deserialized TxMessage does not match original: %+v", deserialized)
	}
}

func TestBlockMessageSerialization(t *testing.T) {
	originalMessage := BlockMessage{
		BlockID:      "block123",
		MerkleRoot:   "merkleRoot123",
		PreviousHash: "prevHash123",
		Transactions: []TxMessage{
			{Type: "transaction", TxID: "tx1", DataHash: "data1", AlgoHash: "algo1", Metadata: "meta1", Timestamp: 12345},
			{Type: "transaction", TxID: "tx2", DataHash: "data2", AlgoHash: "algo2", Metadata: "meta2", Timestamp: 67890},
		},
		Timestamp: 1234567890,
	}

	serialized, err := SerializeMessage(originalMessage)
	if err != nil {
		t.Fatalf("Error serializing BlockMessage: %v", err)
	}

	deserialized := BlockMessage{}
	err = json.Unmarshal(serialized, &deserialized)
	if err != nil {
		t.Fatalf("Error deserializing BlockMessage: %v", err)
	}

	if deserialized.BlockID != originalMessage.BlockID || len(deserialized.Transactions) != len(originalMessage.Transactions) {
		t.Errorf("Deserialized BlockMessage does not match original: %+v", deserialized)
	}
}

func TestDeserializeInvalidJSON(t *testing.T) {
	invalidJSON := []byte(`{invalid_json}`)
	_, err := DeserializeMessage(invalidJSON)
	if err == nil {
		t.Fatalf("Expected error while deserializing invalid JSON but got none")
	}
}

func TestEmptyMessageSerialization(t *testing.T) {
	originalMessage := TxMessage{}

	serialized, err := SerializeMessage(originalMessage)
	if err != nil {
		t.Fatalf("Error serializing empty TxMessage: %v", err)
	}

	deserialized := TxMessage{}
	err = json.Unmarshal(serialized, &deserialized)
	if err != nil {
		t.Fatalf("Error deserializing empty TxMessage: %v", err)
	}

	if deserialized != originalMessage {
		t.Errorf("Deserialized empty TxMessage does not match original: %+v", deserialized)
	}
}

func TestNestedStructureSerialization(t *testing.T) {
	originalMessage := BlockMessage{
		BlockID:      "block123",
		MerkleRoot:   "merkleRoot123",
		PreviousHash: "prevHash123",
		Transactions: []TxMessage{
			{Type: "transaction", TxID: "tx1", DataHash: "data1", AlgoHash: "algo1", Metadata: "meta1", Timestamp: 12345},
		},
		Timestamp: 1234567890,
	}

	serialized, err := SerializeMessage(originalMessage)
	if err != nil {
		t.Fatalf("Error serializing nested structure: %v", err)
	}

	deserialized := BlockMessage{}
	err = json.Unmarshal(serialized, &deserialized)
	if err != nil {
		t.Fatalf("Error deserializing nested structure: %v", err)
	}

	if deserialized.BlockID != originalMessage.BlockID {
		t.Errorf("Deserialized nested structure does not match original: %+v", deserialized)
	}
}
