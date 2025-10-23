# Skald Go SDK Examples

This directory contains example code demonstrating how to use the Skald Go SDK.

## Prerequisites

Before running these examples, you need to:

1. Install the Skald Go SDK:
   ```bash
   go get github.com/skald-org/skald-go
   ```

2. Set your Skald API key as an environment variable:
   ```bash
   export SKALD_API_KEY="your-api-key-here"
   ```

3. Have an active Skald instance with some memos in your knowledge base.

## Examples

### Basic Memo Operations
**File:** `basic_memo_operations.go`

Demonstrates:
- Creating a new memo with metadata, tags, and other fields
- Listing memos with pagination
- Getting a memo by UUID or reference ID
- Updating a memo
- Deleting a memo

Run:
```bash
go run examples/basic_memo_operations.go
```

### Search
**File:** `search.go`

Demonstrates:
- Semantic search using vector embeddings
- Title-based search (contains and startswith)
- Search with filters on native fields and custom metadata

Run:
```bash
go run examples/search.go
```

### Chat
**File:** `chat.go`

Demonstrates:
- Simple chat queries
- Chat with filters
- Streaming chat for real-time responses

Run:
```bash
go run examples/chat.go
```

### Document Generation
**File:** `document_generation.go`

Demonstrates:
- Simple document generation
- Document generation with custom rules
- Document generation with filters
- Streaming document generation

Run:
```bash
go run examples/document_generation.go
```

## Notes

- All examples require a valid `SKALD_API_KEY` environment variable
- Examples that create/modify data have delete operations commented out to prevent accidental data loss
- Streaming examples demonstrate real-time token-by-token output
- All examples include proper error handling and context usage
