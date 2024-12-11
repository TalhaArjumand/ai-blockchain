package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/TalhaArjumand/ai-blockchain/pkg/blockchain"
	"github.com/TalhaArjumand/ai-blockchain/pkg/network"
)

func main() {
	// CLI Flags
	dataHash := flag.String("dataHash", "", "IPFS data hash")
	algoHash := flag.String("algoHash", "", "IPFS algorithm hash")
	peers := flag.String("peers", "localhost:6001,localhost:6002", "Comma-separated list of peers")
	flag.Parse()

	err := runCLI(*dataHash, *algoHash, *peers, network.BroadcastTransaction)
	if err != nil {
		log.Fatal(err)
	}
}

type BroadcastFunc func(tx network.TxMessage, peers []string)

func runCLI(dataHash, algoHash, peers string, broadcast func(tx network.TxMessage, peers []string)) error {
	if dataHash == "" || algoHash == "" {
		return errors.New("dataHash and algoHash are required")
	}

	// Create a Transaction
	tx := blockchain.Transaction{
		DataHash:      dataHash,
		AlgorithmHash: algoHash,
		Timestamp:     time.Now().Unix(),
		Metadata:      "Generated via CLI",
	}
	tx.GenerateTxID()

	// Encapsulate the transaction in a TxMessage
	txMessage := network.TxMessage{
		Type:      "transaction",
		TxID:      fmt.Sprintf("%x", tx.TxID),
		DataHash:  tx.DataHash,
		AlgoHash:  tx.AlgorithmHash,
		Metadata:  tx.Metadata,
		Timestamp: tx.Timestamp,
	}

	// Parse peers and validate addresses
	peerList := strings.Split(peers, ",")
	if len(peerList) == 0 || (len(peerList) == 1 && peerList[0] == "") {
		return errors.New("invalid or empty peer list")
	}

	for _, peer := range peerList {
		if err := validatePeerAddress(peer); err != nil {
			return fmt.Errorf("invalid peer address: %s, error: %v", peer, err)
		}
	}

	// Display the TxMessage details
	fmt.Printf("TxMessage Created:\n%+v\n", txMessage)

	// Broadcast the transaction
	broadcast(txMessage, peerList)
	return nil
}

// validatePeerAddress checks if the peer address is in the format host:port
func validatePeerAddress(peer string) error {
	_, _, err := net.SplitHostPort(strings.TrimSpace(peer))
	return err
}
