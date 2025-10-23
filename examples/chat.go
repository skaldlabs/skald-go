//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/skald-org/skald-go"
)

func main() {
	// Create a new Skald client
	apiKey := os.Getenv("SKALD_API_KEY")
	if apiKey == "" {
		log.Fatal("SKALD_API_KEY environment variable not set")
	}

	client := skald.NewClient(apiKey)
	ctx := context.Background()

	// Example 1: Simple chat query
	fmt.Println("=== Simple Chat Query ===")
	chatResp, err := client.Chat(ctx, "What are the main features of Go?", nil)
	if err != nil {
		log.Fatalf("Failed to chat: %v", err)
	}

	fmt.Printf("Response: %s\n\n", chatResp.Response)

	// Example 2: Chat with filters
	fmt.Println("=== Chat with Filters ===")
	filteredChatResp, err := client.Chat(ctx, "Explain error handling", []skald.Filter{
		{
			Field:      "language",
			Operator:   skald.FilterOperatorEq,
			Value:      "go",
			FilterType: skald.FilterTypeCustomMetadata,
		},
	})

	if err != nil {
		log.Fatalf("Failed to chat with filters: %v", err)
	}

	fmt.Printf("Response: %s\n\n", filteredChatResp.Response)

	// Example 3: Streaming chat
	fmt.Println("=== Streaming Chat ===")
	eventChan, errChan := client.StreamedChat(ctx, "What is concurrency in Go?", nil)

	fmt.Print("Response: ")
	for event := range eventChan {
		if event.Type == "token" && event.Content != nil {
			fmt.Print(*event.Content)
		} else if event.Type == "done" {
			fmt.Println()
			break
		}
	}

	// Check for errors
	select {
	case err := <-errChan:
		if err != nil {
			log.Fatalf("Error during streaming: %v", err)
		}
	default:
	}

	fmt.Println()
}
