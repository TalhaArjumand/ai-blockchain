package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	NetworkPort            int    `json:"networkPort"`
	MiningDifficultyTarget string `json:"miningDifficultyTarget"`
	IPFSGatewayURL         string `json:"ipfsGatewayURL"`
	DataDir                string `json:"dataDir"`
	MaxBlockTransactions   int    `json:"maxBlockTransactions"`
	VMExecutionTimeout     int    `json:"vmExecutionTimeout"`
}

func LoadConfig(filepath string) (*Config, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := &Config{}
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}
	return config, nil
}
