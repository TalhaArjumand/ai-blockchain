package network

import (
	"encoding/json"
	"os"
	"testing"
)

func TestLoadPeers(t *testing.T) {
	// Create a temporary peers.json file
	peersData := `[{"host":"127.0.0.1","port":"5001"},{"host":"127.0.0.1","port":"5002"}]`
	err := os.WriteFile("peers_test.json", []byte(peersData), 0644)
	if err != nil {
		t.Fatalf("Error creating peers_test.json: %v", err)
	}
	defer os.Remove("peers_test.json")

	// Load peers from the test file
	peers, err := LoadPeers("peers_test.json")
	if err != nil {
		t.Fatalf("Error loading peers: %v", err)
	}

	// Verify the peers
	if len(peers) != 2 {
		t.Errorf("Expected 2 peers, got %d", len(peers))
	}
	if peers[0].Host != "127.0.0.1" || peers[0].Port != "5001" {
		t.Errorf("Peer 1 data mismatch: %v", peers[0])
	}
	if peers[1].Host != "127.0.0.1" || peers[1].Port != "5002" {
		t.Errorf("Peer 2 data mismatch: %v", peers[1])
	}
}

func TestLoadPeers_Success(t *testing.T) {
	// Create a temporary peers.json file
	peersData := `[{"host":"127.0.0.1","port":"5001"},{"host":"127.0.0.1","port":"5002"}]`
	err := os.WriteFile("peers_test.json", []byte(peersData), 0644)
	if err != nil {
		t.Fatalf("Error creating peers_test.json: %v", err)
	}
	defer os.Remove("peers_test.json")

	// Load peers from the test file
	peers, err := LoadPeers("peers_test.json")
	if err != nil {
		t.Fatalf("Error loading peers: %v", err)
	}

	// Verify the peers
	if len(peers) != 2 {
		t.Errorf("Expected 2 peers, got %d", len(peers))
	}
	if peers[0].Host != "127.0.0.1" || peers[0].Port != "5001" {
		t.Errorf("Peer 1 data mismatch: %v", peers[0])
	}
	if peers[1].Host != "127.0.0.1" || peers[1].Port != "5002" {
		t.Errorf("Peer 2 data mismatch: %v", peers[1])
	}
}

func TestLoadPeers_FileNotFound(t *testing.T) {
	_, err := LoadPeers("non_existent_file.json")
	if err == nil {
		t.Error("Expected error for missing file, but got none")
	}
}

func TestLoadPeers_InvalidJSON(t *testing.T) {
	// Create a temporary invalid JSON file
	invalidData := `INVALID JSON`
	err := os.WriteFile("invalid_peers_test.json", []byte(invalidData), 0644)
	if err != nil {
		t.Fatalf("Error creating invalid_peers_test.json: %v", err)
	}
	defer os.Remove("invalid_peers_test.json")

	// Attempt to load the invalid JSON
	_, err = LoadPeers("invalid_peers_test.json")
	if err == nil {
		t.Error("Expected error for invalid JSON, but got none")
	}
}

func TestSavePeers_Success(t *testing.T) {
	// Prepare test peers
	peers := []Peer{
		{Host: "127.0.0.1", Port: "6001"},
		{Host: "192.168.1.1", Port: "6002"},
	}

	// Save peers to a temporary file
	err := SavePeers("peers_save_test.json", peers)
	if err != nil {
		t.Fatalf("Error saving peers: %v", err)
	}
	defer os.Remove("peers_save_test.json")

	// Verify the saved file contents
	data, err := os.ReadFile("peers_save_test.json")
	if err != nil {
		t.Fatalf("Error reading saved file: %v", err)
	}

	var savedPeers []Peer
	err = json.Unmarshal(data, &savedPeers)
	if err != nil {
		t.Fatalf("Error unmarshalling saved peers: %v", err)
	}

	// Compare saved peers with original
	if len(savedPeers) != len(peers) {
		t.Errorf("Expected %d peers, got %d", len(peers), len(savedPeers))
	}
	for i, peer := range peers {
		if peer != savedPeers[i] {
			t.Errorf("Mismatch in saved peer at index %d: expected %v, got %v", i, peer, savedPeers[i])
		}
	}
}

func TestSavePeers_Error(t *testing.T) {
	// Save to an invalid directory to trigger an error
	peers := []Peer{
		{Host: "127.0.0.1", Port: "6001"},
	}

	err := SavePeers("/invalid_path/peers.json", peers)
	if err == nil {
		t.Error("Expected error for invalid save path, but got none")
	}
}
