package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"

	"blockhead.consulting/internal/storage/git"
)

func main() {
	var (
		repoPath = flag.String("repo", "data/messages", "Path to messages repository")
		keyHex   = flag.String("key", "", "Encryption key in hex format (64 chars)")
		msgID    = flag.String("id", "", "Message ID to update (required)")
		status   = flag.String("status", "", "New status: new, read, replied, closed (required)")
		push     = flag.Bool("push", false, "Push changes to remote")
	)
	
	flag.Parse()
	
	if *msgID == "" || *status == "" {
		flag.Usage()
		os.Exit(1)
	}
	
	// Validate status
	validStatuses := map[string]bool{
		"new":     true,
		"read":    true,
		"replied": true,
		"closed":  true,
	}
	
	if !validStatuses[*status] {
		log.Fatalf("Invalid status. Must be one of: new, read, replied, closed")
	}
	
	if *keyHex == "" {
		// Try to get from environment
		*keyHex = os.Getenv("MESSAGE_ENCRYPTION_KEY")
		if *keyHex == "" {
			log.Fatal("Encryption key required: use -key flag or MESSAGE_ENCRYPTION_KEY env var")
		}
	}
	
	// Decode hex key
	key, err := hex.DecodeString(*keyHex)
	if err != nil {
		log.Fatalf("Invalid hex key: %v", err)
	}
	
	if len(key) != 32 {
		log.Fatal("Key must be 32 bytes (64 hex characters)")
	}
	
	// Create storage config
	config := git.StorageConfig{
		RepoPath:      *repoPath,
		EncryptionKey: string(key),
		PushOnWrite:   *push,
		RemoteURL:     "origin", // Assumes remote is already configured
		Branch:        "main",
		CommitAuthor:  "Message Status Tool",
		CommitEmail:   "noreply@blockhead.consulting",
	}
	
	// Create logger
	logger := log.New(os.Stdout, "[status] ", log.LstdFlags)
	
	// Create storage service
	storage, err := git.NewService(config, logger, nil)
	if err != nil {
		log.Fatalf("Failed to create storage service: %v", err)
	}
	
	ctx := context.Background()
	
	// Update status
	if err := storage.UpdateStatus(ctx, *msgID, *status); err != nil {
		log.Fatalf("Failed to update status: %v", err)
	}
	
	fmt.Printf("Updated message %s status to: %s\n", *msgID, *status)
	
	if *push {
		fmt.Println("Changes will be pushed to remote repository")
	}
}