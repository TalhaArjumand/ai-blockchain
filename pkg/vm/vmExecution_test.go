package kmeans

import (
	"encoding/json"
	"testing"
)

func TestRunVM_KMeans_ValidInput(t *testing.T) {
	data := [][]float64{
		{1.0, 2.0},
		{1.5, 1.8},
		{5.0, 8.0},
	}

	// Serialize input data
	dataBytes, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to serialize input data: %v", err)
	}

	// Run K-Means using VM
	output, err := RunVM(nil, dataBytes) // `nil` algorithm since we call K-Means directly
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}

	t.Logf("Output from VM: %s", string(output))
}

func TestRunVM_KMeans_LargeDataset(t *testing.T) {
	data := make([][]float64, 1000)
	for i := range data {
		data[i] = []float64{float64(i), float64(i * 2)}
	}

	// Serialize input data
	dataBytes, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to serialize input data: %v", err)
	}

	// Run K-Means using VM
	output, err := RunVM(nil, dataBytes)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}

	t.Logf("Output from VM with large dataset: %s", string(output))
}

func TestRunVM_KMeans_EmptyDataset(t *testing.T) {
	data := [][]float64{} // Empty dataset

	// Serialize input data
	dataBytes, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to serialize input data: %v", err)
	}

	// Run K-Means using VM
	output, err := RunVM(nil, dataBytes)
	if err == nil {
		t.Fatalf("Expected an error for empty dataset, but got none. Output: %s", string(output))
	}

	t.Logf("Expected error received: %v", err)
}

func TestRunVM_KMeans_SinglePoint(t *testing.T) {
	data := [][]float64{
		{2.0, 3.0},
	}

	// Serialize input data
	dataBytes, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to serialize input data: %v", err)
	}

	// Run K-Means using VM
	output, err := RunVM(nil, dataBytes)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}

	t.Logf("Output from VM with single data point: %s", string(output))
}

func TestRunVM_KMeans_InvalidData(t *testing.T) {
	data := "InvalidData" // Non-JSON input

	// Convert to bytes
	dataBytes := []byte(data)

	// Run K-Means using VM
	output, err := RunVM(nil, dataBytes)
	if err == nil {
		t.Fatalf("Expected an error for invalid input data, but got none. Output: %s", string(output))
	}

	t.Logf("Expected error received for invalid input: %v", err)
}

func TestRunVM_KMeans_HighClusters(t *testing.T) {
	data := [][]float64{
		{1.0, 2.0},
		{3.0, 4.0},
	}

	// Serialize input data
	dataBytes, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to serialize input data: %v", err)
	}

	// Run K-Means with a higher number of clusters than data points
	output, err := RunVM(nil, dataBytes)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}

	t.Logf("Output from VM with high cluster count: %s", string(output))
}
