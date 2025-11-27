package skald

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// mockRoundTripper is a mock HTTP transport for testing
type mockRoundTripper struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

// newMockClient creates a client with a mock HTTP client
func newMockClient(roundTripFunc func(req *http.Request) (*http.Response, error)) *Client {
	client := NewClient("test-api-key")
	client.httpClient = &http.Client{
		Transport: &mockRoundTripper{roundTripFunc: roundTripFunc},
	}
	return client
}

// mockResponse creates a mock HTTP response
func mockResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		apiKey      string
		baseURL     []string
		expectedURL string
	}{
		{
			name:        "default base URL",
			apiKey:      "test-key",
			baseURL:     nil,
			expectedURL: "https://api.useskald.com",
		},
		{
			name:        "custom base URL",
			apiKey:      "test-key",
			baseURL:     []string{"https://custom.api.com"},
			expectedURL: "https://custom.api.com",
		},
		{
			name:        "base URL with trailing slash",
			apiKey:      "test-key",
			baseURL:     []string{"https://custom.api.com/"},
			expectedURL: "https://custom.api.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.apiKey, tt.baseURL...)
			if client.baseURL != tt.expectedURL {
				t.Errorf("expected baseURL %q, got %q", tt.expectedURL, client.baseURL)
			}
			if client.apiKey != tt.apiKey {
				t.Errorf("expected apiKey %q, got %q", tt.apiKey, client.apiKey)
			}
		})
	}
}

func TestCreateMemo(t *testing.T) {
	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		if req.Method != "POST" {
			t.Errorf("expected POST request, got %s", req.Method)
		}
		if req.URL.Path != "/api/v1/memo" {
			t.Errorf("expected path /api/v1/memo, got %s", req.URL.Path)
		}
		if req.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("expected Authorization header with Bearer token")
		}
		return mockResponse(200, `{"memo_uuid": "123e4567-e89b-12d3-a456-426614174000"}`), nil
	})

	resp, err := client.CreateMemo(context.Background(), MemoData{
		Title:   "Test Memo",
		Content: "This is test content",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.MemoUUID.String() != "123e4567-e89b-12d3-a456-426614174000" {
		t.Error("expected MemoUUID to be 123e4567-e89b-12d3-a456-426614174000")
	}
}

func TestCreateMemoInitializesMetadata(t *testing.T) {
	var capturedBody []byte
	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		var err error
		capturedBody, err = io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		return mockResponse(200, `{"memo_uuid": "123e4567-e89b-12d3-a456-426614174000"}`), nil
	})

	_, err := client.CreateMemo(context.Background(), MemoData{
		Title:   "Test Memo",
		Content: "This is test content",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that metadata is included (as empty object {} or null since omitempty is used)
	bodyStr := string(capturedBody)
	if !strings.Contains(bodyStr, `"metadata"`) {
		t.Error("expected metadata field in request body")
	}
}

func TestGetMemo(t *testing.T) {
	tests := []struct {
		name           string
		memoID         string
		idType         IDType
		expectedPath   string
		expectedParams string
	}{
		{
			name:           "get by UUID",
			memoID:         "test-uuid",
			idType:         IDTypeMemoUUID,
			expectedPath:   "/api/v1/memo/test-uuid",
			expectedParams: "",
		},
		{
			name:           "get by reference ID",
			memoID:         "test-ref-id",
			idType:         IDTypeReferenceID,
			expectedPath:   "/api/v1/memo/test-ref-id",
			expectedParams: "id_type=reference_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newMockClient(func(req *http.Request) (*http.Response, error) {
				if req.Method != "GET" {
					t.Errorf("expected GET request, got %s", req.Method)
				}
				if req.URL.Path != tt.expectedPath {
					t.Errorf("expected path %s, got %s", tt.expectedPath, req.URL.Path)
				}
				if req.URL.RawQuery != tt.expectedParams {
					t.Errorf("expected params %s, got %s", tt.expectedParams, req.URL.RawQuery)
				}
				return mockResponse(200, `{
					"uuid": "test-uuid",
					"created_at": "2024-01-01T00:00:00Z",
					"updated_at": "2024-01-01T00:00:00Z",
					"title": "Test Memo",
					"content": "Test content",
					"summary": "Test summary",
					"content_length": 100,
					"metadata": {},
					"client_reference_id": null,
					"source": null,
					"type": "memo",
					"expiration_date": null,
					"archived": false,
					"pending": false,
					"tags": [],
					"chunks": []
				}`), nil
			})

			memo, err := client.GetMemo(context.Background(), tt.memoID, tt.idType)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if memo.UUID != "test-uuid" {
				t.Errorf("expected UUID test-uuid, got %s", memo.UUID)
			}
		})
	}
}

