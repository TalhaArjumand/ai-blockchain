package network

import (
	"fmt"
	"net"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

// Mock message handler for testing
func mockMessageHandler(messages *[]string, mu *sync.Mutex) func([]byte) {
	return func(data []byte) {
		mu.Lock()
		defer mu.Unlock()
		*messages = append(*messages, strings.TrimSpace(string(data)))
	}
}

// Test the StartServer function with a mock handler
func TestStartServer(t *testing.T) {
	port := "6008"
	messages := []string{}
	var mu sync.Mutex

	// Start the server in a separate goroutine
	go StartServer(port, mockMessageHandler(&messages, &mu))
	time.Sleep(1 * time.Second) // Allow the server to start

	// Connect to the server and send a message
	conn, err := net.Dial("tcp", "localhost:"+port)
	if err != nil {
		t.Fatalf("Error connecting to server: %v", err)
	}
	defer conn.Close()

	message := "Hello, Server!\n"
	_, err = conn.Write([]byte(message))
	if err != nil {
		t.Fatalf("Error sending message to server: %v", err)
	}

	// Wait for the message to be processed
	time.Sleep(1 * time.Second)

	// Verify that the message was received
	mu.Lock()
	defer mu.Unlock()
	if len(messages) != 1 || messages[0] != "Hello, Server!" {
		t.Errorf("Expected message 'Hello, Server!', but got: %+v", messages)
	}
}

// Test multiple connections and messages
func TestMultipleConnections(t *testing.T) {
	port := "6009"
	messages := []string{"Message 1", "Message 2", "Message 3"}
	receivedMessages := make(chan string, 3)
	errorChannel := make(chan error, len(messages))

	go StartServer(port, func(message []byte) {
		receivedMessages <- string(message)
	})

	time.Sleep(1 * time.Second) // Allow the server to start

	// Simulate multiple connections
	for _, msg := range messages {
		go func(m string) {
			conn, err := net.Dial("tcp", "localhost:"+port)
			if err != nil {
				errorChannel <- fmt.Errorf("error connecting to server: %v", err)
				return
			}
			defer conn.Close()

			_, writeErr := fmt.Fprintln(conn, m) // Send message
			if writeErr != nil {
				errorChannel <- fmt.Errorf("error writing to server: %v", writeErr)
			}
		}(msg)
	}

	// Collect errors from goroutines
	for i := 0; i < len(messages); i++ {
		select {
		case err := <-errorChannel:
			t.Errorf("Error during connection: %v", err)
		default:
			// No error, continue
		}
	}

	// Collect received messages
	var results []string
	for i := 0; i < len(messages); i++ {
		select {
		case msg := <-receivedMessages:
			results = append(results, strings.TrimSpace(msg))
		case <-time.After(2 * time.Second):
			t.Fatalf("Timed out waiting for messages")
		}
	}

	// Sort both expected and received messages for comparison
	sort.Strings(results)
	sort.Strings(messages)

	if !reflect.DeepEqual(results, messages) {
		t.Errorf("Expected messages %v, but got %v", messages, results)
	}
}

// Test server handling an empty message
func TestEmptyMessage(t *testing.T) {
	port := "6010"
	messages := []string{}
	var mu sync.Mutex

	// Start the server in a separate goroutine
	go StartServer(port, mockMessageHandler(&messages, &mu))
	time.Sleep(1 * time.Second) // Allow the server to start

	// Connect to the server and send an empty message
	conn, err := net.Dial("tcp", "localhost:"+port)
	if err != nil {
		t.Fatalf("Error connecting to server: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte("\n"))
	if err != nil {
		t.Fatalf("Error sending empty message to server: %v", err)
	}

	// Wait for the message to be processed
	time.Sleep(1 * time.Second)

	// Verify that the message was received as an empty string
	mu.Lock()
	defer mu.Unlock()
	if len(messages) != 1 || messages[0] != "" {
		t.Errorf("Expected empty message, but got: %+v", messages)
	}
}

// Test server failing to start on an invalid port
func TestServerStartFailure(t *testing.T) {
	invalidPort := "invalidPort"

	// Attempt to start the server with an invalid port
	go StartServer(invalidPort, func([]byte) {})
	time.Sleep(1 * time.Second)

	// Verify that the server did not start
	_, err := net.Dial("tcp", "localhost:"+invalidPort)
	if err == nil {
		t.Fatalf("Expected error connecting to invalid port, but connection was successful")
	}
}
