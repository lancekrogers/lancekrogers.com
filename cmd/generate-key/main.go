package main

import (
	"encoding/hex"
	"fmt"
	"log"

	"blockhead.consulting/internal/storage/git"
)

func main() {
	// Generate a new encryption key
	key, err := git.GenerateKey()
	if err != nil {
		log.Fatalf("Failed to generate key: %v", err)
	}
	
	// Display in hex format
	keyHex := hex.EncodeToString(key)
	
	fmt.Println("Generated new encryption key:")
	fmt.Println("============================")
	fmt.Printf("Hex format: %s\n", keyHex)
	fmt.Println("\nUsage:")
	fmt.Println("1. Add to .env file:")
	fmt.Printf("   MESSAGE_ENCRYPTION_KEY=%s\n", keyHex)
	fmt.Println("\n2. Or use with decrypt tool:")
	fmt.Printf("   go run cmd/decrypt-messages/main.go -key %s\n", keyHex)
	fmt.Println("\nIMPORTANT: Store this key securely! You cannot decrypt messages without it.")
}