func TestGetMemoInvalidIDType(t *testing.T) {
	client := NewClient("test-key")
	_, err := client.GetMemo(context.Background(), "test-id", IDType("invalid"))
	if err == nil {
		t.Error("expected error for invalid idType")
	}
}

func TestListMemos(t *testing.T) {
	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		if req.Method != "GET" {
			t.Errorf("expected GET request, got %s", req.Method)
		}
		if req.URL.Path != "/api/v1/memo" {
			t.Errorf("expected path /api/v1/memo, got %s", req.URL.Path)
		}

		// Check query parameters
		page := req.URL.Query().Get("page")
		pageSize := req.URL.Query().Get("page_size")
		if page != "2" || pageSize != "50" {
			t.Errorf("expected page=2&page_size=50, got page=%s&page_size=%s", page, pageSize)
		}

		return mockResponse(200, `{
			"count": 100,
			"next": null,
			"previous": "https://api.useskald.com/api/v1/memo?page=1",
			"results": [
				{
					"uuid": "test-uuid",
					"created_at": "2024-01-01T00:00:00Z",
					"updated_at": "2024-01-01T00:00:00Z",
					"title": "Test Memo",
					"summary": "Test summary",
					"content_length": 100,
					"metadata": {},
					"client_reference_id": null
				}
			]
		}`), nil
	})

	page := 2
	pageSize := 50
	resp, err := client.ListMemos(context.Background(), &ListMemosParams{
		Page:     &page,
		PageSize: &pageSize,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Count != 100 {
		t.Errorf("expected count 100, got %d", resp.Count)
	}
	if len(resp.Results) != 1 {
		t.Errorf("expected 1 result, got %d", len(resp.Results))
	}
}

func TestUpdateMemo(t *testing.T) {
	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		if req.Method != "PATCH" {
			t.Errorf("expected PATCH request, got %s", req.Method)
		}
		if req.URL.Path != "/api/v1/memo/test-uuid" {
			t.Errorf("expected path /api/v1/memo/test-uuid, got %s", req.URL.Path)
		}
		return mockResponse(200, `{"memo_uuid": "123e4567-e89b-12d3-a456-426614174000"}`), nil
	})

	title := "Updated Title"
	resp, err := client.UpdateMemo(context.Background(), "test-uuid", UpdateMemoData{
		Title: &title,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.MemoUUID.String() != "123e4567-e89b-12d3-a456-426614174000" {
		t.Error("expected OK to be true")
	}
}

func TestDeleteMemo(t *testing.T) {
	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		if req.Method != "DELETE" {
			t.Errorf("expected DELETE request, got %s", req.Method)
		}
		if req.URL.Path != "/api/v1/memo/test-uuid" {
			t.Errorf("expected path /api/v1/memo/test-uuid, got %s", req.URL.Path)
		}
		return mockResponse(204, ``), nil
	})

	err := client.DeleteMemo(context.Background(), "test-uuid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSearch(t *testing.T) {
	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		if req.Method != "POST" {
			t.Errorf("expected POST request, got %s", req.Method)
		}
		if req.URL.Path != "/api/v1/search" {
			t.Errorf("expected path /api/v1/search, got %s", req.URL.Path)
		}
		return mockResponse(200, `{
			"results": [
				{
					"memo_uuid": "test-uuid",
					"chunk_uuid": "test-chunk-uuid",
					"memo_title": "Test Memo",
					"memo_summary": "Test summary",
					"content_snippet": "Test snippet",
					"distance": 0.5
				}
			]
		}`), nil
	})

	limit := 10
	resp, err := client.Search(context.Background(), SearchRequest{
		Query: "test query",
		Limit: &limit,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Results) != 1 {
		t.Errorf("expected 1 result, got %d", len(resp.Results))
	}
	if resp.Results[0].MemoUUID != "test-uuid" {
		t.Errorf("expected MemoUUID test-uuid, got %s", resp.Results[0].MemoUUID)
	}
}

