package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/skaldlabs/skald-go"
)

func main() {
	// Initialize the client
	apiKey := os.Getenv("SKALD_API_KEY")
	if apiKey == "" {
		log.Fatal("SKALD_API_KEY environment variable is required")
	}

	client := skald.NewClient(apiKey)
	ctx := context.Background()

	// Path to the file you want to upload
	// Supported formats: PDF, DOC, DOCX, PPTX (max 100MB)
	filePath := "examples/localcurrency-snippet.pdf"

	// Optional: Provide metadata for the memo
	title := "Overcoming economic instability with programmable money (Snippet)"
	source := "google-scholar"
	refID := "local-currency-snippet"

	memoData := &skald.MemoFileData{
		Title:       &title,
		Source:      &source,
		ReferenceID: &refID,
		Tags:        []string{"research", "paper", "article"},
		Metadata: map[string]interface{}{
			"journal":    "Journal of Risk Finance",
		},
	}

	fmt.Printf("Uploading file: %s\n", filePath)

	// Upload the file
	result, err := client.CreateMemoFromFile(ctx, filePath, memoData)
	if err != nil {
		log.Fatalf("Failed to upload file: %v", err)
	}

	memoUUID := result.MemoUUID.String()
	fmt.Printf("File uploaded successfully. Memo UUID: %s\n", memoUUID)
	fmt.Println("Processing document...")

	// Poll for status until processing is complete
	maxAttempts := 30               // Maximum number of polling attempts
	pollInterval := 2 * time.Second // Wait 2 seconds between checks

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Check the memo status
		status, err := client.CheckMemoStatus(ctx, memoUUID)
		if err != nil {
			log.Fatalf("Failed to check memo status: %v", err)
		}

		fmt.Printf("Status check %d/%d: %s\n", attempt, maxAttempts, status.Status)

		switch status.Status {
		case skald.MemoStatusProcessed:
			fmt.Println("\nDocument processed successfully!")

			// Retrieve the processed memo
			memo, err := client.GetMemo(ctx, memoUUID)
			if err != nil {
				log.Fatalf("Failed to retrieve memo: %v", err)
			}

			fmt.Printf("\nMemo Details:\n")
			fmt.Printf("Title: %s\n", memo.Title)
			fmt.Printf("Summary: %s\n", memo.Summary)
			fmt.Printf("Tags: %v\n", memo.Tags)
			fmt.Printf("Content Length: %d characters\n", memo.ContentLength)
			fmt.Printf("Chunks: %d\n", len(memo.Chunks))
			return

		case skald.MemoStatusError:
			errorMsg := "Unknown error"
			if status.ErrorReason != nil {
				errorMsg = *status.ErrorReason
			}
			log.Fatalf("Document processing failed: %s", errorMsg)

		case skald.MemoStatusProcessing:
			// Still processing, wait and try again
			if attempt < maxAttempts {
				time.Sleep(pollInterval)
			}
		}
	}

	fmt.Println("\nProcessing timeout. The document may still be processing.")
	fmt.Println("You can check the status later using CheckMemoStatus()")
}
