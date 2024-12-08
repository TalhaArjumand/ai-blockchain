package main

import (
	"fmt"
	"log"

	"ai-blockchain/pkg/config"
)

func main() {
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Printf("Loaded config: %+v\n", cfg)
}
