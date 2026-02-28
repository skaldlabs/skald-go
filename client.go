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
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Client is the main Skald SDK client
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	tracer     trace.Tracer
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

// NewClientWithOptions creates a new Skald client with functional options.
// Use this constructor to enable OpenTelemetry tracing or provide a custom HTTP client.
func NewClientWithOptions(apiKey string, baseURL string, opts ...Option) *Client {
	if baseURL == "" {
		baseURL = "https://api.useskald.com"
	}
	baseURL = strings.TrimRight(baseURL, "/")

	cfg := &clientConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	httpClient := cfg.httpClient
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	var tracer trace.Tracer
	if cfg.tracerProvider != nil {
		tracer = cfg.tracerProvider.Tracer("github.com/skaldlabs/skald-go")
		transport := httpClient.Transport
		if transport == nil {
			transport = http.DefaultTransport
		}
		httpClient.Transport = otelhttp.NewTransport(transport, otelhttp.WithTracerProvider(cfg.tracerProvider))
	}

	return &Client{
		apiKey:     apiKey,
		baseURL:    baseURL,
		httpClient: httpClient,
		tracer:     tracer,
	}
}

// startSpan starts a new span if tracing is enabled, otherwise returns a no-op span.
func (c *Client) startSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	if c.tracer == nil {
		return ctx, trace.SpanFromContext(ctx)
	}
	return c.tracer.Start(ctx, name)
}

// CreateMemo creates a new memo
func (c *Client) CreateMemo(ctx context.Context, memoData MemoData) (*CreateMemoResponse, error) {
	ctx, span := c.startSpan(ctx, "skald.CreateMemo")
	defer span.End()

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
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := c.checkResponse(resp); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	var result CreateMemoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	span.SetAttributes(attribute.String("skald.memo_uuid", result.MemoUUID.String()))
	return &result, nil
}

