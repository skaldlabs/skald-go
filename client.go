package skald

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Client is the main Skald SDK client
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Skald client
func NewClient(apiKey string, baseURL ...string) *Client {
	url := "https://api.useskald.com"
	if len(baseURL) > 0 && baseURL[0] != "" {
		url = strings.TrimRight(baseURL[0], "/")
	}

	return &Client{
		apiKey:     apiKey,
		baseURL:    url,
		httpClient: &http.Client{},
	}
}

// CreateMemo creates a new memo
func (c *Client) CreateMemo(ctx context.Context, memoData MemoData) (*CreateMemoResponse, error) {
	// Initialize metadata to empty map if not provided
	if memoData.Metadata == nil {
		memoData.Metadata = make(map[string]interface{})
	}

	body, err := json.Marshal(memoData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal memo data: %w", err)
	}

	resp, err := c.doRequest(ctx, "POST", "/api/v1/memo", nil, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}

	var result CreateMemoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// CreateMemoFromFile creates a new memo by uploading a file
// Supported file formats: PDF, DOC, DOCX, PPTX
// Maximum file size: 100MB
func (c *Client) CreateMemoFromFile(ctx context.Context, filePath string, memoData *MemoFileData) (*CreateMemoResponse, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Get file info for validation
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Check file size (100MB limit)
	const maxFileSize = 100 * 1024 * 1024 // 100MB
	if fileInfo.Size() > maxFileSize {
		return nil, fmt.Errorf("file size exceeds 100MB limit")
	}

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file field
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}

	// Add memo data if provided
	if memoData != nil {
		// Convert memoData to JSON
		memoDataJSON, err := json.Marshal(memoData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal memo data: %w", err)
		}

		if err := writer.WriteField("memo_data", string(memoDataJSON)); err != nil {
			return nil, fmt.Errorf("failed to write memo data field: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create request
	urlStr := c.baseURL + "/api/v1/memo/upload"
	req, err := http.NewRequestWithContext(ctx, "POST", urlStr, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}

	var result CreateMemoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetMemo retrieves a memo by ID
func (c *Client) GetMemo(ctx context.Context, memoID string, idType ...IDType) (*Memo, error) {
	idTypeValue := IDTypeMemoUUID
	if len(idType) > 0 {
		idTypeValue = idType[0]
		if idTypeValue != IDTypeMemoUUID && idTypeValue != IDTypeReferenceID {
			return nil, fmt.Errorf("invalid idType: must be 'memo_uuid' or 'reference_id'")
		}
	}

	params := url.Values{}
	if idTypeValue != IDTypeMemoUUID {
		params.Set("id_type", string(idTypeValue))
	}

	path := fmt.Sprintf("/api/v1/memo/%s", url.PathEscape(memoID))
	resp, err := c.doRequest(ctx, "GET", path, params, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}

	var memo Memo
	if err := json.NewDecoder(resp.Body).Decode(&memo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &memo, nil
}

// ListMemos retrieves a paginated list of memos
func (c *Client) ListMemos(ctx context.Context, params *ListMemosParams) (*ListMemosResponse, error) {
	queryParams := url.Values{}
	if params != nil {
		if params.Page != nil {
			queryParams.Set("page", fmt.Sprintf("%d", *params.Page))
		}
		if params.PageSize != nil {
			queryParams.Set("page_size", fmt.Sprintf("%d", *params.PageSize))
		}
	}

	resp, err := c.doRequest(ctx, "GET", "/api/v1/memo", queryParams, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}

	var result ListMemosResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// UpdateMemo updates an existing memo
func (c *Client) UpdateMemo(ctx context.Context, memoID string, updateData UpdateMemoData, idType ...IDType) (*UpdateMemoResponse, error) {
	idTypeValue := IDTypeMemoUUID
	if len(idType) > 0 {
		idTypeValue = idType[0]
		if idTypeValue != IDTypeMemoUUID && idTypeValue != IDTypeReferenceID {
			return nil, fmt.Errorf("invalid idType: must be 'memo_uuid' or 'reference_id'")
		}
	}

	params := url.Values{}
	if idTypeValue != IDTypeMemoUUID {
		params.Set("id_type", string(idTypeValue))
	}

	body, err := json.Marshal(updateData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal update data: %w", err)
	}

	path := fmt.Sprintf("/api/v1/memo/%s", url.PathEscape(memoID))
	resp, err := c.doRequest(ctx, "PATCH", path, params, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}

	var result UpdateMemoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// DeleteMemo deletes a memo
func (c *Client) DeleteMemo(ctx context.Context, memoID string, idType ...IDType) error {
	idTypeValue := IDTypeMemoUUID
	if len(idType) > 0 {
		idTypeValue = idType[0]
		if idTypeValue != IDTypeMemoUUID && idTypeValue != IDTypeReferenceID {
			return fmt.Errorf("invalid idType: must be 'memo_uuid' or 'reference_id'")
		}
	}

	params := url.Values{}
	if idTypeValue != IDTypeMemoUUID {
		params.Set("id_type", string(idTypeValue))
	}

	path := fmt.Sprintf("/api/v1/memo/%s", url.PathEscape(memoID))
	resp, err := c.doRequest(ctx, "DELETE", path, params, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := c.checkResponse(resp); err != nil {
		return err
	}

	return nil
}

// CheckMemoStatus checks the processing status of a memo
// The memo can be identified by UUID (default) or reference ID
func (c *Client) CheckMemoStatus(ctx context.Context, memoID string, idType ...IDType) (*MemoStatusResponse, error) {
	idTypeValue := IDTypeMemoUUID
	if len(idType) > 0 {
		idTypeValue = idType[0]
		if idTypeValue != IDTypeMemoUUID && idTypeValue != IDTypeReferenceID {
			return nil, fmt.Errorf("invalid idType: must be 'memo_uuid' or 'reference_id'")
		}
	}

	params := url.Values{}
	if idTypeValue != IDTypeMemoUUID {
		params.Set("id_type", string(idTypeValue))
	}

	path := fmt.Sprintf("/api/v1/memo/%s/status", url.PathEscape(memoID))
	resp, err := c.doRequest(ctx, "GET", path, params, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}

	var status MemoStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &status, nil
}

// Search searches for memos
func (c *Client) Search(ctx context.Context, searchReq SearchRequest) (*SearchResponse, error) {
	body, err := json.Marshal(searchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search request: %w", err)
	}

	resp, err := c.doRequest(ctx, "POST", "/api/v1/search", nil, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// Chat performs a non-streaming chat query and returns the response
func (c *Client) Chat(ctx context.Context, params ChatParams) (*ChatResponse, error) {
	chatReq := chatRequest{
		Query:        params.Query,
		Stream:       false,
		SystemPrompt: params.SystemPrompt,
		Filters:      params.Filters,
	}

	body, err := json.Marshal(chatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal chat request: %w", err)
	}

	resp, err := c.doRequest(ctx, "POST", "/api/v1/chat", nil, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}

	var result ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// StreamedChat performs a streaming chat query
func (c *Client) StreamedChat(ctx context.Context, params ChatParams) (<-chan ChatStreamEvent, <-chan error) {
	eventChan := make(chan ChatStreamEvent)
	errChan := make(chan error, 1)

	go func() {
		defer close(eventChan)
		defer close(errChan)

		chatReq := chatRequest{
			Query:        params.Query,
			Stream:       true,
			SystemPrompt: params.SystemPrompt,
			Filters:      params.Filters,
		}

		body, err := json.Marshal(chatReq)
		if err != nil {
			errChan <- fmt.Errorf("failed to marshal chat request: %w", err)
			return
		}

		resp, err := c.doRequest(ctx, "POST", "/api/v1/chat", nil, bytes.NewReader(body))
		if err != nil {
			errChan <- err
			return
		}
		defer func() { _ = resp.Body.Close() }()

		if err := c.checkResponse(resp); err != nil {
			errChan <- err
			return
		}

		if err := c.parseSSEStream(resp.Body, eventChan); err != nil {
			errChan <- err
			return
		}
	}()

	return eventChan, errChan
}

// doRequest performs an HTTP request
func (c *Client) doRequest(ctx context.Context, method, path string, params url.Values, body io.Reader) (*http.Response, error) {
	urlStr := c.baseURL + path
	if len(params) > 0 {
		urlStr += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, urlStr, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}

// checkResponse checks if the HTTP response indicates an error
func (c *Client) checkResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("skald API error (%d): %s", resp.StatusCode, string(bodyBytes))
}

// parseSSEStream parses Server-Sent Events stream
func (c *Client) parseSSEStream(body io.Reader, eventChan chan<- ChatStreamEvent) error {
	scanner := bufio.NewScanner(body)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines and ping lines
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		// Parse data lines
		if after, ok := strings.CutPrefix(line, "data: "); ok {
			var event ChatStreamEvent
			if err := json.Unmarshal([]byte(after), &event); err != nil {
				// Skip invalid JSON
				continue
			}

			eventChan <- event

			// Stop on 'done' event
			if event.Type == "done" {
				return nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stream: %w", err)
	}

	return nil
}
