// Check Memo Status Example
//
// This example demonstrates how to check the processing status of a memo,
// which is especially useful after uploading a file.
//
// Prerequisites:
// - Set SKALD_API_KEY environment variable
// - Have a memo UUID (from file upload or memo creation)
//
// Usage:
// go run examples/check_status.go <memo-uuid>

package main

import (
	"context"
	"fmt"
	"os"

	skald "github.com/skaldlabs/skald-go"
)

func main() {
	apiKey := os.Getenv("SKALD_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: SKALD_API_KEY environment variable not set")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run examples/check_status.go <memo-uuid>")
		os.Exit(1)
	}

	memoUUID := os.Args[1]
	client := skald.NewClient(apiKey)
	ctx := context.Background()

	fmt.Printf("Checking status for memo: %s\n", memoUUID)

	status, err := client.CheckMemoStatus(ctx, memoUUID)
	if err != nil {
		fmt.Printf("Error checking status: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Status: %s\n", status.Status)

	switch status.Status {
	case skald.MemoStatusProcessing:
		fmt.Println("The memo is still being processed...")
	case skald.MemoStatusProcessed:
		fmt.Println("[OK] The memo has been processed successfully!")
		fmt.Println("You can now search, chat, and retrieve this memo.")
	case skald.MemoStatusError:
		fmt.Println("[FAIL] An error occurred during processing")
		if status.ErrorReason != nil {
			fmt.Printf("Error reason: %s\n", *status.ErrorReason)
		}
	}

	fmt.Println("\nDone!")
}