func TestSearchWithFilters(t *testing.T) {
	var capturedBody []byte
	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		var err error
		capturedBody, err = io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		return mockResponse(200, `{"results": []}`), nil
	})

	limit := 10
	_, err := client.Search(context.Background(), SearchRequest{
		Query: "test query",
		Limit: &limit,
		Filters: []Filter{
			{
				Field:      "source",
				Operator:   FilterOperatorEq,
				Value:      "notion",
				FilterType: FilterTypeNativeField,
			},
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify filters were included in request
	if !strings.Contains(string(capturedBody), `"filters"`) {
		t.Error("expected filters in request body")
	}
}

func TestChat(t *testing.T) {
	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		if req.Method != "POST" {
			t.Errorf("expected POST request, got %s", req.Method)
		}
		if req.URL.Path != "/api/v1/chat" {
			t.Errorf("expected path /api/v1/chat, got %s", req.URL.Path)
		}

		// Verify stream is false
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		if !strings.Contains(string(body), `"stream":false`) {
			t.Error("expected stream to be false")
		}

		return mockResponse(200, `{
			"ok": true,
			"response": "Test response with citation [[1]]",
			"intermediate_steps": []
		}`), nil
	})

	resp, err := client.Chat(context.Background(), ChatParams{
		Query: "What is the capital?",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("expected non-nil response")
		return
	}

	if !resp.OK {
		t.Error("expected OK to be true")
	}

	if !strings.Contains(resp.Response, "[[1]]") {
		t.Error("expected citation in response")
	}
}

func TestStreamedChat(t *testing.T) {
	sseData := `data: {"type":"token","content":"Hello"}
data: {"type":"token","content":" world"}
data: {"type":"done"}
`

	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		if req.Method != "POST" {
			t.Errorf("expected POST request, got %s", req.Method)
		}
		if req.URL.Path != "/api/v1/chat" {
			t.Errorf("expected path /api/v1/chat, got %s", req.URL.Path)
		}

		// Verify stream is true
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		if !strings.Contains(string(body), `"stream":true`) {
			t.Error("expected stream to be true")
		}

		return mockResponse(200, sseData), nil
	})

	eventChan, errChan := client.StreamedChat(context.Background(), ChatParams{
		Query: "test query",
	})

	var events []ChatStreamEvent
	for event := range eventChan {
		events = append(events, event)
	}

	// Check for errors
	select {
	case err := <-errChan:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	default:
	}

	if len(events) != 3 {
		t.Errorf("expected 3 events, got %d", len(events))
	}

	if events[0].Type != "token" || *events[0].Content != "Hello" {
		t.Error("unexpected first event")
	}
	if events[1].Type != "token" || *events[1].Content != " world" {
		t.Error("unexpected second event")
	}
	if events[2].Type != "done" {
		t.Error("unexpected third event")
	}
}

func TestStreamedChatWithInvalidJSON(t *testing.T) {
	sseData := `data: {"type":"token","content":"Valid"}
data: invalid json here
data: {"type":"done"}
`

	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockResponse(200, sseData), nil
	})

	eventChan, errChan := client.StreamedChat(context.Background(), ChatParams{
		Query: "test query",
	})

	var events []ChatStreamEvent
	for event := range eventChan {
		events = append(events, event)
	}

	// Check for errors
	select {
	case err := <-errChan:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	default:
	}

	// Invalid JSON should be skipped
	if len(events) != 2 {
		t.Errorf("expected 2 events (invalid JSON skipped), got %d", len(events))
	}
}

