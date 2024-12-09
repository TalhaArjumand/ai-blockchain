package main

import (
	// Add this import for parsing command-line arguments

	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/TalhaArjumand/ai-blockchain/pkg/blockchain"
	"github.com/TalhaArjumand/ai-blockchain/pkg/config"
	"github.com/TalhaArjumand/ai-blockchain/pkg/ipfs"
	"github.com/TalhaArjumand/ai-blockchain/pkg/network"
)

var blockchainInstance *blockchain.Blockchain
var mempoolInstance = blockchain.NewMempool()
var knownPeers []string

func main() {

	// Initialize the blockchain instance
	blockchainInstance = &blockchain.Blockchain{
		Blocks: make(map[int]*blockchain.Block), // Initialize the Blocks map
	}
	// Step 0: Accept port as a command-line argument
	port := flag.String("port", "8081", "Port for the server to listen on")
	flag.Parse()

	// Step 1: Setup logger
	config.SetupLogger()
	config.Log.Infof("Starting AI Blockchain Node on port %s...", *port)

	// Step 2: Load configuration
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		config.Log.Fatalf("Failed to load config: %v", err)
	}
	config.Log.Infof("Loaded config: %+v", cfg)

	// Step 3: Initialize the blockchain
	chain := blockchain.NewBlockchain()

	// Step 4: Initialize the IPFS client
	client := ipfs.NewIPFSClient("localhost:5001")

	// Check if datasetHash is already provided
	var datasetHash, algoHash string

	if cfg.DatasetHash == "" {
		datasetHash, err = client.UploadFile("data.csv")
		if err != nil {
			config.Log.Fatalf("Failed to upload dataset to IPFS: %v", err)
		}
		config.Log.Infof("Dataset uploaded to IPFS with hash: %s", datasetHash)
	} else {
		datasetHash = cfg.DatasetHash
		config.Log.Infof("Using existing dataset hash: %s", datasetHash)
	}

	// Check if algorithmHash is already provided
	if cfg.AlgorithmHash == "" {
		algoHash, err = client.UploadFile("../../pkg/vm/kmeans.go")
		if err != nil {
			config.Log.Fatalf("Failed to upload algorithm to IPFS: %v", err)
		}
		config.Log.Infof("Algorithm uploaded to IPFS with hash: %s", algoHash)
	} else {
		algoHash = cfg.AlgorithmHash
		config.Log.Infof("Using existing algorithm hash: %s", algoHash)
	}

	// Step 7: Create a transaction using IPFS hashes
	tx := &blockchain.Transaction{
		DataHash:      datasetHash,
		AlgorithmHash: algoHash,
		Metadata:      "Dataset and algorithm for K-Means",
	}
	tx.GenerateTxID()
	config.Log.Infof("Generated transaction: %+v", tx)

	// Step 8: Create a block and add the transaction
	block := &blockchain.Block{
		Transactions: []blockchain.Transaction{*tx},
	}
	block.ComputeMerkleRoot()
	config.Log.Infof("Generated block with Merkle Root: %x", block.Header.MerkleRoot)

	// Step 9: Add the block to the blockchain
	chain.AddBlock(block)
	config.Log.Infof("Added block to blockchain. Current state: %+v", chain.Blocks)

	// Step 10: Persist the blockchain
	err = chain.Persist()
	if err != nil {
		config.Log.Fatalf("Failed to persist blockchain: %v", err)
	}
	config.Log.Info("Blockchain persisted to disk.")

	// Step 11: Initialize Network Layer
	config.Log.Infof("Starting server on port %s", *port)
	go network.StartServer(*port, handleMessage)

	// Step 12: Connect to known peers
	peers, err := network.LoadPeers("peers.json")
	if err != nil {
		config.Log.Fatalf("Failed to load peers: %v", err)
	}

	// Populate knownPeers with peer addresses
	for _, peer := range peers {
		knownPeers = append(knownPeers, peer.Host+":"+peer.Port)
	}

	for _, peer := range peers {
		go func(peer network.Peer) {
			for {
				message, _ := network.SerializeMessage(map[string]string{"type": "handshake", "version": "1.0"})
				err := network.SendMessage(peer.Host+":"+peer.Port, message)
				if err == nil {
					config.Log.Infof("Connected to peer: %s:%s", peer.Host, peer.Port)
					break
				}
				config.Log.Warnf("Retrying connection to peer: %s:%s. Error: %v", peer.Host, peer.Port, err)
				time.Sleep(3 * time.Second) // Retry after 3 seconds
			}
		}(peer)
	}

	// Keep the application running
	select {}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////
//
//
//
//////////////////////////////////////////////////////////////////////////////////////////////////////////////

func convertTransactionsToTxMessages(transactions []blockchain.Transaction) []network.TxMessage {
	var txMessages []network.TxMessage
	for _, tx := range transactions {
		txMessage := network.TxMessage{
			TxID:      string(tx.TxID),
			DataHash:  tx.DataHash,
			AlgoHash:  tx.AlgorithmHash,
			Metadata:  tx.Metadata,
			Timestamp: tx.Timestamp,
		}
		txMessages = append(txMessages, txMessage)
	}
	return txMessages
}

