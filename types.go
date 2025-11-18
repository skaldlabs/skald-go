package skald

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// IDType specifies how to identify a memo
type IDType string

const (
	// IDTypeMemoUUID identifies a memo by its UUID
	IDTypeMemoUUID IDType = "memo_uuid"
	// IDTypeReferenceID identifies a memo by its client reference ID
	IDTypeReferenceID IDType = "reference_id"
)

// FilterOperator defines comparison operators for filtering
type FilterOperator string

const (
	// FilterOperatorEq matches exact equality
	FilterOperatorEq FilterOperator = "eq"
	// FilterOperatorNeq matches inequality
	FilterOperatorNeq FilterOperator = "neq"
	// FilterOperatorContains matches substring (case-insensitive)
	FilterOperatorContains FilterOperator = "contains"
	// FilterOperatorStartsWith matches prefix (case-insensitive)
	FilterOperatorStartsWith FilterOperator = "startswith"
	// FilterOperatorEndsWith matches suffix (case-insensitive)
	FilterOperatorEndsWith FilterOperator = "endswith"
	// FilterOperatorIn matches if value is in array
	FilterOperatorIn FilterOperator = "in"
	// FilterOperatorNotIn matches if value is not in array
	FilterOperatorNotIn FilterOperator = "not_in"
)

// FilterType specifies whether filter applies to native field or custom metadata
type FilterType string

const (
	// FilterTypeNativeField filters on built-in memo fields
	FilterTypeNativeField FilterType = "native_field"
	// FilterTypeCustomMetadata filters on custom metadata fields
	FilterTypeCustomMetadata FilterType = "custom_metadata"
)

// LLMProvider specifies the LLM provider for RAG
type LLMProvider string

const (
	// LLMProviderOpenAI uses OpenAI models
	LLMProviderOpenAI LLMProvider = "openai"
	// LLMProviderAnthropic uses Anthropic models
	LLMProviderAnthropic LLMProvider = "anthropic"
	// LLMProviderGroq uses Groq models
	LLMProviderGroq LLMProvider = "groq"
)

// QueryRewriteConfig configures query rewriting for RAG
type QueryRewriteConfig struct {
	Enabled bool `json:"enabled"`
}

// VectorSearchConfig configures vector search parameters for RAG
type VectorSearchConfig struct {
	TopK                int     `json:"topK"`
	SimilarityThreshold float64 `json:"similarityThreshold"`
}

// RerankingConfig configures reranking parameters for RAG
type RerankingConfig struct {
	Enabled bool `json:"enabled"`
	TopK    int  `json:"topK"`
}

// ReferencesConfig configures references in RAG responses
type ReferencesConfig struct {
	Enabled bool `json:"enabled"`
}

// RAGConfig contains all RAG configuration options
type RAGConfig struct {
	LLMProvider  LLMProvider         `json:"llmProvider,omitempty"`
	QueryRewrite *QueryRewriteConfig `json:"queryRewrite,omitempty"`
	VectorSearch *VectorSearchConfig `json:"vectorSearch,omitempty"`
	Reranking    *RerankingConfig    `json:"reranking,omitempty"`
	References   *ReferencesConfig   `json:"references,omitempty"`
}

// MemoReference represents a reference to a memo in chat responses
type MemoReference struct {
	MemoUUID  string `json:"memo_uuid"`
	MemoTitle string `json:"memo_title"`
}

// References maps citation numbers to memo references
type References map[string]MemoReference

// MemoData contains the data for creating a new memo
type MemoData struct {
	Title          string                 `json:"title"`
	Content        string                 `json:"content"`
	Metadata       map[string]interface{} `json:"metadata"`
	ReferenceID    *string                `json:"reference_id,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
	Source         *string                `json:"source,omitempty"`
	ExpirationDate *time.Time             `json:"expiration_date,omitempty"`
}

// CreateMemoResponse is the response from creating a memo
type CreateMemoResponse struct {
	MemoUUID uuid.UUID `json:"memo_uuid"`
}

// UpdateMemoData contains the fields that can be updated on a memo
type UpdateMemoData struct {
	Title             *string                `json:"title,omitempty"`
	Content           *string                `json:"content,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	ClientReferenceID *string                `json:"client_reference_id,omitempty"`
	Source            *string                `json:"source,omitempty"`
	ExpirationDate    *time.Time             `json:"expiration_date,omitempty"`
}

// UpdateMemoResponse is the response from updating a memo
type UpdateMemoResponse struct {
	MemoUUID uuid.UUID `json:"memo_uuid"`
}

// MemoTag represents a tag associated with a memo
type MemoTag struct {
	UUID string `json:"uuid"`
	Tag  string `json:"tag"`
}

// MemoChunk represents a content chunk used for semantic search
type MemoChunk struct {
	UUID         string `json:"uuid"`
	ChunkContent string `json:"chunk_content"`
	ChunkIndex   int    `json:"chunk_index"`
}