func TestStreamedChatWithPingLines(t *testing.T) {
	sseData := `: ping
data: {"type":"token","content":"Hello"}
: another ping
data: {"type":"done"}
`

	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockResponse(200, sseData), nil
	})

	eventChan, errChan := client.StreamedChat(context.Background(), ChatParams{
		Query: "test query",
	})

	var events []ChatStreamEvent
	for event := range eventChan {
		events = append(events, event)
	}

	// Check for errors
	select {
	case err := <-errChan:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	default:
	}

	// Ping lines should be skipped
	if len(events) != 2 {
		t.Errorf("expected 2 events (ping lines skipped), got %d", len(events))
	}
}

func TestAPIError(t *testing.T) {
	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockResponse(401, `{"error": "Invalid API key"}`), nil
	})

	_, err := client.CreateMemo(context.Background(), MemoData{
		Title:   "Test",
		Content: "Test",
	})

	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("expected error to contain status code 401, got: %v", err)
	}
}

func TestURLEncoding(t *testing.T) {
	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		// Check that special characters are properly encoded in the raw URL
		// req.URL.Path is automatically decoded, so we check EscapedPath() instead
		escapedPath := req.URL.EscapedPath()
		if !strings.Contains(escapedPath, "test%2Fid") {
			t.Errorf("expected URL-encoded path with %%2F, got %s", escapedPath)
		}
		return mockResponse(204, ``), nil
	})

	err := client.DeleteMemo(context.Background(), "test/id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMemoWithAllFields(t *testing.T) {
	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		bodyStr := string(body)

		// Verify all fields are present
		requiredFields := []string{"title", "content", "metadata", "reference_id", "tags", "source"}
		for _, field := range requiredFields {
			if !strings.Contains(bodyStr, field) {
				t.Errorf("expected field %s in request body", field)
			}
		}

		return mockResponse(200, `{"ok": true}`), nil
	})

	refID := "test-ref-123"
	source := "test-source"
	expirationDate := time.Now().Add(24 * time.Hour)

	_, err := client.CreateMemo(context.Background(), MemoData{
		Title:          "Test Memo",
		Content:        "Test content",
		Metadata:       map[string]interface{}{"key": "value"},
		ReferenceID:    &refID,
		Tags:           []string{"tag1", "tag2"},
		Source:         &source,
		ExpirationDate: &expirationDate,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateMemoFromFile(t *testing.T) {
	// Create a temporary test file
	tmpFile, err := os.CreateTemp("", "test-*.pdf")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	// Write some test content
	content := []byte("test PDF content")
	if _, err := tmpFile.Write(content); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}

	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		if req.Method != "POST" {
			t.Errorf("expected POST request, got %s", req.Method)
		}
		if req.URL.Path != "/api/v1/memo" {
			t.Errorf("expected path /api/v1/memo, got %s", req.URL.Path)
		}
		if req.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("expected Authorization header with Bearer token")
		}
		if !strings.Contains(req.Header.Get("Content-Type"), "multipart/form-data") {
			t.Errorf("expected multipart/form-data content type")
		}
		return mockResponse(200, `{"memo_uuid": "123e4567-e89b-12d3-a456-426614174000"}`), nil
	})

	title := "Test Document"
	source := "test-source"
	resp, err := client.CreateMemoFromFile(context.Background(), tmpFile.Name(), &MemoFileData{
		Title:  &title,
		Source: &source,
		Tags:   []string{"test", "document"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.MemoUUID.String() != "123e4567-e89b-12d3-a456-426614174000" {
		t.Error("expected MemoUUID to be 123e4567-e89b-12d3-a456-426614174000")
	}
}

func TestCreateMemoFromFileWithoutMemoData(t *testing.T) {
	// Create a temporary test file
	tmpFile, err := os.CreateTemp("", "test-*.pdf")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	content := []byte("test PDF content")
	if _, err := tmpFile.Write(content); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}

	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockResponse(200, `{"memo_uuid": "123e4567-e89b-12d3-a456-426614174000"}`), nil
	})

	resp, err := client.CreateMemoFromFile(context.Background(), tmpFile.Name(), nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.MemoUUID.String() != "123e4567-e89b-12d3-a456-426614174000" {
		t.Error("expected MemoUUID to be 123e4567-e89b-12d3-a456-426614174000")
	}
}

