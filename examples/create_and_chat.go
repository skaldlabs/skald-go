//go:build ignore

// Create Memo, Wait, and Chat Example
//
// This example demonstrates a complete workflow:
// 1. Create a memo with text content
// 2. Poll the status until processing completes
// 3. Ask questions about the memo using the chat API
//
// Prerequisites:
// - Set SKALD_API_KEY environment variable
//
// Usage:
// go run examples/create_and_chat.go

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

	// Step 1: Create a memo with text content
	fmt.Println("Step 1: Creating memo...")

	memoContent := strings.TrimSpace(`
# Team Meeting Notes - Q4 Planning

## Date: November 12, 2025

## Attendees
- Sarah Chen (Product Manager)
- Marcus Rodriguez (Engineering Lead)
- Julia Kim (Design Lead)
- Ahmed Hassan (Marketing Director)

## Key Discussion Points

### Product Roadmap
- Launch of new AI-powered search feature planned for January 2026
- Mobile app redesign to begin in December
- Integration with third-party tools (Slack, Microsoft Teams) scheduled for Q1 2026

### Engineering Updates
- Migration to microservices architecture 60% complete
- Performance improvements reduced page load times by 40%
- New deployment pipeline reduces release time from 2 hours to 20 minutes

### Design Initiatives
- User research showed 85% satisfaction with current interface
- Accessibility improvements needed for WCAG 2.1 AA compliance
- Dark mode feature requested by 67% of surveyed users

### Marketing Strategy
- Q3 user growth exceeded targets by 23%
- Focus on enterprise customers for Q4
- Partnership with TechCrunch for product launch coverage
- Budget allocation: 40% digital ads, 30% content marketing, 30% events

## Action Items
1. Sarah to finalize product specifications by November 20
2. Marcus to hire 2 additional backend engineers
3. Julia to complete accessibility audit by end of month
4. Ahmed to schedule partnership meetings with potential clients

## Next Meeting
December 10, 2025 at 2:00 PM
`)

	createResult, err := client.CreateMemo(ctx, skald.MemoData{
		Title:   "Q4 Planning Meeting Notes",
		Content: memoContent,
		Metadata: map[string]interface{}{
			"meetingDate":     "2025-11-12",
			"department":      "Product",
			"confidentiality": "internal",
		},
		Tags:   []string{"meeting-notes", "Q4-2025", "planning"},
		Source: ptr("example-script"),
	})
	if err != nil {
		fmt.Printf("Error creating memo: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("[OK] Memo created successfully!")
	fmt.Printf("Memo UUID: %s\n", createResult.MemoUUID)

	// Step 2: Poll the status until processing completes
	fmt.Println("\nStep 2: Waiting for processing to complete...")

	// Use WaitForMemoReady helper with timeout
	waitCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	err = client.WaitForMemoReady(waitCtx, createResult.MemoUUID.String(), 2*time.Second)
	if err != nil {
		fmt.Printf("\n[FAIL] Processing failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("[OK] Processing complete!")

	// Step 3: Retrieve the processed memo to see details
	fmt.Println("\nStep 3: Retrieving processed memo...")
	memo, err := client.GetMemo(ctx, createResult.MemoUUID.String())
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

	// Step 4: Ask questions about the memo
	fmt.Println("\n=== Step 4: Asking Questions About the Memo ===\n")

	// Question 1: Summarize key points
	fmt.Println("Q1: What were the main topics discussed in this meeting?")
	answer1, err := client.Chat(ctx, skald.ChatParams{
		Query: "What were the main topics discussed in this meeting?",
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("A1: %s\n\n", answer1.Response)
	}

	// Question 2: Specific information
	fmt.Println("Q2: Who needs to hire additional engineers and how many?")
	answer2, err := client.Chat(ctx, skald.ChatParams{
		Query: "Who needs to hire additional engineers and how many?",
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("A2: %s\n\n", answer2.Response)
	}

	// Question 3: Timeline information
	fmt.Println("Q3: What is the timeline for the mobile app redesign?")
	answer3, err := client.Chat(ctx, skald.ChatParams{
		Query: "What is the timeline for the mobile app redesign?",
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("A3: %s\n\n", answer3.Response)
	}

	// Question 4: Metrics and numbers
	fmt.Println("Q4: What were the performance improvements mentioned?")
	answer4, err := client.Chat(ctx, skald.ChatParams{
		Query: "What were the performance improvements mentioned?",
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("A4: %s\n\n", answer4.Response)
	}

	fmt.Println("\n[OK] Complete workflow finished successfully!")
}

func ptr[T any](v T) *T {
	return &v
}
