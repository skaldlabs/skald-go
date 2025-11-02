# Skald Go SDK

Go client library for the Skald API.

## Installation

```bash
go get github.com/skald-org/skald-go
```

## Requirements

- Go 1.18 or higher

## Usage

### Initialize the client

```go
import "github.com/skald-org/skald-go"

client := skald.NewClient("your-api-key-here")
```

You can optionally specify a custom base URL (e.g., for self-hosted instances):

```go
client := skald.NewClient("your-api-key-here", "https://custom-api.example.com")
```

### Memo Management

#### Create a Memo

Create a new memo that will be automatically processed (summarized, tagged, chunked, and indexed for search):

```go
refID := "external-id-123"
source := "notion"
expirationDate := time.Now().Add(30 * 24 * time.Hour)

result, err := client.CreateMemo(ctx, skald.MemoData{
    Title:   "Meeting Notes",
    Content: "Full content of the memo...",
    Metadata: map[string]interface{}{
        "type":   "notes",
        "author": "John Doe",
    },
    ReferenceID:    &refID,
    Tags:           []string{"meeting", "q1"},
    Source:         &source,
    ExpirationDate: &expirationDate,
})

if err != nil {
    log.Fatal(err)
}

fmt.Println(result.OK) // true
```

**Required Fields:**
- `Title` (string, max 255 chars) - The title of the memo
- `Content` (string) - The full content of the memo

**Optional Fields:**
- `Metadata` (map[string]interface{}) - Custom JSON metadata
- `ReferenceID` (*string, max 255 chars) - An ID from your side that you can use to match Skald memo UUIDs with e.g. documents on your end
- `Tags` ([]string) - Tags for categorization
- `Source` (*string, max 255 chars) - An indication from your side of the source of this content, useful when building integrations
- `ExpirationDate` (*time.Time) - Timestamp for automatic memo expiration

#### Get a Memo

Retrieve a memo by its UUID or your reference ID:

```go
// Get by UUID
memo, err := client.GetMemo(ctx, "550e8400-e29b-41d4-a716-446655440000")

// Get by reference ID
memo, err := client.GetMemo(ctx, "external-id-123", skald.IDTypeReferenceID)

if err != nil {
    log.Fatal(err)
}

fmt.Println(memo.Title)
fmt.Println(memo.Content)
fmt.Println(memo.Summary)
fmt.Println(memo.Tags)
fmt.Println(memo.Chunks)
```

The `GetMemo()` method returns complete memo details including content, AI-generated summary, tags, and content chunks.

#### List Memos

List all memos with pagination:

```go
// Get first page with default page size (20)
memos, err := client.ListMemos(ctx, nil)

// Get specific page with custom page size
page := 2
pageSize := 50
memos, err := client.ListMemos(ctx, &skald.ListMemosParams{
    Page:     &page,
    PageSize: &pageSize,
})

if err != nil {
    log.Fatal(err)
}

fmt.Printf("Total memos: %d\n", memos.Count)
fmt.Printf("Results: %d\n", len(memos.Results))
```

**Parameters:**
- `Page` (*int, optional) - Page number (default: 1)
- `PageSize` (*int, optional) - Results per page (default: 20, max: 100)

#### Update a Memo

Update an existing memo by UUID or reference ID:

```go
// Update by UUID
title := "Updated Title"
_, err := client.UpdateMemo(ctx, "550e8400-e29b-41d4-a716-446655440000", skald.UpdateMemoData{
    Title: &title,
    Metadata: map[string]interface{}{
        "status": "reviewed",
    },
})

// Update by reference ID and trigger reprocessing
content := "New content that will be reprocessed"
_, err := client.UpdateMemo(ctx, "external-id-123", skald.UpdateMemoData{
    Content: &content,
}, skald.IDTypeReferenceID)
```

**Note:** When you update the `Content` field, the memo will be automatically reprocessed (summary, tags, and chunks regenerated).

**Updatable Fields:**
- `Title` (*string)
- `Content` (*string)
- `Metadata` (map[string]interface{})
- `ClientReferenceID` (*string)
- `Source` (*string)
- `ExpirationDate` (*time.Time)

#### Delete a Memo

Permanently delete a memo and all associated data:

