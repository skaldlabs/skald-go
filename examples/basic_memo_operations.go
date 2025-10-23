//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

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

	// Create a new memo
	fmt.Println("Creating a new memo...")
	refID := "example-ref-123"
	source := "example-app"
	expirationDate := time.Now().Add(30 * 24 * time.Hour) // 30 days from now

	createResp, err := client.CreateMemo(ctx, skald.MemoData{
		Title:   "Example Memo",
		Content: "This is an example memo created with the Skald Go SDK. It contains information about Go programming language features.",
		Metadata: map[string]interface{}{
			"category": "programming",
			"language": "go",
			"level":    "beginner",
		},
		ReferenceID:    &refID,
		Tags:           []string{"go", "programming", "example"},
		Source:         &source,
		ExpirationDate: &expirationDate,
	})

	if err != nil {
		log.Fatalf("Failed to create memo: %v", err)
	}

	fmt.Printf("Memo created successfully: %+v\n\n", createResp)

	// List memos
	fmt.Println("Listing memos...")
	page := 1
	pageSize := 20
	listResp, err := client.ListMemos(ctx, &skald.ListMemosParams{
		Page:     &page,
		PageSize: &pageSize,
	})

	if err != nil {
		log.Fatalf("Failed to list memos: %v", err)
	}

	fmt.Printf("Found %d total memos\n", listResp.Count)
	for i, memo := range listResp.Results {
		fmt.Printf("%d. %s - %s\n", i+1, memo.Title, memo.Summary)
	}
	fmt.Println()

	// Get memo by reference ID
	if len(listResp.Results) > 0 {
		fmt.Println("Getting memo by reference ID...")
		memo, err := client.GetMemo(ctx, refID, skald.IDTypeReferenceID)
		if err != nil {
			log.Printf("Failed to get memo: %v", err)
		} else {
			fmt.Printf("Retrieved memo:\n")
			fmt.Printf("  Title: %s\n", memo.Title)
			fmt.Printf("  Content: %s\n", memo.Content)
			fmt.Printf("  Summary: %s\n", memo.Summary)
			fmt.Printf("  Tags: %v\n", memo.Tags)
			fmt.Printf("  Chunks: %d\n", len(memo.Chunks))
			fmt.Println()

			// Update the memo
			fmt.Println("Updating memo...")
			newTitle := "Updated Example Memo"
			updateResp, err := client.UpdateMemo(ctx, refID, skald.UpdateMemoData{
				Title: &newTitle,
			}, skald.IDTypeReferenceID)

			if err != nil {
				log.Printf("Failed to update memo: %v", err)
			} else {
				fmt.Printf("Memo updated successfully: %+v\n\n", updateResp)
			}

			// Get the updated memo
			fmt.Println("Getting updated memo...")
			updatedMemo, err := client.GetMemo(ctx, memo.UUID)
			if err != nil {
				log.Printf("Failed to get updated memo: %v", err)
			} else {
				fmt.Printf("Updated title: %s\n\n", updatedMemo.Title)
			}

			// Delete the memo (commented out to avoid accidental deletion)
			// fmt.Println("Deleting memo...")
			// err = client.DeleteMemo(ctx, memo.UUID)
			// if err != nil {
			// 	log.Printf("Failed to delete memo: %v", err)
			// } else {
			// 	fmt.Println("Memo deleted successfully")
			// }
		}
	}
}
