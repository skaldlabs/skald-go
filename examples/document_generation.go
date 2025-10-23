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

	// Example 1: Simple document generation
	fmt.Println("=== Simple Document Generation ===")
	genResp, err := client.GenerateDoc(ctx, "Write a brief guide on Go interfaces", nil, nil)
	if err != nil {
		log.Fatalf("Failed to generate document: %v", err)
	}

	fmt.Printf("Generated Document:\n%s\n\n", genResp.Response)

	// Example 2: Document generation with rules
	fmt.Println("=== Document Generation with Rules ===")
	rules := "Use bullet points and include code examples. Keep it under 500 words."
	rulesGenResp, err := client.GenerateDoc(ctx, "Create a tutorial on Go error handling", &rules, nil)
	if err != nil {
		log.Fatalf("Failed to generate document with rules: %v", err)
	}

	fmt.Printf("Generated Document:\n%s\n\n", rulesGenResp.Response)

	// Example 3: Document generation with filters
	fmt.Println("=== Document Generation with Filters ===")
	filters := []skald.Filter{
		{
			Field:      "category",
			Operator:   skald.FilterOperatorEq,
			Value:      "programming",
			FilterType: skald.FilterTypeCustomMetadata,
		},
		{
			Field:      "level",
			Operator:   skald.FilterOperatorEq,
			Value:      "beginner",
			FilterType: skald.FilterTypeCustomMetadata,
		},
	}

	filteredGenResp, err := client.GenerateDoc(ctx, "Write an introduction to Go for beginners", nil, filters)
	if err != nil {
		log.Fatalf("Failed to generate document with filters: %v", err)
	}

	fmt.Printf("Generated Document:\n%s\n\n", filteredGenResp.Response)

	// Example 4: Streaming document generation
	fmt.Println("=== Streaming Document Generation ===")
	eventChan, errChan := client.StreamedGenerateDoc(ctx, "Write a comparison between Go and other languages", nil, nil)

	fmt.Println("Generated Document (streaming):")
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
