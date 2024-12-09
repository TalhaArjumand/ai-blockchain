package ipfs

import (
	"bytes"
	"os"
	"testing"
)

func TestFetchAndUpload(t *testing.T) {
	client := NewIPFSClient("localhost:5001")

	// Test uploading raw data
	data := []byte("Sample Data for IPFS")
	hash, err := client.UploadData(data)
	if err != nil {
		t.Fatalf("Failed to upload data to IPFS. Error: %v", err)
	}

	t.Logf("Uploaded data hash: %s", hash)

	// Test fetching the same data
	fetchedData, err := client.FetchData(hash)
	if err != nil {
		t.Fatalf("Failed to fetch data from IPFS. Error: %v", err)
	}

	if string(fetchedData) != string(data) {
		t.Fatalf("Fetched data does not match original: %s != %s", string(fetchedData), string(data))
	}
}

func TestFetchAlgorithm(t *testing.T) {
	client := NewIPFSClient("localhost:5001")

	// Test uploading algorithm data
	algorithmData := []byte("Algorithm Data for IPFS")
	hash, err := client.UploadData(algorithmData)
	if err != nil {
		t.Fatalf("Failed to upload algorithm data to IPFS: %v", err)
	}

	// Test fetching the algorithm data
	fetchedData, err := client.FetchAlgorithm(hash)
	if err != nil {
		t.Fatalf("Failed to fetch algorithm data from IPFS: %v", err)
	}

	if string(fetchedData) != string(algorithmData) {
		t.Fatalf("Fetched algorithm data does not match original: %s != %s", string(fetchedData), string(algorithmData))
	}
}

func TestUploadFile(t *testing.T) {
	client := NewIPFSClient("localhost:5001")

	// Create a temporary file for testing
	filepath := "./test_file.txt"
	content := []byte("File content for IPFS")
	err := os.WriteFile(filepath, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(filepath)

	// Test uploading the file
	hash, err := client.UploadFile(filepath)
	if err != nil {
		t.Fatalf("Failed to upload file to IPFS: %v", err)
	}

	// Test fetching the file content
	fetchedData, err := client.FetchData(hash)
	if err != nil {
		t.Fatalf("Failed to fetch file from IPFS: %v", err)
	}

	if string(fetchedData) != string(content) {
		t.Fatalf("Fetched file content does not match original: %s != %s", string(fetchedData), string(content))
	}
}

func TestFetchInvalidHash(t *testing.T) {
	client := NewIPFSClient("localhost:5001")

	invalidHash := "InvalidHash12345"
	_, err := client.FetchData(invalidHash)
	if err == nil {
		t.Fatalf("Expected error when fetching data with invalid hash, but got none")
	}
}

func TestCacheUsage(t *testing.T) {
	client := NewIPFSClient("localhost:5001")

	// Test uploading raw data
	data := []byte("Sample Data for Caching")
	hash, err := client.UploadData(data)
	if err != nil {
		t.Fatalf("Failed to upload data to IPFS: %v", err)
	}

	// Test fetching data and confirm caching
	client.FetchData(hash) // First fetch populates the cache
	fetchedData, exists := client.cache.Load(hash)
	if !exists {
		t.Fatalf("Expected data to be cached, but it was not found")
	}

	if !bytes.Equal(fetchedData.([]byte), data) {
		t.Fatalf("Cached data does not match original: %s != %s", string(fetchedData.([]byte)), string(data))
	}
}

func TestPinAndUnpin(t *testing.T) {
	client := NewIPFSClient("localhost:5001")

	// Test uploading raw data
	data := []byte("Data for Pinning Test")
	hash, err := client.UploadData(data)
	if err != nil {
		t.Fatalf("Failed to upload data to IPFS: %v", err)
	}

	// Test pinning the data
	err = client.Pin(hash)
	if err != nil {
		t.Fatalf("Failed to pin data on IPFS: %v", err)
	}

	// Test unpinning the data
	err = client.Unpin(hash)
	if err != nil {
		t.Fatalf("Failed to unpin data from IPFS: %v", err)
	}
}

func TestUploadLargeData(t *testing.T) {
	client := NewIPFSClient("localhost:5001")

	// Generate a large data sample
	largeData := make([]byte, 10*1024*1024) // 10 MB
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	// Test uploading the large data
	hash, err := client.UploadData(largeData)
	if err != nil {
		t.Fatalf("Failed to upload large data to IPFS: %v", err)
	}

	// Test fetching the large data
	fetchedData, err := client.FetchData(hash)
	if err != nil {
		t.Fatalf("Failed to fetch large data from IPFS: %v", err)
	}

	if !bytes.Equal(fetchedData, largeData) {
		t.Fatalf("Fetched large data does not match original")
	}
}
