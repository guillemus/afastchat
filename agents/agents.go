package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"log/slog"
	"strings"

	"afastchat.com/db"
	"afastchat.com/env"
	"afastchat.com/exa"

	"github.com/zendev-sh/goai"
	"github.com/zendev-sh/goai/provider"
	"github.com/zendev-sh/goai/provider/fireworks"
	"github.com/zendev-sh/goai/provider/openai"
)

const (
	chatModel  = "accounts/fireworks/models/kimi-k2p6"
	titleModel = "gpt-5.4-nano"
)

type Agent struct {
	model      provider.LanguageModel
	titleModel provider.LanguageModel
	exaClient  *exa.Client
}

func NewAgent(env env.Env) *Agent {
	return &Agent{
		model:      fireworks.Chat(chatModel, fireworks.WithAPIKey(env.FireworksAPIKey)),
		titleModel: openai.Chat(titleModel, openai.WithAPIKey(env.OpenAIAPIKey)),
		exaClient:  exa.NewClient(env.ExaAPIKey),
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
		}
	}

	return messages
}

var sysPrompt = strings.TrimSpace(`
You are a helpful assistant. Before each tool call, briefly tell the user what you are about to do in one short sentence.

When writing math, use LaTeX inside $...$ for inline math and $$...$$ for display math. Do not put LaTeX math in code blocks unless the user asks for source syntax.
`)

func (x *Agent) Stream(ctx context.Context, messages []db.Message) iter.Seq2[StreamEvent, error] {
	return func(yield func(StreamEvent, error) bool) {
		streamCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		stream, err := goai.StreamText(streamCtx, x.model,
			goai.WithSystem(sysPrompt),
			goai.WithMessages(MapDBMessages(messages)...),
			goai.WithTools(x.tools()...),
			goai.WithMaxSteps(30),
			goai.WithMaxOutputTokens(32768),
			goai.WithTemperature(0.6),
			goai.WithTopP(1),
			goai.WithTopK(40),
			goai.WithPresencePenalty(0),
			goai.WithFrequencyPenalty(0),
			goai.WithProviderOptions(map[string]any{
				"reasoning_effort": "medium",
				"strictJsonSchema": true,
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
						Provider:          "fireworks",
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

func (x *Agent) tools() []goai.Tool {
	tools := []Tool{webSearchTool, webReadTool, webSearchRecentTool}
	result := make([]goai.Tool, 0, len(tools))
	for _, tool := range tools {
		schema, err := json.Marshal(tool.Parameters)
		if err != nil {
			panic(fmt.Sprintf("marshal tool schema %q: %v", tool.Name, err))
		}
		tool := tool
		result = append(result, goai.Tool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: schema,
			Execute: func(ctx context.Context, input json.RawMessage) (string, error) {
				output, err := x.runTool(ctx, tool.Name, string(input))
				if err != nil {
					return fmt.Sprintf(`{"error":%q}`, err.Error()), nil
				}
				return output, nil
			},
		})
	}
	return result
}

func (x *Agent) runTool(ctx context.Context, name, rawArgs string) (string, error) {
	switch name {
	case "web_search":
		var args WebSearchArgs
		if err := json.Unmarshal([]byte(rawArgs), &args); err != nil {
			return "", err
		}
		return x.webSearch(ctx, args)
	case "web_read":
		var args WebReadArgs
		if err := json.Unmarshal([]byte(rawArgs), &args); err != nil {
			return "", err
		}
		return x.webRead(ctx, args)
	case "web_search_recent":
		var args WebSearchRecentArgs
		if err := json.Unmarshal([]byte(rawArgs), &args); err != nil {
			return "", err
		}
		return x.webSearchRecent(ctx, args)
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
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
