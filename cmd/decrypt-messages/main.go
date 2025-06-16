package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"text/tabwriter"
	"time"

	"blockhead.consulting/internal/storage/git"
)

func main() {
	var (
		repoPath = flag.String("repo", "data/messages", "Path to messages repository")
		keyHex   = flag.String("key", "", "Encryption key in hex format (64 chars)")
		msgID    = flag.String("id", "", "Message ID to decrypt (optional, shows all if empty)")
		status   = flag.String("status", "", "Filter by status: new, read, replied, closed")
	)
	
	flag.Parse()
	
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
		PushOnWrite:   false,
	}
	
	// Create logger
	logger := log.New(os.Stdout, "[decrypt] ", log.LstdFlags)
	
	// Create storage service
	storage, err := git.NewService(config, logger, nil)
	if err != nil {
		log.Fatalf("Failed to create storage service: %v", err)
	}
	
	ctx := context.Background()
	
	// If specific message ID requested
	if *msgID != "" {
		msg, err := storage.GetMessage(ctx, *msgID)
		if err != nil {
			log.Fatalf("Failed to get message: %v", err)
		}
		
		printMessage(msg)
		return
	}
	
	// List all messages
	messages, err := storage.ListMessages(ctx, *status)
	if err != nil {
		log.Fatalf("Failed to list messages: %v", err)
	}
	
	if len(messages) == 0 {
		fmt.Println("No messages found")
		return
	}
	
	// Print in table format
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDate\tName\tEmail\tStatus\tSubject")
	fmt.Fprintln(w, "---\t----\t----\t-----\t------\t-------")
	
	for _, msg := range messages {
		subject := msg.Message
		if len(subject) > 50 {
			subject = subject[:47] + "..."
		}
		
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			msg.ID,
			msg.Timestamp.Format("2006-01-02 15:04"),
			msg.Name,
			msg.Email,
			msg.Status,
			subject,
		)
	}
	w.Flush()
	
	fmt.Printf("\nTotal messages: %d\n", len(messages))
	fmt.Println("\nTo view a specific message, use: -id <message-id>")
}

func printMessage(msg *git.Message) {
	fmt.Println("=====================================")
	fmt.Printf("Message ID: %s\n", msg.ID)
	fmt.Printf("Status: %s\n", msg.Status)
	fmt.Printf("Date: %s\n", msg.Timestamp.Format(time.RFC3339))
	fmt.Println("-------------------------------------")
	fmt.Printf("From: %s <%s>\n", msg.Name, msg.Email)
	if msg.Company != "" {
		fmt.Printf("Company: %s\n", msg.Company)
	}
	fmt.Printf("IP: %s\n", msg.IP)
	fmt.Printf("User Agent: %s\n", msg.UserAgent)
	fmt.Println("-------------------------------------")
	fmt.Println("Message:")
	fmt.Println(msg.Message)
	fmt.Println("=====================================")
}