package vm

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/TalhaArjumand/ai-blockchain/pkg/config"
	"github.com/TalhaArjumand/ai-blockchain/pkg/kmeans"
)

func RunVM(algorithm []byte, data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("data cannot be empty")
	}

	// Preprocess data if it's in CSV format
	if strings.Contains(string(data), ",") {
		parsedData, err := config.ParseCSVToJSON(data)
		if err != nil {
			return nil, fmt.Errorf("failed to parse input data: %v", err)
		}
		data = parsedData
	}

	algoStr := string(algorithm)
	if strings.Contains(algoStr, "KMeans") {
		return RunKMeans(data)
	}

	return nil, errors.New("unsupported algorithm")
}

// Example of directly embedding K-Means execution
func RunKMeans(data []byte) ([]byte, error) {
	// Step 1: Deserialize data into a usable structure
	var input [][]float64
	err := json.Unmarshal(data, &input)
	if err != nil {
		return nil, fmt.Errorf("failed to parse input data: %v", err)
	}

	// Step 2: Validate input data
	if len(input) == 0 {
		return nil, fmt.Errorf("input data is empty")
	}

	// Step 3: Execute K-Means using the imported function
	k := 2                                        // Number of clusters
	maxIter := 10                                 // Maximum number of iterations
	centroids := kmeans.KMeans(input, k, maxIter) // Call imported K-Means function

	// Step 4: Serialize the output
	output, err := json.Marshal(centroids)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize K-Means output: %v", err)
	}

	// Step 5: Return serialized output
	return output, nil
}
