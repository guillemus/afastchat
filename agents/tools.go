package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"afastchat.com/exa"
)

var webSearchTool = Tool{
	Name:        "web_search",
	Description: "Search the web for relevant pages and optional excerpts. Use this for general research, source discovery, and domain/date filtered search. Prefer includeDomains when the user asks for official or trusted sources. Set highlights when you need evidence snippets from pages without reading each full URL.",
	Parameters: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "Search query or research question.",
			},
			"numResults": map[string]any{
				"type":        "integer",
				"description": "Number of results to return. Defaults to 5. Use 10 for broader research.",
			},
			"includeDomains": map[string]any{
				"type":        "array",
				"description": "Only return results from these domains, for example [\"docs.exa.ai\"].",
				"items":       map[string]any{"type": "string"},
			},
			"excludeDomains": map[string]any{
				"type":        "array",
				"description": "Exclude results from these domains.",
				"items":       map[string]any{"type": "string"},
			},
			"startPublishedDate": map[string]any{
				"type":        "string",
				"description": "Only include pages published after this ISO 8601 date, for example 2026-01-01.",
			},
			"endPublishedDate": map[string]any{
				"type":        "string",
				"description": "Only include pages published before this ISO 8601 date, for example 2026-05-03.",
			},
			"category": map[string]any{
				"type":        "string",
				"description": "Optional result category.",
				"enum":        []string{"company", "research paper", "news", "personal site", "people", "financial report"},
			},
			"highlights": map[string]any{
				"type":        "boolean",
				"description": "Return relevant excerpts from each result. Use when answer needs evidence snippets.",
			},
			"highlightQuery": map[string]any{
				"type":        "string",
				"description": "Focus query for excerpts. Defaults to query.",
			},
			"maxCharacters": map[string]any{
				"type":        "integer",
				"description": "Maximum highlight characters per result. Defaults to 400.",
			},
		},
		"required":             []string{"query"},
		"additionalProperties": false,
	},
	Strict: true,
}

var webReadTool = Tool{
	Name:        "web_read",
	Description: "Read or extract content from known URLs. Use this when the user provides URLs, or after web_search when specific results need deeper inspection. Can return focused highlights, summaries, or full text. Some sites may block crawlers; failed URLs include status details.",
	Parameters: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"urls": map[string]any{
				"type":        "array",
				"description": "URLs to fetch and extract.",
				"items":       map[string]any{"type": "string"},
			},
			"mode": map[string]any{
				"type":        "string",
				"description": "Extraction mode. highlights is best for agent research, summary for quick overview, text for full page content.",
				"enum":        []string{"highlights", "summary", "text"},
			},
			"query": map[string]any{
				"type":        "string",
				"description": "Focus for highlights or summary.",
			},
			"maxCharacters": map[string]any{
				"type":        "integer",
				"description": "Maximum characters for highlights or text. Defaults to 1000.",
			},
		},
		"required":             []string{"urls"},
		"additionalProperties": false,
	},
	Strict: true,
}

var webSearchRecentTool = Tool{
	Name:        "web_search_recent",
	Description: "Search for recent or current information. Use this for latest news, recent releases, current status, or time-sensitive questions. Internally searches pages published within the requested recent window.",
	Parameters: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "Search query about recent or current information.",
			},
			"daysBack": map[string]any{
				"type":        "integer",
				"description": "How many days back to search. Defaults to 30.",
			},
			"numResults": map[string]any{
				"type":        "integer",
				"description": "Number of results to return. Defaults to 5.",
			},
			"category": map[string]any{
				"type":        "string",
				"description": "Optional result category. Use news for current events.",
				"enum":        []string{"news", "company", "research paper", "personal site", "people", "financial report"},
			},
		},
		"required":             []string{"query"},
		"additionalProperties": false,
	},
	Strict: true,
}