```go
// Delete by UUID
err := client.DeleteMemo(ctx, "550e8400-e29b-41d4-a716-446655440000")

// Delete by reference ID
err := client.DeleteMemo(ctx, "external-id-123", skald.IDTypeReferenceID)

if err != nil {
    log.Fatal(err)
}
```

**Warning:** This operation permanently deletes the memo and all related data (content, summary, tags, chunks) and cannot be undone.

### Search Memos

Search through your memos using semantic search:

```go
// Basic semantic search
limit := 10
results, err := client.Search(ctx, skald.SearchRequest{
    Query:        "quarterly goals",
    Limit:        &limit,
})

// Search with filters
filtered, err := client.Search(ctx, skald.SearchRequest{
    Query:        "python tutorial",
    Filters: []skald.Filter{
        {
            Field:      "source",
            Operator:   skald.FilterOperatorEq,
            Value:      "notion",
            FilterType: skald.FilterTypeNativeField,
        },
        {
            Field:      "level",
            Operator:   skald.FilterOperatorEq,
            Value:      "beginner",
            FilterType: skald.FilterTypeCustomMetadata,
        },
    },
})

if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d results\n", len(filtered.Results))
for _, memo := range filtered.Results {
    fmt.Printf("- %s (distance: %.4f)\n", memo.Title, *memo.Distance)
}
```
#### Search Parameters

- `Query` (string, required) - The search query
- `Limit` (*int, optional) - Maximum results to return (1-50, default 10)
- `Filters` ([]Filter, optional) - Array of filter objects to narrow results (see Filters section below)

#### Search Response

```go
type SearchResponse struct {
    Results []SearchResult
}

type SearchResult struct {
    UUID           string
    Title          string
    Summary        string
    ContentSnippet string
    Distance       *float64  // Only populated for semantic search
}
```

- `UUID` - Unique identifier for the memo
- `Title` - Memo title
- `Summary` - Auto-generated summary for the memo
- `ContentSnippet` - A snippet containing the beginning of the memo
- `Distance` - A decimal from 0 to 2 determining how close the result was deemed to be to the query.

### Chat with Your Knowledge Base

Ask questions about your memos using an AI agent. The agent retrieves relevant context and generates answers with inline citations.

#### Non-Streaming Chat

```go
result, err := client.Chat(ctx, "What were the main points discussed in the Q1 meeting?", nil)
if err != nil {
    log.Fatal(err)
}

fmt.Println(result.Response)
// "The main points discussed in the Q1 meeting were:
// 1. Revenue targets [[1]]
// 2. Hiring plans [[2]]
// 3. Product roadmap [[1]][[3]]"

fmt.Println(result.OK) // true
```

#### Streaming Chat

For real-time responses, use streaming chat:

```go
eventChan, errChan := client.StreamedChat(ctx, "What are our quarterly goals?", nil)

for event := range eventChan {
    if event.Type == "token" && event.Content != nil {
        // Write each token as it arrives
        fmt.Print(*event.Content)
    } else if event.Type == "done" {
        fmt.Println("\nDone!")
        break
    }
}

// Check for errors
select {
case err := <-errChan:
    if err != nil {
        log.Fatal(err)
    }
default:
}
```

#### Chat Parameters

- `query` (string, required) - The question to ask
- `filters` ([]Filter, optional) - Array of filter objects to focus chat context on specific sources (see Filters section below)

#### Chat Response

Non-streaming responses include:
- `OK` (bool) - Success status
- `Response` (string) - The AI's answer with inline citations in format `[[N]]`
- `IntermediateSteps` ([]interface{}) - Steps taken by the agent (for debugging)

Streaming responses yield events:
- `{ Type: "token", Content: *string }` - Each text token as it's generated
- `{ Type: "done" }` - Indicates the stream has finished

### Generate Documents

Generate documents based on prompts and retrieved context from your knowledge base. Similar to chat but optimized for document generation with optional style/format rules.

#### Non-Streaming Document Generation

```go
rules := "Use formal business language. Include sections for: Overview, Requirements, Technical Specifications, Timeline"
result, err := client.GenerateDoc(ctx, "Create a product requirements document for a new mobile app", &rules, nil)

if err != nil {
    log.Fatal(err)
}

fmt.Println(result.Response)
// "# Product Requirements Document
//
// ## Overview
// This document outlines the requirements for...
//
// ## Requirements
// 1. User authentication [[1]]
// 2. Push notifications [[2]]..."

fmt.Println(result.OK) // true
```

