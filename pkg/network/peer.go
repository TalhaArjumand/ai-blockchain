package network

import (
	"encoding/json"
	"os"
)

// Peer represents a peer's address
type Peer struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

// LoadPeers loads peers from a JSON file
func LoadPeers(filename string) ([]Peer, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var peers []Peer
	jsonParser := json.NewDecoder(file)
	err = jsonParser.Decode(&peers)
	return peers, err
}

// SavePeers saves the updated peer list to the JSON file
func SavePeers(filename string, peers []Peer) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	jsonEncoder := json.NewEncoder(file)
	return jsonEncoder.Encode(peers)
}
