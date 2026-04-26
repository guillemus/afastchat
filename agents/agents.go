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

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type Agent struct {
	provider   Provider
	fastClient openai.Client
	exaClient  *exa.Client
}

func NewAgent(env env.Env) *Agent {
	// provider := NewFireworksChatProvider(env.FireworksAPIKey, "accounts/fireworks/models/kimi-k2p6")
	provider := NewOpenAIResponsesProvider(env.OpenAIAPIKey, openai.ChatModelGPT5_4Nano)

	fastClient := openai.NewClient(option.WithAPIKey(env.OpenAIAPIKey))

	return &Agent{
		exaClient:  exa.NewClient(env.ExaAPIKey),
		provider:   provider,
		fastClient: fastClient,
	}
}

func MapDBMessages(dbMessages []db.Message) []Message {
	messages := make([]Message, 0, len(dbMessages))
	for _, it := range dbMessages {
		if it.Content == "" {
			continue
		}

		switch it.Type {
		case db.UserMessage:
			messages = append(messages, Message{Role: "user_message", Content: it.Content})
		case db.AssistantMessage:
			messages = append(messages, Message{Role: "assistant_message", Content: it.Content})
		case db.SystemMessage:
			messages = append(messages, Message{Role: "system_message", Content: it.Content})
		case db.AssistantToolCall:
			messages = append(messages, Message{Role: "assistant_tool_call", Content: it.Content, ToolCallID: it.ToolCallID, ToolName: it.ToolName})
		case db.ToolOutput:
			messages = append(messages, Message{Role: "tool_output", Content: it.Content, ToolCallID: it.ToolCallID, ToolName: it.ToolName})
		}
	}

	return messages
}

func (x *Agent) Stream(ctx context.Context, messages []db.Message) iter.Seq2[StreamEvent, error] {
	return func(yield func(StreamEvent, error) bool) {
		req := ProviderRequest{
			Messages:     MapDBMessages(messages),
			Tools:        []Tool{webSearchTool},
			Instructions: "You are a helpful assistant. Before each tool call, briefly tell the user what you are about to do in one short sentence.",
		}

		for range 10 {
			var calledTools []Toolcall

			for event, err := range x.provider.Stream(ctx, req) {
				if err != nil {
					yield(nil, err)
					return
				}

				switch ev := event.(type) {
				case AssistantDelta:
					if !yield(ev, nil) {
						return
					}
				case ReasoningDelta:
					if !yield(ev, nil) {
						return
					}
				case Toolcall:
					calledTools = append(calledTools, ev)
					if !yield(ev, nil) {
						return
					}
				case ProviderStateDelta:
					req.State = ev.State
				}
			}

			if len(calledTools) == 0 {
				return
			}

			for _, call := range calledTools {
				req.Messages = append(req.Messages, Message{
					Role:       "assistant_tool_call",
					Content:    call.Args,
					ToolCallID: call.CallID,
					ToolName:   call.Name,
				})

				result, err := x.runTool(ctx, call.Name, call.Args)
				if err != nil {
					result = fmt.Sprintf(`{"error":%q}`, err.Error())
				}

				if !yield(ToolOutput{CallID: call.CallID, Name: call.Name, Output: result}, nil) {
					return
				}

				req.Messages = append(req.Messages, Message{
					Role:       "tool_output",
					Content:    result,
					ToolCallID: call.CallID,
					ToolName:   call.Name,
				})
			}
		}

		yield(nil, fmt.Errorf("too many tool rounds"))
	}
}

func (x *Agent) runTool(ctx context.Context, name, rawArgs string) (string, error) {
	switch name {
	case "web_search":
		var args webSearchArgs
		if err := json.Unmarshal([]byte(rawArgs), &args); err != nil {
			return "", err
		}

		return x.webSearch(ctx, args)
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

func (x *Agent) GenerateTitleFromUserMessage(ctx context.Context, userMessage string) string {
	resp, err := x.fastClient.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openai.ChatModelGPT5_4Nano,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("Generate a concise title (max 6 words) for a chat thread based on the user's first message. Return only the title, no quotes or punctuation."),
			openai.UserMessage(userMessage),
		},
	})
	if err != nil {
		slog.Error("error generating title", "err", err)
		return "New Chat"
	}
	if len(resp.Choices) == 0 {
		return "New Chat"
	}

	title := strings.TrimSpace(resp.Choices[0].Message.Content)
	if title == "" {
		return "New Chat"
	}

	return title
}