#### Streaming Document Generation

For real-time document generation, use streaming:

```go
rules := "Include sections for: Architecture, Security, API Endpoints, Data Models"
eventChan, errChan := client.StreamedGenerateDoc(ctx, "Write a technical specification for user authentication", &rules, nil)

for event := range eventChan {
    if event.Type == "token" && event.Content != nil {
        // Write each token as it arrives
        fmt.Print(*event.Content)
    } else if event.Type == "done" {
        fmt.Println("\nDone!")
        break
    }
}

// Check for errors
select {
case err := <-errChan:
    if err != nil {
        log.Fatal(err)
    }
default:
}
```

#### Generate Document Parameters

- `prompt` (string, required) - The prompt describing what document to generate
- `rules` (*string, optional) - Optional style/format rules (e.g., "Use formal language. Include sections: X, Y, Z")
- `filters` ([]Filter, optional) - Array of filter objects to control which memos are used for generation (see Filters section below)

#### Generate Document Response

Non-streaming responses include:
- `OK` (bool) - Success status
- `Response` (string) - The generated document with inline citations in format `[[N]]`
- `IntermediateSteps` ([]interface{}) - Steps taken by the agent (for debugging)

Streaming responses yield events:
- `{ Type: "token", Content: *string }` - Each text token as it's generated
- `{ Type: "done" }` - Indicates the stream has finished

### Filters

Filters allow you to narrow down results based on memo metadata. You can filter by native fields or custom metadata fields. Filters are supported in `Search()`, `Chat()`, `StreamedChat()`, `GenerateDoc()`, and `StreamedGenerateDoc()`.

#### Filter Structure

```go
type Filter struct {
    Field      string         // Field name to filter on
    Operator   FilterOperator // Comparison operator
    Value      interface{}    // Value(s) to compare against (string or []string)
    FilterType FilterType     // 'native_field' or 'custom_metadata'
}
```

#### Native Fields

Native fields are built-in memo properties:
- `title` - Memo title
- `source` - Source system (e.g., "notion", "confluence")
- `client_reference_id` - Your external reference ID
- `tags` - Memo tags (array)

#### Custom Metadata Fields

You can filter on any field from the `Metadata` map you provided when creating the memo.

#### Filter Operators

- **`FilterOperatorEq`** - Equals (exact match)
- **`FilterOperatorNeq`** - Not equals
- **`FilterOperatorContains`** - Contains substring (case-insensitive)
- **`FilterOperatorStartsWith`** - Starts with prefix (case-insensitive)
- **`FilterOperatorEndsWith`** - Ends with suffix (case-insensitive)
- **`FilterOperatorIn`** - Value is in array (requires array value)
- **`FilterOperatorNotIn`** - Value is not in array (requires array value)

#### Filter Examples

```go
// Filter by source
skald.Filter{
    Field:      "source",
    Operator:   skald.FilterOperatorEq,
    Value:      "notion",
    FilterType: skald.FilterTypeNativeField,
}

// Filter by multiple tags
skald.Filter{
    Field:      "tags",
    Operator:   skald.FilterOperatorIn,
    Value:      []string{"security", "compliance"},
    FilterType: skald.FilterTypeNativeField,
}

// Filter by title containing text
skald.Filter{
    Field:      "title",
    Operator:   skald.FilterOperatorContains,
    Value:      "meeting",
    FilterType: skald.FilterTypeNativeField,
}

// Filter by custom metadata field
skald.Filter{
    Field:      "department",
    Operator:   skald.FilterOperatorEq,
    Value:      "engineering",
    FilterType: skald.FilterTypeCustomMetadata,
}

// Exclude specific sources
skald.Filter{
    Field:      "source",
    Operator:   skald.FilterOperatorNotIn,
    Value:      []string{"draft", "archive"},
    FilterType: skald.FilterTypeNativeField,
}
```

#### Combining Multiple Filters

When you provide multiple filters, they are combined with AND logic (all filters must match):