// Memo represents a complete memo with all its data
type Memo struct {
	UUID              string                 `json:"uuid"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	Title             string                 `json:"title"`
	Content           string                 `json:"content"`
	Summary           string                 `json:"summary"`
	ContentLength     int                    `json:"content_length"`
	Metadata          map[string]interface{} `json:"metadata"`
	ClientReferenceID *string                `json:"client_reference_id"`
	Source            *string                `json:"source"`
	Type              string                 `json:"type"`
	ExpirationDate    *time.Time             `json:"expiration_date"`
	Archived          bool                   `json:"archived"`
	Pending           bool                   `json:"pending"`
	Tags              []MemoTag              `json:"tags"`
	Chunks            []MemoChunk            `json:"chunks"`
}

// MemoListItem represents a memo in a list response
type MemoListItem struct {
	UUID              string                 `json:"uuid"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	Title             string                 `json:"title"`
	Summary           string                 `json:"summary"`
	ContentLength     int                    `json:"content_length"`
	Metadata          map[string]interface{} `json:"metadata"`
	ClientReferenceID *string                `json:"client_reference_id"`
}

// ListMemosParams contains parameters for listing memos
type ListMemosParams struct {
	Page     *int `json:"page,omitempty"`
	PageSize *int `json:"page_size,omitempty"`
}

// ListMemosResponse is the response from listing memos
type ListMemosResponse struct {
	Count    int            `json:"count"`
	Next     *string        `json:"next"`
	Previous *string        `json:"previous"`
	Results  []MemoListItem `json:"results"`
}

// Filter represents a filter condition for queries
type Filter struct {
	Field      string         `json:"field"`
	Operator   FilterOperator `json:"operator"`
	Value      interface{}    `json:"value"` // Can be string or []string
	FilterType FilterType     `json:"filter_type"`
}

// SearchRequest contains parameters for searching memos
type SearchRequest struct {
	Query   string   `json:"query"`
	Limit   *int     `json:"limit,omitempty"`
	Filters []Filter `json:"filters,omitempty"`
}

// SearchResult represents a single search result
type SearchResult struct {
	MemoUUID       string   `json:"memo_uuid"`
	ChunkUUID      string   `json:"chunk_uuid"`
	MemoTitle      string   `json:"memo_title"`
	MemoSummary    string   `json:"memo_summary"`
	ContentSnippet string   `json:"content_snippet"`
	Distance       *float64 `json:"distance"` // Only populated for semantic search
}

// SearchResponse is the response from a search query
type SearchResponse struct {
	Results []SearchResult `json:"results"`
}

// ChatParams contains parameters for chat queries.
// This is the public API struct that users pass to Chat() and StreamedChat() methods.
type ChatParams struct {
	Query        string     `json:"query"`
	Filters      []Filter   `json:"filters,omitempty"`
	SystemPrompt string     `json:"system_prompt,omitempty"`
	ChatID       string     `json:"chat_id,omitempty"`
	RAGConfig    *RAGConfig `json:"rag_config,omitempty"`
}

// chatRequest is the internal HTTP request payload structure.
// It includes the Stream field which is set automatically based on which method is called.
type chatRequest struct {
	Query        string     `json:"query"`
	Stream       bool       `json:"stream"`
	SystemPrompt string     `json:"system_prompt,omitempty"`
	Filters      []Filter   `json:"filters,omitempty"`
	ChatID       string     `json:"chat_id,omitempty"`
	RAGConfig    *RAGConfig `json:"rag_config,omitempty"`
}

// ChatResponse is the response from a non-streaming chat query
type ChatResponse struct {
	OK                bool          `json:"ok"`
	Response          string        `json:"response"`
	IntermediateSteps []interface{} `json:"intermediate_steps"`
	ChatID            string        `json:"chat_id,omitempty"`
	References        References    `json:"references,omitempty"`
}

// ChatStreamEvent represents a streaming event from chat
type ChatStreamEvent struct {
	Type       string     `json:"type"`
	Content    *string    `json:"content,omitempty"`
	ChatID     string     `json:"chat_id,omitempty"`
	References References `json:"references,omitempty"`
}

// MemoStatus represents the processing status of a memo
type MemoStatus string

const (
	// MemoStatusProcessing indicates the memo is being processed
	MemoStatusProcessing MemoStatus = "processing"
	// MemoStatusProcessed indicates the memo has been successfully processed
	MemoStatusProcessed MemoStatus = "processed"
	// MemoStatusError indicates the memo processing failed
	MemoStatusError MemoStatus = "error"
)

// MemoFileData contains the data for creating a memo from a file
type MemoFileData struct {
	Title          *string                `json:"title,omitempty"`
	Source         *string                `json:"source,omitempty"`
	ReferenceID    *string                `json:"reference_id,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	ExpirationDate *time.Time             `json:"expiration_date,omitempty"`
}

// MemoStatusResponse represents the response from checking memo status
type MemoStatusResponse struct {
	Status      MemoStatus `json:"status"`
	ErrorReason *string    `json:"error_reason,omitempty"`
}

// APIError represents an error returned by the Skald API
type APIError struct {
	StatusCode int
	Message    string
}

// Error implements the error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("skald API error (%d): %s", e.StatusCode, e.Message)
}

// IsNotFound returns true if the error is a 404 Not Found error
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == 404
}

// IsUnauthorized returns true if the error is a 401 Unauthorized error
func (e *APIError) IsUnauthorized() bool {
	return e.StatusCode == 401
}

// IsBadRequest returns true if the error is a 400 Bad Request error
func (e *APIError) IsBadRequest() bool {
	return e.StatusCode == 400
}
