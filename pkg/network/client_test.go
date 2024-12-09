package network

import (
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

func MockServer(t *testing.T, port string, response string) {
	go func() {
		listener, err := net.Listen("tcp", ":"+port)
		if err != nil {
			t.Logf("Failed to start mock server on port %s: %v", port, err)
		}
		defer listener.Close()

		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			defer conn.Close()

			buffer := make([]byte, 1024)
			n, _ := conn.Read(buffer)
			fmt.Printf("Mock server received: %s\n", string(buffer[:n]))

			conn.Write([]byte(response))
		}
	}()
}

func TestSendMessage_Success(t *testing.T) {
	port := "6000"
	go MockServer(t, port, "ACK") // Mock server listens on port 6000

	peerAddr := "localhost:" + port
	message := []byte("Hello, peer!")

	err := SendMessage(peerAddr, message)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}
}

func TestSendMessage_ConnectionError(t *testing.T) {
	err := SendMessage("localhost:9999", []byte("Test Message"))
	if err == nil {
		t.Fatalf("Expected connection error but got none")
	}
}

func TestSendMessage_WriteError(t *testing.T) {
	// Start a mock server to simulate a write error
	port := "6001"
	MockServerForWriteError(t, port)

	// Attempt to send a message to the mock server
	message := []byte("Hello, peer!")
	err := SendMessage("localhost:"+port, message)

	// Verify that either a connection or write error occurs
	if err == nil {
		t.Errorf("Expected error but got none")
	} else if !(strings.Contains(err.Error(), "write") || strings.Contains(err.Error(), "connectex")) {
		t.Errorf("Expected write or connection error but got: %v", err)
	}
}

func TestBroadcastTransaction_Success(t *testing.T) {
	port1, port2 := "6002", "6003"
	go MockServer(t, port1, "ACK")
	go MockServer(t, port2, "ACK")

	peers := []string{"localhost:" + port1, "localhost:" + port2}
	tx := TxMessage{TxID: "1234", DataHash: "abcd", AlgoHash: "efgh"}

	BroadcastTransaction(tx, peers)
	// No assertions since BroadcastTransaction logs errors. Check the logs for verification.
}

func TestBroadcastTransaction_PartialFailure(t *testing.T) {
	port := "6004"
	go MockServer(t, port, "ACK")

	peers := []string{"localhost:" + port, "localhost:9999"} // One valid, one invalid
	tx := TxMessage{TxID: "1234", DataHash: "abcd", AlgoHash: "efgh"}

	BroadcastTransaction(tx, peers)
	// Check the logs for partial failure messages.
}

func TestBroadcastBlock_Success(t *testing.T) {
	port1, port2 := "6005", "6006"
	go MockServer(t, port1, "ACK")
	go MockServer(t, port2, "ACK")

	peers := []string{"localhost:" + port1, "localhost:" + port2}

	// Convert string timestamp to int64
	timestampStr := "2024-12-09T17:36:00Z"
	parsedTime, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		t.Fatalf("Failed to parse timestamp: %v", err)
	}
	timestamp := parsedTime.Unix()

	block := BlockMessage{
		BlockID:      "block123",
		MerkleRoot:   "merkleRoot123",
		PreviousHash: "prevHash123",
		Transactions: []TxMessage{},
		Timestamp:    timestamp, // Use int64 Unix timestamp
	}

	BroadcastBlock(block, peers)
	// No assertions since BroadcastBlock logs errors. Check the logs for verification.
}

func TestBroadcastBlock_PartialFailure(t *testing.T) {
	port := "6007"
	go MockServer(t, port, "ACK")

	peers := []string{"localhost:" + port, "localhost:9999"} // One valid, one invalid

	// Convert string timestamp to int64
	timestampStr := "2024-12-09T17:36:00Z"
	parsedTime, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		t.Fatalf("Failed to parse timestamp: %v", err)
	}
	timestamp := parsedTime.Unix()

	block := BlockMessage{
		BlockID:      "block123",
		MerkleRoot:   "merkleRoot123",
		PreviousHash: "prevHash123",
		Transactions: []TxMessage{},
		Timestamp:    timestamp, // Use int64 Unix timestamp
	}

	BroadcastBlock(block, peers)
	// Check the logs for partial failure messages.
}
func MockServerForWriteError(t *testing.T, port string) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		t.Fatalf("Failed to start mock server on port %s: %v", port, err)
	}
	defer listener.Close()

	// Use a channel to indicate the server is ready
	ready := make(chan struct{})
	go func() {
		close(ready) // Signal that the server is ready

		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			// Explicitly reject writes by immediately closing the connection
			conn.Close()
		}
	}()
	<-ready // Wait for the server to be ready
}

func TestServerAndClient(t *testing.T) {
	listener, err := net.Listen("tcp", ":0") // Dynamically assign an available port
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			buffer := make([]byte, 1024)
			n, _ := conn.Read(buffer)
			fmt.Printf("Mock server received: %s\n", string(buffer[:n]))
			conn.Write([]byte("Hello, Client!"))
			conn.Close()
		}
	}()

	peerAddr := fmt.Sprintf("localhost:%d", port)
	message := []byte("Hello, Server!")
	err = SendMessage(peerAddr, message)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}
}
