//go:build ignore

// Upload File and Wait for Processing Example
//
// This example demonstrates a complete workflow:
// 1. Upload a file
// 2. Poll the status until processing completes
// 3. Retrieve the processed memo
//
// Prerequisites:
// - Set SKALD_API_KEY environment variable
// - Have a file to upload (PDF, DOC, DOCX, or PPTX)
//
// Usage:
// go run examples/upload_and_wait.go

package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

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

	// Step 1: Upload the file
	filePath := "examples/localcurrency-snippet.pdf"

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("File not found: %s\n", filePath)
		fmt.Println("Please place a sample PDF in the examples directory")
		os.Exit(1)
	}

	fmt.Println("Step 1: Uploading file...")
	uploadResult, err := client.CreateMemoFromFile(ctx, filePath, &skald.MemoFileData{
		Metadata: map[string]interface{}{
			"uploadedAt": time.Now().Format(time.RFC3339),
			"type":       "example-document",
		},
		Tags:   []string{"example", "automated-workflow"},
		Source: ptr("example-script"),
	})
	if err != nil {
		fmt.Printf("Error uploading file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("[OK] File uploaded successfully!")
	fmt.Printf("Memo UUID: %s\n", uploadResult.MemoUUID)

	// Step 2: Poll the status until processing completes
	fmt.Println("\nStep 2: Waiting for processing to complete...")

	// Use WaitForMemoReady helper with timeout
	waitCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	err = client.WaitForMemoReady(waitCtx, uploadResult.MemoUUID.String(), 2*time.Second)
	if err != nil {
		fmt.Printf("\n[FAIL] Processing failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("[OK] Processing complete!")

	// Step 3: Retrieve the processed memo
	fmt.Println("\nStep 3: Retrieving processed memo...")
	memo, err := client.GetMemo(ctx, uploadResult.MemoUUID.String())
	if err != nil {
		fmt.Printf("Error retrieving memo: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n=== Processed Memo ===")
	fmt.Printf("Title: %s\n", memo.Title)
	fmt.Printf("Summary: %s\n", memo.Summary)
	fmt.Printf("Content Length: %d characters\n", memo.ContentLength)
	tags := make([]string, len(memo.Tags))
	for i, t := range memo.Tags {
		tags[i] = t.Tag
	}
	fmt.Printf("Tags: %s\n", strings.Join(tags, ", "))
	fmt.Printf("Chunks: %d\n", len(memo.Chunks))

	fmt.Println("\n[OK] Complete workflow finished successfully!")
}

func ptr[T any](v T) *T {
	return &v
}
