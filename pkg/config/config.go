package config

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"strconv"
	"strings"
)

type NetworkConfig struct {
	Port string `json:"port"`
}

type Config struct {
	NetworkPort            int           `json:"networkPort"`
	MiningDifficultyTarget string        `json:"miningDifficultyTarget"`
	IPFSGatewayURL         string        `json:"ipfsGatewayURL"`
	DataDir                string        `json:"dataDir"`
	MaxBlockTransactions   int           `json:"maxBlockTransactions"`
	VMExecutionTimeout     int           `json:"vmExecutionTimeout"`
	DatasetHash            string        `json:"datasetHash"`
	AlgorithmHash          string        `json:"algorithmHash"`
	Network                NetworkConfig `json:"network"`
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

func ParseCSVToJSON(data []byte) ([]byte, error) {
	reader := csv.NewReader(strings.NewReader(string(data)))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var parsedData [][]float64
	for _, record := range records[1:] { // Skip the header
		var row []float64
		for _, val := range record {
			floatVal, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return nil, err
			}
			row = append(row, floatVal)
		}
		parsedData = append(parsedData, row)
	}

	return json.Marshal(parsedData)
}
