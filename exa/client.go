package exa

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
)

const (
	DefaultBaseURL = "https://api.exa.ai"
)

type Client struct {
	rc     *resty.Client
	apiKey string
}

func NewClient(apiKey string) *Client {
	rc := resty.New()
	rc.SetBaseURL(DefaultBaseURL)
	rc.SetHeader("x-api-key", apiKey)
	rc.SetHeader("Content-Type", "application/json")

	return &Client{
		rc:     rc,
		apiKey: apiKey,
	}
}

// Search performs a search with Exa's prompt-engineered query.
// Supports neural, fast, auto, deep, deep-reasoning, and instant search types.
// Optionally retrieves contents with text, highlights, and summaries.
func (c *Client) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	var resp SearchResponse

	r, err := c.rc.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&resp).
		Post("/search")

	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}

	if r.IsError() {
		return nil, fmt.Errorf("search request failed with status %d: %s", r.StatusCode(), r.String())
	}

	return &resp, nil
}

// FindSimilar finds similar links to the provided URL.
// Optionally retrieves contents of the similar results.
func (c *Client) FindSimilar(ctx context.Context, req *FindSimilarRequest) (*FindSimilarResponse, error) {
	var resp FindSimilarResponse

	r, err := c.rc.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&resp).
		Post("/findSimilar")

	if err != nil {
		return nil, fmt.Errorf("findSimilar request failed: %w", err)
	}

	if r.IsError() {
		return nil, fmt.Errorf("findSimilar request failed with status %d: %s", r.StatusCode(), r.String())
	}

	return &resp, nil
}

// GetContents retrieves contents from provided URLs.
// Supports text extraction, highlights, summaries, and subpage crawling.
func (c *Client) GetContents(ctx context.Context, req *ContentsRequest) (*ContentsResponse, error) {
	var resp ContentsResponse

	r, err := c.rc.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&resp).
		Post("/contents")

	if err != nil {
		return nil, fmt.Errorf("contents request failed: %w", err)
	}

	if r.IsError() {
		return nil, fmt.Errorf("contents request failed with status %d: %s", r.StatusCode(), r.String())
	}

	return &resp, nil
}

// Answer performs a search and generates either a direct answer or detailed summary with citations.
// Supports streaming responses and structured output schemas.
func (c *Client) Answer(ctx context.Context, req *AnswerRequest) (*AnswerResponse, error) {
	var resp AnswerResponse

	r, err := c.rc.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&resp).
		Post("/answer")

	if err != nil {
		return nil, fmt.Errorf("answer request failed: %w", err)
	}

	if r.IsError() {
		return nil, fmt.Errorf("answer request failed with status %d: %s", r.StatusCode(), r.String())
	}

	return &resp, nil
}

// CreateResearchTask creates a new research task with instructions and optional output schema.
// Research tasks run asynchronously and can be retrieved later by ID.
func (c *Client) CreateResearchTask(ctx context.Context, req *CreateResearchTaskRequest) (*ResearchTask, error) {
	var resp ResearchTask

	r, err := c.rc.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&resp).
		Post("/research/v0/tasks")

	if err != nil {
		return nil, fmt.Errorf("create research task failed: %w", err)
	}

	if r.IsError() {
		return nil, fmt.Errorf("create research task failed with status %d: %s", r.StatusCode(), r.String())
	}

	return &resp, nil
}

// ListResearchTasks retrieves a paginated list of research tasks.
// Use cursor and limit for pagination.
func (c *Client) ListResearchTasks(ctx context.Context, cursor string, limit int) (*ListResearchTasksResponse, error) {
	var resp ListResearchTasksResponse

	req := c.rc.R().
		SetContext(ctx).
		SetResult(&resp)

	if cursor != "" {
		req.SetQueryParam("cursor", cursor)
	}
	if limit > 0 {
		req.SetQueryParam("limit", fmt.Sprintf("%d", limit))
	}

	r, err := req.Get("/research/v0/tasks")

	if err != nil {
		return nil, fmt.Errorf("list research tasks failed: %w", err)
	}

	if r.IsError() {
		return nil, fmt.Errorf("list research tasks failed with status %d: %s", r.StatusCode(), r.String())
	}

	return &resp, nil
}

// GetResearchTask retrieves a specific research task by its ID.
func (c *Client) GetResearchTask(ctx context.Context, taskID string) (*ResearchTask, error) {
	var resp ResearchTask

	r, err := c.rc.R().
		SetContext(ctx).
		SetResult(&resp).
		Get(fmt.Sprintf("/research/v0/tasks/%s", taskID))

	if err != nil {
		return nil, fmt.Errorf("get research task failed: %w", err)
	}

	if r.IsError() {
		return nil, fmt.Errorf("get research task failed with status %d: %s", r.StatusCode(), r.String())
	}

	return &resp, nil
}
