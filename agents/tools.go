package agents

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3/responses"
)

var websearchToolDef = responses.ToolParamOfFunction("web_search", map[string]any{
	"type": "object",
	"properties": map[string]any{
		"query": map[string]any{
			"type":        "string",
			"description": "Search query.",
		},
	},
	"required":             []string{"query"},
	"additionalProperties": false,
}, true)

func (x *Agent) webSearch(ctx context.Context, rawArgs string) (string, error) {
	_ = ctx
	_ = rawArgs

	return "{}", fmt.Errorf("TODO")
}