func handleMessage(message []byte) {
	// Attempt to deserialize the message
	var msg map[string]interface{}
	err := json.Unmarshal(message, &msg)
	if err != nil {
		fmt.Printf("Failed to deserialize message: %v\n", err)
		return
	}

	// Check if the "type" field is present
	messageType, ok := msg["type"].(string)
	if !ok {
		fmt.Println("Message missing 'type' field. Full message:", string(message))
		return
	}

	// Handle known message types
	switch messageType {
	case "handshake":
		// Example: Log or process handshake messages
		version, _ := msg["version"].(string)
		fmt.Printf("Handshake received from version: %s\n", version)
	case "TxMessage":
		var tx network.TxMessage
		if err := json.Unmarshal(message, &tx); err != nil {
			log.Printf("Error unmarshalling TxMessage: %v", err)
			return
		}
		log.Printf("Received transaction: %+v", tx)
	case "BlockMessage":
		var block network.BlockMessage
		if err := json.Unmarshal(message, &block); err != nil {
			log.Printf("Error unmarshalling BlockMessage: %v", err)
			return
		}
		log.Printf("Received block: %+v", block)
	case "GetBlocksMessage":
		var request network.GetBlocksMessage
		if err := json.Unmarshal(message, &request); err != nil {
			log.Printf("Error unmarshalling GetBlocksMessage: %v", err)
			return
		}
		log.Printf("Received request for blocks from height %d to %d from node %s", request.StartHeight, request.EndHeight, request.RequestingNode)

		// Fetch blocks using the blockchain instance
		blocks, err := blockchain.FetchBlocks(request.StartHeight, request.EndHeight, blockchainInstance)
		if err != nil {
			log.Printf("Error fetching blocks: %v", err)
			return
		}

		// Serialize the blocks and send them back to the requesting node
		response := network.BlocksMessage{
			Blocks: blocks,
		}
		responseMessage, err := network.SerializeMessage(response)
		if err != nil {
			log.Printf("Error serializing BlocksMessage: %v", err)
			return
		}

		// Send response back to the requesting node
		err = network.SendMessage(request.RequestingNode, responseMessage)
		if err != nil {
			log.Printf("Error sending blocks to node %s: %v", request.RequestingNode, err)
		}
	case "InvMessage":
		var inventory network.InvMessage
		if err := json.Unmarshal(message, &inventory); err != nil {
			log.Printf("Error unmarshalling InvMessage: %v", err)
			return
		}
		log.Printf("Received inventory: %+v", inventory)

		for _, peerAddr := range knownPeers { // knownPeers is a slice of peer addresses
			for _, hash := range inventory.Hashes {
				switch inventory.Type {
				case "block":
					if !blockchainInstance.HasBlock(hash) {
						log.Printf("Requesting missing block with hash: %s", hash)
						request := network.GetDataMessage{
							Type: "block",
							Hash: hash,
						}
						requestMessage, _ := network.SerializeMessage(request)
						network.SendMessage(peerAddr, requestMessage)
					}
				case "transaction":
					if !mempoolInstance.HasTransaction(hash) { // Ensure mempoolInstance is accessible
						log.Printf("Requesting missing transaction with hash: %s", hash)
						request := network.GetDataMessage{
							Type: "transaction",
							Hash: hash,
						}
						requestMessage, _ := network.SerializeMessage(request)
						network.SendMessage(peerAddr, requestMessage)
					}
				default:
					log.Printf("Unknown inventory type: %s", inventory.Type)
				}
			}
		}

	case "GetDataMessage":
		var request network.GetDataMessage
		if err := json.Unmarshal(message, &request); err != nil {
			log.Printf("Error unmarshalling GetDataMessage: %v", err)
			return
		}
		log.Printf("Received request for data: %+v", request)

		if request.Type == "block" {
			block := blockchainInstance.GetBlock(request.Hash)
			if block != nil {
				response := network.BlockMessage{
					BlockID:      string(block.Header.MerkleRoot),
					MerkleRoot:   string(block.Header.MerkleRoot),
					PreviousHash: string(block.Header.PreviousHash),
					Transactions: convertTransactionsToTxMessages(block.Transactions),
					Timestamp:    block.Header.Timestamp,
				}
				responseMessage, err := network.SerializeMessage(response)
				if err != nil {
					log.Printf("Error serializing BlockMessage: %v", err)
					return
				}
				err = network.SendMessage(request.PeerAddress, responseMessage)
				if err != nil {
					log.Printf("Error sending block: %v", err)
				}
			} else {
				log.Printf("Block not found for hash: %s", request.Hash)
			}
		}

		if request.Type == "transaction" {
			tx := mempoolInstance.GetTransaction(request.Hash)
			if tx != nil {
				response := network.TxMessage{
					TxID:      string(tx.TxID),
					DataHash:  tx.DataHash,
					AlgoHash:  tx.AlgorithmHash,
					Metadata:  tx.Metadata,
					Timestamp: tx.Timestamp,
				}
				responseMessage, err := network.SerializeMessage(response)
				if err != nil {
					log.Printf("Error serializing TxMessage: %v", err)
					return
				}
				err = network.SendMessage(request.PeerAddress, responseMessage)
				if err != nil {
					log.Printf("Error sending transaction: %v", err)
				}
			} else {
				log.Printf("Transaction not found for hash: %s", request.Hash)
			}
		}

	default:
		// Log and gracefully handle unsupported message types
		fmt.Printf("Unknown message type: %s. Full message: %s\n", messageType, string(message))
	}
}
