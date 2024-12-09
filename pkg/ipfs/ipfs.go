package ipfs

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	shell "github.com/ipfs/go-ipfs-api"
)

type IPFSClient struct {
	shell *shell.Shell
	cache sync.Map
}

type IPFSConfig struct {
	GatewayURL string
	Timeout    time.Duration
	Retries    int
	Delay      time.Duration
}

type IPFSInterface interface {
	FetchData(hash string) ([]byte, error)
	FetchAlgorithm(hash string) ([]byte, error)
	FetchInputs(txID string) ([]byte, []byte, error) // Add this method
}

// MockIPFSClient simulates the behavior of the IPFSClient for testing
type MockIPFSClient struct {
	Valid bool
}

// FetchData for mock client
func (m *MockIPFSClient) FetchData(hash string) ([]byte, error) {
	if m.Valid {
		return []byte("mockData"), nil
	}
	return nil, errors.New("mock invalid data")
}

// FetchAlgorithm for mock client
func (m *MockIPFSClient) FetchAlgorithm(hash string) ([]byte, error) {
	if m.Valid {
		return []byte("mockAlgorithm"), nil
	}
	return nil, errors.New("mock invalid algorithm")
}

func (m *MockIPFSClient) FetchInputs(txID string) ([]byte, []byte, error) {
	if m.Valid {
		return []byte("mockData"), []byte("mockAlgo"), nil
	}
	return nil, nil, errors.New("invalid inputs")
}

func NewIPFSClient(gatewayURL string) *IPFSClient {
	sh := shell.NewShell(gatewayURL)
	if !sh.IsUp() {
		log.Fatalf("IPFS daemon at %s is unreachable. Ensure the daemon is running and accessible.", gatewayURL)
	}
	return &IPFSClient{
		shell: sh,
		cache: sync.Map{},
	}
}

// Initialize IPFS client with config
func NewIPFSClientWithConfig(config IPFSConfig) *IPFSClient {
	return &IPFSClient{
		shell: shell.NewShell(config.GatewayURL),
		cache: sync.Map{},
	}
}

// Fetch data with timeout and retries
func (client *IPFSClient) FetchData(hash string) ([]byte, error) {
	// Fetch from IPFS without direct context support
	reader, err := client.shell.Cat(hash)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	client.cache.Store(hash, data)
	return data, nil
}

func (client *IPFSClient) FetchDataWithTimeout(hash string, timeout time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resultChan := make(chan []byte)
	errChan := make(chan error)

	go func() {
		reader, err := client.shell.Cat(hash)
		if err != nil {
			errChan <- err
			return
		}
		defer reader.Close()

		data, err := ioutil.ReadAll(reader)
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- data
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case data := <-resultChan:
		client.cache.Store(hash, data)
		return data, nil
	case err := <-errChan:
		return nil, err
	}
}

// Fetch algorithm from IPFS
func (client *IPFSClient) FetchAlgorithm(hash string) ([]byte, error) {
	return client.FetchData(hash) // Same logic as FetchData
}

// Retry mechanism for fetching data
func (client *IPFSClient) FetchDataWithRetry(hash string, retries int, delay time.Duration) ([]byte, error) {
	var err error
	for i := 0; i < retries; i++ {
		data, err := client.FetchData(hash)
		if err == nil {
			return data, nil
		}
		time.Sleep(delay)
	}
	return nil, err
}

// Upload data to IPFS
func (client *IPFSClient) UploadData(data []byte) (string, error) {
	log.Println("Uploading data to IPFS...")
	hash, err := client.shell.Add(bytes.NewReader(data))
	if err != nil {
		log.Printf("Failed to upload data: %v", err)
		return "", err
	}
	log.Printf("Uploaded data with hash: %s", hash)
	return hash, nil
}

func (client *IPFSClient) UploadFile(filepath string) (string, error) {
	// Open the file
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Add the file to IPFS
	hash, err := client.shell.Add(file)
	if err != nil {
		return "", err
	}
	return hash, nil
}

// Pin data on IPFS
func (client *IPFSClient) Pin(hash string) error {
	return client.shell.Pin(hash)
}

// Unpin data from IPFS
// Unpin data from IPFS
func (client *IPFSClient) Unpin(hash string) error {
	return client.shell.Unpin(hash)
}
