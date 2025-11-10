//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/skaldlabs/skald-go"
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
	chatResp, err := client.Chat(ctx, skald.ChatParams{
		Query: "What are the main features of Go?",
	})
	if err != nil {
		log.Fatalf("Failed to chat: %v", err)
	}

	fmt.Printf("Response: %s\n\n", chatResp)

	// Example 2: Chat with filters
	fmt.Println("=== Chat with Filters ===")
	filteredChatResp, err := client.Chat(ctx, skald.ChatParams{
		Query: "Explain error handling",
		Filters: []skald.Filter{
			{
				Field:      "language",
				Operator:   skald.FilterOperatorEq,
				Value:      "go",
				FilterType: skald.FilterTypeCustomMetadata,
			},
		},
	})

	if err != nil {
		log.Fatalf("Failed to chat with filters: %v", err)
	}

	fmt.Printf("Response: %s\n\n", filteredChatResp)

	// Example 3: Streaming chat
	fmt.Println("=== Streaming Chat ===")
	eventChan, errChan := client.StreamedChat(ctx, skald.ChatParams{
		Query: "What is concurrency in Go?",
	})

	fmt.Print("Response: ")
	for event := range eventChan {
		if event.Type == "token" && event.Content != nil {
			fmt.Print(*event.Content)
		} else if event.Type == "done" {
			fmt.Println()
			break
		}
	}

	// Example 4: Chat with system prompt
	fmt.Println("=== Chat with System Prompt ===")
	systemPrompt := "You are a helpful assistant that answers questions about Go."
	systemChatResp, err := client.Chat(ctx, skald.ChatParams{
		Query:        "What are the main features of Go?",
		SystemPrompt: systemPrompt,
	})
	if err != nil {
		log.Fatalf("Failed to chat with system prompt: %v", err)
	}

	fmt.Printf("Response: %s\n\n", systemChatResp)

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
