package main

import (
	"bytes"
	"log"
	"os"

	shell "github.com/ipfs/go-ipfs-api"
)

func main() {
	// Initialize IPFS client
	client := shell.NewShell("http://127.0.0.1:5001")

	// Test uploading raw data
	data := []byte("Test Data")
	hash, err := client.Add(bytes.NewReader(data))
	if err != nil {
		log.Fatalf("Failed to upload data to IPFS: %v", err)
	}
	log.Printf("Data uploaded successfully. Hash: %s", hash)

	// Test uploading a file
	filePath := "C:\\Users\\asas 1\\Desktop\\abc.txt"
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	fileHash, err := client.Add(file)
	if err != nil {
		log.Fatalf("Failed to upload file to IPFS: %v", err)
	}
	log.Printf("File uploaded successfully. Hash: %s", fileHash)
}