// CreateMemoFromFile creates a new memo by uploading a file
// Supported file formats: PDF, DOC, DOCX, PPTX
// Maximum file size: 100MB
func (c *Client) CreateMemoFromFile(ctx context.Context, filePath string, memoData *MemoFileData) (*CreateMemoResponse, error) {
	ctx, span := c.startSpan(ctx, "skald.CreateMemoFromFile")
	defer span.End()

	span.SetAttributes(attribute.String("skald.file_name", filepath.Base(filePath)))

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
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

	// Add memo data fields if provided
	if memoData != nil {
		// Add title field
		if memoData.Title != nil {
			if err := writer.WriteField("title", *memoData.Title); err != nil {
				return nil, fmt.Errorf("failed to write title field: %w", err)
			}
		}

		// Add source field
		if memoData.Source != nil {
			if err := writer.WriteField("source", *memoData.Source); err != nil {
				return nil, fmt.Errorf("failed to write source field: %w", err)
			}
		}

		// Add reference_id field
		if memoData.ReferenceID != nil {
			if err := writer.WriteField("reference_id", *memoData.ReferenceID); err != nil {
				return nil, fmt.Errorf("failed to write reference_id field: %w", err)
			}
		}

		// Add tags as JSON array
		if len(memoData.Tags) > 0 {
			tagsJSON, err := json.Marshal(memoData.Tags)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal tags: %w", err)
			}
			if err := writer.WriteField("tags", string(tagsJSON)); err != nil {
				return nil, fmt.Errorf("failed to write tags field: %w", err)
			}
		}

		// Add metadata as JSON
		if len(memoData.Metadata) > 0 {
			metadataJSON, err := json.Marshal(memoData.Metadata)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal metadata: %w", err)
			}
			if err := writer.WriteField("metadata", string(metadataJSON)); err != nil {
				return nil, fmt.Errorf("failed to write metadata field: %w", err)
			}
		}

		// Add expiration_date field (RFC3339 format)
		if memoData.ExpirationDate != nil {
			if err := writer.WriteField("expiration_date", memoData.ExpirationDate.Format(time.RFC3339)); err != nil {
				return nil, fmt.Errorf("failed to write expiration_date field: %w", err)
			}
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create request
	urlStr := c.baseURL + "/api/v1/memo"
	req, err := http.NewRequestWithContext(ctx, "POST", urlStr, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if err := c.checkResponse(resp); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	var result CreateMemoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	span.SetAttributes(attribute.String("skald.memo_uuid", result.MemoUUID.String()))
	return &result, nil
}

// GetMemo retrieves a memo by ID
func (c *Client) GetMemo(ctx context.Context, memoID string, idType ...IDType) (*Memo, error) {
	ctx, span := c.startSpan(ctx, "skald.GetMemo")
	defer span.End()

	idTypeValue := IDTypeMemoUUID
	if len(idType) > 0 {
		idTypeValue = idType[0]
		if idTypeValue != IDTypeMemoUUID && idTypeValue != IDTypeReferenceID {
			return nil, fmt.Errorf("invalid idType: must be 'memo_uuid' or 'reference_id'")
		}
	}

	span.SetAttributes(
		attribute.String("skald.memo_id", memoID),
		attribute.String("skald.id_type", string(idTypeValue)),
	)

	params := url.Values{}
	if idTypeValue != IDTypeMemoUUID {
		params.Set("id_type", string(idTypeValue))
	}

	path := fmt.Sprintf("/api/v1/memo/%s", url.PathEscape(memoID))
	resp, err := c.doRequest(ctx, "GET", path, params, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := c.checkResponse(resp); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
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
	ctx, span := c.startSpan(ctx, "skald.ListMemos")
	defer span.End()

	queryParams := url.Values{}
	if params != nil {
		if params.Page != nil {
			span.SetAttributes(attribute.Int("skald.page", *params.Page))
			queryParams.Set("page", fmt.Sprintf("%d", *params.Page))
		}
		if params.PageSize != nil {
			span.SetAttributes(attribute.Int("skald.page_size", *params.PageSize))
			queryParams.Set("page_size", fmt.Sprintf("%d", *params.PageSize))
		}
	}

	resp, err := c.doRequest(ctx, "GET", "/api/v1/memo", queryParams, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := c.checkResponse(resp); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
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
	ctx, span := c.startSpan(ctx, "skald.UpdateMemo")
	defer span.End()

	idTypeValue := IDTypeMemoUUID
	if len(idType) > 0 {
		idTypeValue = idType[0]
		if idTypeValue != IDTypeMemoUUID && idTypeValue != IDTypeReferenceID {
			return nil, fmt.Errorf("invalid idType: must be 'memo_uuid' or 'reference_id'")
		}
	}

	span.SetAttributes(
		attribute.String("skald.memo_id", memoID),
		attribute.String("skald.id_type", string(idTypeValue)),
	)

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
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := c.checkResponse(resp); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
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
	ctx, span := c.startSpan(ctx, "skald.DeleteMemo")
	defer span.End()

	idTypeValue := IDTypeMemoUUID
	if len(idType) > 0 {
		idTypeValue = idType[0]
		if idTypeValue != IDTypeMemoUUID && idTypeValue != IDTypeReferenceID {
			return fmt.Errorf("invalid idType: must be 'memo_uuid' or 'reference_id'")
		}
	}

	span.SetAttributes(
		attribute.String("skald.memo_id", memoID),
		attribute.String("skald.id_type", string(idTypeValue)),
	)

	params := url.Values{}
	if idTypeValue != IDTypeMemoUUID {
		params.Set("id_type", string(idTypeValue))
	}

	path := fmt.Sprintf("/api/v1/memo/%s", url.PathEscape(memoID))
	resp, err := c.doRequest(ctx, "DELETE", path, params, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := c.checkResponse(resp); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

// CheckMemoStatus checks the processing status of a memo
// The memo can be identified by UUID (default) or reference ID
func (c *Client) CheckMemoStatus(ctx context.Context, memoID string, idType ...IDType) (*MemoStatusResponse, error) {
	ctx, span := c.startSpan(ctx, "skald.CheckMemoStatus")
	defer span.End()

	idTypeValue := IDTypeMemoUUID
	if len(idType) > 0 {
		idTypeValue = idType[0]
		if idTypeValue != IDTypeMemoUUID && idTypeValue != IDTypeReferenceID {
			return nil, fmt.Errorf("invalid idType: must be 'memo_uuid' or 'reference_id'")
		}
	}

	span.SetAttributes(
		attribute.String("skald.memo_id", memoID),
		attribute.String("skald.id_type", string(idTypeValue)),
	)

	params := url.Values{}
	if idTypeValue != IDTypeMemoUUID {
		params.Set("id_type", string(idTypeValue))
	}

	path := fmt.Sprintf("/api/v1/memo/%s/status", url.PathEscape(memoID))
	resp, err := c.doRequest(ctx, "GET", path, params, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := c.checkResponse(resp); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	var status MemoStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	span.SetAttributes(attribute.String("skald.status", string(status.Status)))
	return &status, nil
}

// WaitForMemoReady polls the memo status until it's processed or an error occurs.
// It returns when the memo is processed, or an error if processing fails or context is cancelled.
// The pollInterval specifies how long to wait between status checks.
func (c *Client) WaitForMemoReady(ctx context.Context, memoID string, pollInterval time.Duration, idType ...IDType) error {
	ctx, span := c.startSpan(ctx, "skald.WaitForMemoReady")
	defer span.End()

	idTypeValue := IDTypeMemoUUID
	if len(idType) > 0 {
		idTypeValue = idType[0]
	}

	span.SetAttributes(
		attribute.String("skald.memo_id", memoID),
		attribute.String("skald.id_type", string(idTypeValue)),
	)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		status, err := c.CheckMemoStatus(ctx, memoID, idType...)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}

		switch status.Status {
		case MemoStatusProcessed:
			return nil
		case MemoStatusError:
			errMsg := "memo processing failed"
			if status.ErrorReason != nil {
				errMsg = *status.ErrorReason
			}
			waitErr := fmt.Errorf("%s", errMsg)
			span.RecordError(waitErr)
			span.SetStatus(codes.Error, waitErr.Error())
			return waitErr
		case MemoStatusProcessing:
			// Continue polling
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Continue to next iteration
		}
	}
}

// Search searches for memos
func (c *Client) Search(ctx context.Context, searchReq SearchRequest) (*SearchResponse, error) {
	ctx, span := c.startSpan(ctx, "skald.Search")
	defer span.End()

	span.SetAttributes(attribute.Int("skald.query_length", len(searchReq.Query)))

	body, err := json.Marshal(searchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search request: %w", err)
	}

	resp, err := c.doRequest(ctx, "POST", "/api/v1/search", nil, bytes.NewReader(body))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := c.checkResponse(resp); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	span.SetAttributes(attribute.Int("skald.result_count", len(result.Results)))
	return &result, nil
}

// Chat performs a non-streaming chat query and returns the response
func (c *Client) Chat(ctx context.Context, params ChatParams) (*ChatResponse, error) {
	ctx, span := c.startSpan(ctx, "skald.Chat")
	defer span.End()

	span.SetAttributes(
		attribute.Int("skald.query_length", len(params.Query)),
		attribute.String("skald.chat_id", params.ChatID),
		attribute.Bool("skald.has_rag_config", params.RAGConfig != nil),
	)

	chatReq := chatRequest{
		Query:        params.Query,
		Stream:       false,
		SystemPrompt: params.SystemPrompt,
		Filters:      params.Filters,
		ChatID:       params.ChatID,
		RAGConfig:    params.RAGConfig,
	}

	body, err := json.Marshal(chatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal chat request: %w", err)
	}

	resp, err := c.doRequest(ctx, "POST", "/api/v1/chat", nil, bytes.NewReader(body))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := c.checkResponse(resp); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
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

	ctx, span := c.startSpan(ctx, "skald.StreamedChat")

	span.SetAttributes(
		attribute.Int("skald.query_length", len(params.Query)),
		attribute.String("skald.chat_id", params.ChatID),
		attribute.Bool("skald.has_rag_config", params.RAGConfig != nil),
	)

	go func() {
		defer span.End()
		defer close(eventChan)
		defer close(errChan)

		chatReq := chatRequest{
			Query:        params.Query,
			Stream:       true,
			SystemPrompt: params.SystemPrompt,
			Filters:      params.Filters,
			ChatID:       params.ChatID,
			RAGConfig:    params.RAGConfig,
		}

		body, err := json.Marshal(chatReq)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			errChan <- fmt.Errorf("failed to marshal chat request: %w", err)
			return
		}

		resp, err := c.doRequest(ctx, "POST", "/api/v1/chat", nil, bytes.NewReader(body))
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			errChan <- err
			return
		}
		defer func() { _ = resp.Body.Close() }()

		if err := c.checkResponse(resp); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			errChan <- err
			return
		}

		if err := c.parseSSEStream(resp.Body, eventChan); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			errChan <- err
			return
		}
	}()

	return eventChan, errChan
}

// GetChat retrieves a persisted chat conversation by its ID
func (c *Client) GetChat(ctx context.Context, chatID string) (*GetChatResponse, error) {
	ctx, span := c.startSpan(ctx, "skald.GetChat")
	defer span.End()

	span.SetAttributes(attribute.String("skald.chat_id", chatID))

	path := fmt.Sprintf("/api/v1/chat/%s", url.PathEscape(chatID))
	resp, err := c.doRequest(ctx, "GET", path, nil, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := c.checkResponse(resp); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	var result GetChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
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
	return &APIError{
		StatusCode: resp.StatusCode,
		Message:    string(bodyBytes),
	}
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