type WebSearchArgs struct {
	Query              string   `json:"query"`
	NumResults         int      `json:"numResults"`
	IncludeDomains     []string `json:"includeDomains"`
	ExcludeDomains     []string `json:"excludeDomains"`
	StartPublishedDate string   `json:"startPublishedDate"`
	EndPublishedDate   string   `json:"endPublishedDate"`
	Category           string   `json:"category"`
	Highlights         bool     `json:"highlights"`
	HighlightQuery     string   `json:"highlightQuery"`
	MaxCharacters      int      `json:"maxCharacters"`
}

type WebReadArgs struct {
	URLs          []string `json:"urls"`
	Mode          string   `json:"mode"`
	Query         string   `json:"query"`
	MaxCharacters int      `json:"maxCharacters"`
}

type WebSearchRecentArgs struct {
	Query      string `json:"query"`
	DaysBack   int    `json:"daysBack"`
	NumResults int    `json:"numResults"`
	Category   string `json:"category"`
}

func (x *Agent) webSearch(ctx context.Context, args WebSearchArgs) (string, error) {
	if args.NumResults == 0 {
		args.NumResults = 5
	}
	if args.MaxCharacters == 0 {
		args.MaxCharacters = 400
	}
	if args.HighlightQuery == "" {
		args.HighlightQuery = args.Query
	}

	req := &exa.SearchRequest{
		Query:              args.Query,
		Type:               "instant",
		NumResults:         args.NumResults,
		IncludeDomains:     args.IncludeDomains,
		ExcludeDomains:     args.ExcludeDomains,
		StartPublishedDate: args.StartPublishedDate,
		EndPublishedDate:   args.EndPublishedDate,
		Category:           args.Category,
	}
	if args.Highlights {
		req.Contents = &exa.ContentsOptions{
			Highlights: &exa.HighlightsOptions{
				Query:         args.HighlightQuery,
				MaxCharacters: args.MaxCharacters,
			},
		}
	}

	resp, err := x.exaClient.Search(ctx, req)
	if err != nil {
		return "", fmt.Errorf("search web: %w", err)
	}

	result, err := json.Marshal(resp)
	if err != nil {
		return "", fmt.Errorf("encode web_search result: %w", err)
	}

	return string(result), nil
}

func (x *Agent) webRead(ctx context.Context, args WebReadArgs) (string, error) {
	if args.Mode == "" {
		args.Mode = "highlights"
	}
	if args.MaxCharacters == 0 {
		args.MaxCharacters = 1000
	}

	req := &exa.ContentsRequest{URLs: args.URLs}
	switch args.Mode {
	case "summary":
		req.Summary = &exa.SummaryOptions{Query: args.Query}
	case "text":
		req.Text = &exa.TextOptions{MaxCharacters: args.MaxCharacters}
	default:
		req.Highlights = &exa.HighlightsOptions{Query: args.Query, MaxCharacters: args.MaxCharacters}
	}

	resp, err := x.exaClient.GetContents(ctx, req)
	if err != nil {
		return "", fmt.Errorf("read web: %w", err)
	}

	result, err := json.Marshal(resp)
	if err != nil {
		return "", fmt.Errorf("encode web_read result: %w", err)
	}

	return string(result), nil
}

func (x *Agent) webSearchRecent(ctx context.Context, args WebSearchRecentArgs) (string, error) {
	if args.DaysBack == 0 {
		args.DaysBack = 30
	}
	if args.NumResults == 0 {
		args.NumResults = 5
	}

	resp, err := x.exaClient.Search(ctx, &exa.SearchRequest{
		Query:              args.Query,
		Type:               "auto",
		NumResults:         args.NumResults,
		StartPublishedDate: time.Now().AddDate(0, 0, -args.DaysBack).Format("2006-01-02"),
		Category:           args.Category,
		Contents: &exa.ContentsOptions{
			Highlights: &exa.HighlightsOptions{
				Query:         args.Query,
				MaxCharacters: 400,
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("search recent web: %w", err)
	}

	result, err := json.Marshal(resp)
	if err != nil {
		return "", fmt.Errorf("encode web_search_recent result: %w", err)
	}

	return string(result), nil
}
