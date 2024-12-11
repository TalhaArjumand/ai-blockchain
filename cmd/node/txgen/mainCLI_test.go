package main

import (
	"testing"

	"github.com/TalhaArjumand/ai-blockchain/pkg/network"
)

// Global variables for mocking
var mockBroadcastResults []string
var mockBroadcastCalled bool

// Mock implementation of BroadcastTransaction
func mockBroadcastTransaction(tx network.TxMessage, peers []string) {
	mockBroadcastCalled = true
	mockBroadcastResults = peers
}

func TestRunCLI(t *testing.T) {
	tests := []struct {
		name          string
		dataHash      string
		algoHash      string
		peers         string
		expectFail    bool
		expectedPeers []string
	}{
		// Valid Inputs
		{
			name:          "Valid inputs",
			dataHash:      "QmValidDataHash",
			algoHash:      "QmValidAlgoHash",
			peers:         "localhost:6001,localhost:6002",
			expectFail:    false,
			expectedPeers: []string{"localhost:6001", "localhost:6002"},
		},
		{
			name:          "Single peer in list",
			dataHash:      "QmSingleDataHash",
			algoHash:      "QmSingleAlgoHash",
			peers:         "localhost:6001",
			expectFail:    false,
			expectedPeers: []string{"localhost:6001"},
		},
		{
			name:          "Peers with extra whitespace",
			dataHash:      "QmWhitespaceDataHash",
			algoHash:      "QmWhitespaceAlgoHash",
			peers:         " localhost:6001 , localhost:6002 ",
			expectFail:    false,
			expectedPeers: []string{"localhost:6001", "localhost:6002"},
		},
		// Missing Inputs
		{
			name:       "Missing dataHash",
			dataHash:   "",
			algoHash:   "QmValidAlgoHash",
			peers:      "localhost:6001,localhost:6002",
			expectFail: true,
		},
		{
			name:       "Missing algoHash",
			dataHash:   "QmValidDataHash",
			algoHash:   "",
			peers:      "localhost:6001,localhost:6002",
			expectFail: true,
		},
		{
			name:       "Missing peers",
			dataHash:   "QmValidDataHash",
			algoHash:   "QmValidAlgoHash",
			peers:      "",
			expectFail: true,
		},
		// Invalid Inputs
		{
			name:       "Invalid peer address",
			dataHash:   "QmValidDataHash",
			algoHash:   "QmValidAlgoHash",
			peers:      "invalid_peer",
			expectFail: true,
		},
		// Edge Cases
		{
			name:          "Duplicate peers in list",
			dataHash:      "QmDuplicateDataHash",
			algoHash:      "QmDuplicateAlgoHash",
			peers:         "localhost:6001,localhost:6001",
			expectFail:    false,
			expectedPeers: []string{"localhost:6001", "localhost:6001"},
		},
		{
			name:          "Large dataHash and algoHash",
			dataHash:      "Qm" + generateLargeString(256),
			algoHash:      "Qm" + generateLargeString(256),
			peers:         "localhost:6001,localhost:6002",
			expectFail:    false,
			expectedPeers: []string{"localhost:6001", "localhost:6002"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Reset global variables before each test
			mockBroadcastCalled = false
			mockBroadcastResults = nil

			// Run the CLI logic
			err := runCLI(tc.dataHash, tc.algoHash, tc.peers, mockBroadcastTransaction)

			// Check expectations
			if tc.expectFail {
				if err == nil {
					t.Errorf("expected failure but got success")
				}
			} else {
				if err != nil {
					t.Errorf("expected success but got error: %v", err)
				}
				if !mockBroadcastCalled {
					t.Error("expected BroadcastTransaction to be called but it was not")
				}
				if len(mockBroadcastResults) != len(tc.expectedPeers) {
					t.Errorf("expected peers: %v, got: %v", tc.expectedPeers, mockBroadcastResults)
				}
			}
		})
	}
}

// Helper function to generate a large string
func generateLargeString(size int) string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var result string
	for i := 0; i < size; i++ {
		result += string(chars[i%len(chars)])
	}
	return result
}
