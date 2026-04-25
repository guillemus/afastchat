package agents

import (
	"context"
	"encoding/json"
	"fmt"

	"afastchat.com/exa"
)

var webSearchTool = Tool{
	Name:        "web_search",
	Description: "Search the web for relevant results.",
	Parameters: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "Search query.",
			},
		},
		"required":             []string{"query"},
		"additionalProperties": false,
	},
	Strict: true,
}

type webSearchArgs struct {
	Query string `json:"query"`
}

func (x *Agent) webSearch(ctx context.Context, args webSearchArgs) (string, error) {
	resp, err := x.exaClient.Search(ctx, &exa.SearchRequest{
		Query:      args.Query,
		Type:       "auto",
		NumResults: 5,
		Contents: &exa.ContentsOptions{
			Highlights: &exa.HighlightsOptions{
				Query:         args.Query,
				MaxCharacters: 400,
			},
			Summary: &exa.SummaryOptions{
				Query: args.Query,
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("search web: %w", err)
	}

	result, err := json.Marshal(resp)
	if err != nil {
		return "", fmt.Errorf("encode web_search result: %w", err)
	}

	return string(result), nil
}