func TestCreateMemoFromFileNotFound(t *testing.T) {
	client := NewClient("test-key")
	_, err := client.CreateMemoFromFile(context.Background(), "/nonexistent/file.pdf", nil)
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestCreateMemoFromFileTooLarge(t *testing.T) {
	// Create a temporary test file that exceeds the size limit
	tmpFile, err := os.CreateTemp("", "test-large-*.pdf")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	// Seek to create a file larger than 100MB (we don't need to write the actual data)
	const largeSize = 101 * 1024 * 1024 // 101MB
	if err := tmpFile.Truncate(largeSize); err != nil {
		t.Fatalf("failed to truncate file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}

	client := NewClient("test-key")
	_, err = client.CreateMemoFromFile(context.Background(), tmpFile.Name(), nil)
	if err == nil {
		t.Error("expected error for file exceeding size limit")
	}
	if !strings.Contains(err.Error(), "100MB") {
		t.Errorf("expected error message about 100MB limit, got: %v", err)
	}
}

func TestCheckMemoStatus(t *testing.T) {
	tests := []struct {
		name           string
		memoID         string
		idType         IDType
		expectedPath   string
		expectedParams string
		responseStatus string
		expectedStatus MemoStatus
	}{
		{
			name:           "status by UUID - processing",
			memoID:         "test-uuid",
			idType:         IDTypeMemoUUID,
			expectedPath:   "/api/v1/memo/test-uuid/status",
			expectedParams: "",
			responseStatus: `{"status": "processing"}`,
			expectedStatus: MemoStatusProcessing,
		},
		{
			name:           "status by UUID - processed",
			memoID:         "test-uuid",
			idType:         IDTypeMemoUUID,
			expectedPath:   "/api/v1/memo/test-uuid/status",
			expectedParams: "",
			responseStatus: `{"status": "processed"}`,
			expectedStatus: MemoStatusProcessed,
		},
		{
			name:           "status by UUID - error",
			memoID:         "test-uuid",
			idType:         IDTypeMemoUUID,
			expectedPath:   "/api/v1/memo/test-uuid/status",
			expectedParams: "",
			responseStatus: `{"status": "error", "error_reason": "Processing failed"}`,
			expectedStatus: MemoStatusError,
		},
		{
			name:           "status by reference ID",
			memoID:         "test-ref-id",
			idType:         IDTypeReferenceID,
			expectedPath:   "/api/v1/memo/test-ref-id/status",
			expectedParams: "id_type=reference_id",
			responseStatus: `{"status": "processed"}`,
			expectedStatus: MemoStatusProcessed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newMockClient(func(req *http.Request) (*http.Response, error) {
				if req.Method != "GET" {
					t.Errorf("expected GET request, got %s", req.Method)
				}
				if req.URL.Path != tt.expectedPath {
					t.Errorf("expected path %s, got %s", tt.expectedPath, req.URL.Path)
				}
				if req.URL.RawQuery != tt.expectedParams {
					t.Errorf("expected params %s, got %s", tt.expectedParams, req.URL.RawQuery)
				}
				return mockResponse(200, tt.responseStatus), nil
			})

			status, err := client.CheckMemoStatus(context.Background(), tt.memoID, tt.idType)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if status.Status != tt.expectedStatus {
				t.Errorf("expected status %s, got %s", tt.expectedStatus, status.Status)
			}
		})
	}
}

func TestCheckMemoStatusWithErrorReason(t *testing.T) {
	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockResponse(200, `{"status": "error", "error_reason": "File format not supported"}`), nil
	})

	status, err := client.CheckMemoStatus(context.Background(), "test-uuid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Status != MemoStatusError {
		t.Errorf("expected status error, got %s", status.Status)
	}
	if status.ErrorReason == nil || *status.ErrorReason != "File format not supported" {
		t.Error("expected error reason to be 'File format not supported'")
	}
}

func TestCheckMemoStatusInvalidIDType(t *testing.T) {
	client := NewClient("test-key")
	_, err := client.CheckMemoStatus(context.Background(), "test-id", IDType("invalid"))
	if err == nil {
		t.Error("expected error for invalid idType")
	}
}
