package network

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
)

// SendMessage sends a message to a peer
func SendMessage(peerAddr string, message []byte) error {
	conn, err := net.Dial("tcp", peerAddr)
	if err != nil {
		return fmt.Errorf("error connecting to peer %s: %w", peerAddr, err)
	}
	defer conn.Close()

	_, err = conn.Write(message)
	if err != nil {
		return fmt.Errorf("error sending message to peer %s: %w", peerAddr, err)
	}

	fmt.Println("Message sent to", peerAddr)
	return nil
}

func BroadcastTransaction(tx TxMessage, peers []string) {
	message, err := json.Marshal(tx)
	if err != nil {
		log.Printf("Error marshalling transaction: %v", err)
		return
	}

	for _, peer := range peers {
		log.Printf("Broadcasting transaction to peer: %s", peer) // Add this log
		err := SendMessage(peer, message)
		if err != nil {
			log.Printf("Failed to send transaction to peer %s: %v", peer, err)
		} else {
			log.Printf("Transaction sent to peer %s", peer)
		}
	}
}

// BroadcastBlock sends a mined block to all known peers
func BroadcastBlock(block BlockMessage, peers []string) {
	message, err := json.Marshal(block)
	if err != nil {
		log.Printf("Error marshalling block: %v", err)
		return
	}

	for _, peer := range peers {
		err := SendMessage(peer, message)
		if err != nil {
			log.Printf("Failed to send block to peer %s: %v", peer, err)
		} else {
			log.Printf("Block sent to peer %s", peer)
		}
	}
}
