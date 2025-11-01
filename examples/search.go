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

	// Example 1: Semantic search
	fmt.Println("=== Semantic Search ===")
	limit := 10
	searchResp, err := client.Search(ctx, skald.SearchRequest{
		Query:        "golang best practices",
		SearchMethod: skald.SearchMethodChunkVectorSearch,
		Limit:        &limit,
	})

	if err != nil {
		log.Fatalf("Failed to search: %v", err)
	}

	fmt.Printf("Found %d results\n", len(searchResp.Results))
	for i, result := range searchResp.Results {
		fmt.Printf("%d. %s (distance: %.4f)\n", i+1, result.Title, *result.Distance)
		fmt.Printf("   Summary: %s\n", result.Summary)
		fmt.Printf("   Snippet: %s\n\n", result.ContentSnippet)
	}

	// Example 2: Title search with contains
	fmt.Println("=== Title Contains Search ===")
	titleSearchResp, err := client.Search(ctx, skald.SearchRequest{
		Query:        "example",
		SearchMethod: skald.SearchMethodTitleContains,
		Limit:        &limit,
	})

	if err != nil {
		log.Fatalf("Failed to search: %v", err)
	}

	fmt.Printf("Found %d results\n", len(titleSearchResp.Results))
	for i, result := range titleSearchResp.Results {
		fmt.Printf("%d. %s\n", i+1, result.Title)
	}
	fmt.Println()

	// Example 3: Search with filters
	fmt.Println("=== Search with Filters ===")
	filteredSearchResp, err := client.Search(ctx, skald.SearchRequest{
		Query:        "programming",
		SearchMethod: skald.SearchMethodChunkVectorSearch,
		Limit:        &limit,
		Filters: []skald.Filter{
			{
				Field:      "source",
				Operator:   skald.FilterOperatorEq,
				Value:      "example-app",
				FilterType: skald.FilterTypeNativeField,
			},
			{
				Field:      "tags",
				Operator:   skald.FilterOperatorIn,
				Value:      []string{"go", "programming"},
				FilterType: skald.FilterTypeNativeField,
			},
			{
				Field:      "category",
				Operator:   skald.FilterOperatorEq,
				Value:      "programming",
				FilterType: skald.FilterTypeCustomMetadata,
			},
		},
	})

	if err != nil {
		log.Fatalf("Failed to search with filters: %v", err)
	}

	fmt.Printf("Found %d results with filters\n", len(filteredSearchResp.Results))
	for i, result := range filteredSearchResp.Results {
		fmt.Printf("%d. %s (distance: %.4f)\n", i+1, result.Title, *result.Distance)
	}
}