```go
results, err := client.Search(ctx, skald.SearchRequest{
    Query:        "security best practices",
    Filters: []skald.Filter{
        {
            Field:      "source",
            Operator:   skald.FilterOperatorEq,
            Value:      "security-docs",
            FilterType: skald.FilterTypeNativeField,
        },
        {
            Field:      "tags",
            Operator:   skald.FilterOperatorIn,
            Value:      []string{"approved", "current"},
            FilterType: skald.FilterTypeNativeField,
        },
        {
            Field:      "status",
            Operator:   skald.FilterOperatorNeq,
            Value:      "draft",
            FilterType: skald.FilterTypeCustomMetadata,
        },
    },
})
```

#### Filters with Chat

Focus chat context on specific sources:

```go
result, err := client.Chat(ctx, "What are our security practices?", []skald.Filter{
    {
        Field:      "tags",
        Operator:   skald.FilterOperatorIn,
        Value:      []string{"security", "compliance"},
        FilterType: skald.FilterTypeNativeField,
    },
})
```

#### Filters with Document Generation

Control which memos are used for document generation:

```go
rules := "Use technical language with code examples"
doc, err := client.GenerateDoc(ctx, "Create an API integration guide", &rules, []skald.Filter{
    {
        Field:      "source",
        Operator:   skald.FilterOperatorIn,
        Value:      []string{"api-docs", "technical-specs"},
        FilterType: skald.FilterTypeNativeField,
    },
    {
        Field:      "document_type",
        Operator:   skald.FilterOperatorEq,
        Value:      "specification",
        FilterType: skald.FilterTypeCustomMetadata,
    },
})
```

### Error Handling

```go
result, err := client.CreateMemo(ctx, skald.MemoData{
    Title:   "My Memo",
    Content: "Content here",
})

if err != nil {
    log.Printf("Error: %v", err)
    return
}

fmt.Println("Success:", result)
```

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/skald-org/skald-go"
)

func main() {
    // Initialize client
    client := skald.NewClient("your-api-key-here")
    ctx := context.Background()

    // Create a memo
    refID := "example-123"
    source := "example-app"
    result, err := client.CreateMemo(ctx, skald.MemoData{
        Title:   "Go Programming Best Practices",
        Content: "Go is a statically typed, compiled programming language...",
        Metadata: map[string]interface{}{
            "category": "programming",
            "level":    "intermediate",
        },
        ReferenceID: &refID,
        Tags:        []string{"go", "programming"},
        Source:      &source,
    })

    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Created memo: %+v\n", result)

    // Search for memos
    limit := 5
    searchResults, err := client.Search(ctx, skald.SearchRequest{
        Query:        "golang best practices",
        Limit:        &limit,
    })

    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d results:\n", len(searchResults.Results))
    for _, result := range searchResults.Results {
        fmt.Printf("- %s\n", result.Title)
    }

    // Chat with knowledge base
    chatResp, err := client.Chat(ctx, "What are Go best practices?", nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Answer: %s\n", chatResp.Response)

    // Generate a document
    rules := "Use bullet points and be concise"
    doc, err := client.GenerateDoc(ctx, "Create a Go style guide", &rules, nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Generated document:\n%s\n", doc.Response)
}
```

## Type Definitions

The SDK exports the following types for use in your Go code:

```go
// ID and filter types
type IDType string
type FilterOperator string
type FilterType string

// Memo types
type MemoData struct { ... }
type CreateMemoResponse struct { ... }
type UpdateMemoData struct { ... }
type UpdateMemoResponse struct { ... }
type Memo struct { ... }
type MemoListItem struct { ... }
type ListMemosParams struct { ... }
type ListMemosResponse struct { ... }
type MemoTag struct { ... }
type MemoChunk struct { ... }

// Filter types
type Filter struct { ... }

// Search types
type SearchRequest struct { ... }
type SearchResponse struct { ... }
type SearchResult struct { ... }

// Chat types
type ChatRequest struct { ... }
type ChatResponse struct { ... }
type ChatStreamEvent struct { ... }

// Document generation types
type GenerateDocRequest struct { ... }
type GenerateDocResponse struct { ... }
type GenerateDocStreamEvent struct { ... }
```

See the [types.go](types.go) file for complete type definitions.

## Examples

See the [examples](examples/) directory for complete working examples demonstrating all SDK features.

## Testing

Run the test suite:

```bash
go test -v -cover
```

## License

MIT
