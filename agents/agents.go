package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"log/slog"
	"strings"

	"afastchat.com/db"

	"github.com/zendev-sh/goai"
	"github.com/zendev-sh/goai/provider"
	"github.com/zendev-sh/goai/provider/openrouter"
)

const (
	chatModel  = "anthropic/claude-haiku-5.5"
	titleModel = "anthropic/claude-haiku-5.5"
)

type Agent struct {
	model      provider.LanguageModel
	titleModel provider.LanguageModel
}

func NewAgent(apiKey string) *Agent {
	return &Agent{
		model:      openrouter.Chat(chatModel, openrouter.WithAPIKey(apiKey)),
		titleModel: openrouter.Chat(titleModel, openrouter.WithAPIKey(apiKey)),
	}
}

func MapDBMessages(dbMessages []db.Message) []provider.Message {
	messages := make([]provider.Message, 0, len(dbMessages))
	for _, it := range dbMessages {
		if it.Content == "" {
			continue
		}

		switch it.Type {
		case db.UserMessage:
			messages = append(messages, goai.UserMessage(it.Content))
		case db.AssistantMessage:
			messages = append(messages, goai.AssistantMessage(it.Content))
		case db.SystemMessage:
			messages = append(messages, goai.SystemMessage(it.Content))
		case db.AssistantToolCall:
			messages = append(messages, provider.Message{
				Role: provider.RoleAssistant,
				Content: []provider.Part{{
					Type:       provider.PartToolCall,
					ToolCallID: it.ToolCallID,
					ToolName:   it.ToolName,
					ToolInput:  json.RawMessage(it.Content),
				}},
			})
		case db.ToolOutput:
			messages = append(messages, goai.ToolMessage(it.ToolCallID, it.ToolName, it.Content))
		case db.AssistantReasoning, db.AssistantSources:
		}
	}

	return messages
}

var sysPrompt = strings.TrimSpace(`
You are a helpful assistant. Before each tool call, briefly tell the user what you are about to do in one short sentence.

When you use web sources, cite them inline as markdown links placed right after the claim they support. Do not add a separate sources list at the end; the app shows one.

When writing math, use LaTeX inside $...$ for inline math and $$...$$ for display math. Do not put LaTeX math in code blocks unless the user asks for source syntax.
`)

func (x *Agent) Stream(ctx context.Context, messages []db.Message) iter.Seq2[StreamEvent, error] {
	return func(yield func(StreamEvent, error) bool) {
		streamCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		stream, err := goai.StreamText(streamCtx, x.model,
			goai.WithSystem(sysPrompt),
			goai.WithMessages(MapDBMessages(messages)...),
			goai.WithTools(serverTools...),
			goai.WithMaxSteps(30),
			goai.WithMaxOutputTokens(32768),
			goai.WithProviderOptions(map[string]any{
				"reasoning": map[string]any{"effort": "medium"},
			}),
		)
		if err != nil {
			yield(nil, err)
			return
		}

		var streamErr error
		for chunk := range stream.Stream() {
			switch chunk.Type {
			case provider.ChunkText:
				if source, ok := chunk.Metadata["source"].(provider.Source); ok && source.URL != "" {
					if !yield(Source{URL: source.URL, Title: source.Title}, nil) {
						return
					}
				}
				if chunk.Text != "" && !yield(AssistantDelta{Delta: chunk.Text}, nil) {
					return
				}
			case provider.ChunkReasoning:
				if chunk.Text != "" && !yield(ReasoningDelta{Delta: chunk.Text}, nil) {
					return
				}
			case provider.ChunkToolCall:
				if !yield(Toolcall{CallID: chunk.ToolCallID, Name: chunk.ToolName, Args: chunk.ToolInput}, nil) {
					return
				}
			case provider.ChunkStepFinish:
				if results, ok := chunk.Metadata["toolResults"].([]provider.ToolResult); ok {
					for _, result := range results {
						if !yield(ToolOutput{CallID: result.ToolCallID, Name: result.ToolName, Output: result.Output}, nil) {
							return
						}
					}
					continue
				}
				if chunk.Usage.InputTokens != 0 || chunk.Usage.OutputTokens != 0 {
					model := chunk.Response.Model
					if model == "" {
						model = x.model.ModelID()
					}
					if !yield(UsageDelta{Usage: Usage{
						Provider:          "openrouter",
						Model:             model,
						InputTokens:       int64(chunk.Usage.InputTokens),
						CachedInputTokens: int64(chunk.Usage.CacheReadTokens),
						OutputTokens:      int64(chunk.Usage.OutputTokens),
						ReasoningTokens:   int64(chunk.Usage.ReasoningTokens),
					}}, nil) {
						return
					}
				}
			case provider.ChunkError:
				if streamErr == nil {
					streamErr = chunk.Error
				}
			case provider.ChunkToolCallDelta, provider.ChunkToolCallStreamStart, provider.ChunkToolResult:
			case provider.ChunkFinish:
				if chunk.StoppedBy == provider.StopCauseMaxSteps && streamErr == nil {
					streamErr = fmt.Errorf("too many tool rounds")
				}
			}
		}

		if streamErr == nil {
			streamErr = stream.Err()
		}
		if streamErr != nil {
			yield(nil, streamErr)
		}
	}
}

var serverTools = []goai.Tool{
	{
		Name:                "web_search",
		ProviderDefinedType: "openrouter:web_search",
		ProviderDefinedOptions: map[string]any{
			"parameters": map[string]any{"max_results": 8},
		},
	},
	{
		Name:                "web_fetch",
		ProviderDefinedType: "openrouter:web_fetch",
		ProviderDefinedOptions: map[string]any{
			"parameters": map[string]any{"max_content_tokens": 20000},
		},
	},
	{
		Name:                "datetime",
		ProviderDefinedType: "openrouter:datetime",
	},
}

func (x *Agent) GenerateTitleFromUserMessage(ctx context.Context, userMessage string) string {
	resp, err := goai.GenerateText(ctx, x.titleModel,
		goai.WithSystem("Generate a concise title (max 6 words) for a chat thread based on the user's first message. Return only the title, no quotes or punctuation."),
		goai.WithPrompt(userMessage),
	)
	if err != nil {
		slog.Error("error generating title", "err", err)
		return "New Chat"
	}

	title := strings.TrimSpace(resp.Text)
	if title == "" {
		return "New Chat"
	}
	return title
}
