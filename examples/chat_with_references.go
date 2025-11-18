// Chat with References Example
//
// This example demonstrates how to use the references feature in the Skald SDK.
// When enabled, the chat API returns citation markers (e.g., [[1]], [[2]]) in the
// response along with a references object that maps these markers to the source memos.
//
// This is useful for:
// - Tracking which memos contributed to the answer
// - Providing source attribution in your application
// - Building trust with users by showing information sources
// - Linking back to original documents
//
// Prerequisites:
// - Set SKALD_API_KEY environment variable
// - Have some memos already created in your project
//
// Usage:
// go run examples/chat_with_references.go

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	skald "github.com/skaldlabs/skald-go"
)

func main() {
	apiKey := os.Getenv("SKALD_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: SKALD_API_KEY environment variable not set")
		os.Exit(1)
	}

	client := skald.NewClient(apiKey)
	ctx := context.Background()

	fmt.Println("=== Chat with References Example ===\n")

	// Example 1: Basic chat with references enabled
	fmt.Println("Example 1: Basic Chat with References\n")
	fmt.Println("Query: \"What are the key features of our product?\"\n")

	response, err := client.Chat(ctx, skald.ChatParams{
		Query: "What are the key features of our product?",
		RAGConfig: &skald.RAGConfig{
			References: &skald.ReferencesConfig{
				Enabled: true,
			},
		},
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Response: %s\n\n", response.Response)

	// Display references if available
	if response.References != nil && len(response.References) > 0 {
		fmt.Println("Sources:")
		for refNum, refData := range response.References {
			fmt.Printf("  [%s] %s\n", refNum, refData.MemoTitle)
			fmt.Printf("      UUID: %s\n", refData.MemoUUID)
		}
	} else {
		fmt.Println("No references found in the response.")
	}

	fmt.Println("\n" + strings.Repeat("=", 60) + "\n")

	// Example 2: Streaming chat with references
	fmt.Println("Example 2: Streaming Chat with References\n")
	fmt.Println("Query: \"Summarize the latest meeting notes\"\n")
	fmt.Println("Streaming response:")

	var fullResponse strings.Builder
	var chatReferences skald.References
	var chatID string

	eventChan, errChan := client.StreamedChat(ctx, skald.ChatParams{
		Query: "Summarize the latest meeting notes",
		RAGConfig: &skald.RAGConfig{
			References: &skald.ReferencesConfig{
				Enabled: true,
			},
		},
	})

	for event := range eventChan {
		switch event.Type {
		case "token":
			if event.Content != nil {
				fmt.Print(*event.Content)
				fullResponse.WriteString(*event.Content)
			}
		case "references":
			// References arrive as JSON in the content
			if event.Content != nil {
				if err := json.Unmarshal([]byte(*event.Content), &chatReferences); err != nil {
					// Try to use the References field directly if available
					chatReferences = event.References
				}
			} else {
				chatReferences = event.References
			}
		case "done":
			chatID = event.ChatID
		}
	}

	// Check for errors
	select {
	case err := <-errChan:
		if err != nil {
			fmt.Printf("\nError: %v\n", err)
		}
	default:
	}

	fmt.Println("\n")

	if chatReferences != nil && len(chatReferences) > 0 {
		fmt.Println("\nSources:")
		for refNum, refData := range chatReferences {
			fmt.Printf("  [%s] %s\n", refNum, refData.MemoTitle)
			fmt.Printf("      UUID: %s\n", refData.MemoUUID)
		}
	}

	if chatID != "" {
		fmt.Printf("\nChat ID: %s\n", chatID)
		fmt.Println("(You can use this chat_id to continue the conversation with context)")
	}

	fmt.Println("\n" + strings.Repeat("=", 60) + "\n")

	// Example 3: Using references to retrieve full memo content
	fmt.Println("Example 3: Retrieving Full Memo from Reference\n")

	if response.References != nil && len(response.References) > 0 {
		// Get the first reference
		var firstRef skald.MemoReference
		for _, ref := range response.References {
			firstRef = ref
			break
		}

		fmt.Printf("Retrieving full content for: %s\n\n", firstRef.MemoTitle)

		memo, err := client.GetMemo(ctx, firstRef.MemoUUID)
		if err != nil {
			fmt.Printf("Error retrieving memo: %v\n", err)
		} else {
			fmt.Println("Memo Details:")
			fmt.Printf("  Title: %s\n", memo.Title)
			fmt.Printf("  Summary: %s\n", memo.Summary)
			fmt.Printf("  Content Length: %d characters\n", memo.ContentLength)
			tags := make([]string, len(memo.Tags))
			for i, t := range memo.Tags {
				tags[i] = t.Tag
			}
			fmt.Printf("  Tags: %s\n", strings.Join(tags, ", "))
			source := "N/A"
			if memo.Source != nil {
				source = *memo.Source
			}
			fmt.Printf("  Source: %s\n", source)
			fmt.Printf("  Created: %s\n", memo.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Println()
			fmt.Println("Content Preview (first 500 chars):")
			content := memo.Content
			if len(content) > 500 {
				content = content[:500] + "..."
			}
			fmt.Println(content)
		}
	} else {
		fmt.Println("No references available to demonstrate memo retrieval.")
	}

	fmt.Println("\n" + strings.Repeat("=", 60) + "\n")

	fmt.Println("\n[OK] References example completed successfully!")
}
