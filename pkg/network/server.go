package network

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

// StartServer initializes the TCP server for the node
func StartServer(port string, messageHandler func([]byte)) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	fmt.Println("Server listening on port", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// Handle each connection in a separate goroutine
		go handleConnection(conn, messageHandler)
	}
}

func handleConnection(conn net.Conn, messageHandler func([]byte)) {
	defer conn.Close()
	message, _ := bufio.NewReader(conn).ReadString('\n')
	fmt.Println("Received:", strings.TrimSpace(message))

	// Dispatch the message to the handler
	messageHandler([]byte(message))
}
