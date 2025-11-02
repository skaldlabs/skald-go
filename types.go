package skald

import (
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
	Query        string       `json:"query"`
	Limit        *int         `json:"limit,omitempty"`
	Filters      []Filter     `json:"filters,omitempty"`
}

// SearchResult represents a single search result
type SearchResult struct {
	UUID           string   `json:"uuid"`
	Title          string   `json:"title"`
	Summary        string   `json:"summary"`
	ContentSnippet string   `json:"content_snippet"`
	Distance       *float64 `json:"distance"` // Only populated for semantic search
}

// SearchResponse is the response from a search query
type SearchResponse struct {
	Results []SearchResult `json:"results"`
}

// ChatRequest contains parameters for a chat query
type ChatRequest struct {
	Query   string   `json:"query"`
	Stream  bool     `json:"stream"`
	Filters []Filter `json:"filters,omitempty"`
}

// ChatResponse is the response from a non-streaming chat query
type ChatResponse struct {
	OK                bool          `json:"ok"`
	Response          string        `json:"response"`
	IntermediateSteps []interface{} `json:"intermediate_steps"`
}

// ChatStreamEvent represents a streaming event from chat
type ChatStreamEvent struct {
	Type    string  `json:"type"`
	Content *string `json:"content,omitempty"`
}
