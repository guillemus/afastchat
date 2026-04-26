package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"strconv"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/shared"
)

type FireworksChatProvider struct {
	client openai.Client
	model  string
}

func NewFireworksChatProvider(apiKey, model string) *FireworksChatProvider {
	return &FireworksChatProvider{
		client: openai.NewClient(
			option.WithAPIKey(apiKey),
			option.WithBaseURL("https://api.fireworks.ai/inference/v1"),
		),
		model: model,
	}
}

func (x *FireworksChatProvider) Stream(ctx context.Context, req ProviderRequest) iter.Seq2[StreamEvent, error] {
	return func(yield func(StreamEvent, error) bool) {
		params := openai.ChatCompletionNewParams{
			Model:             shared.ChatModel(x.model),
			Messages:          toChatMessages(req.Messages, req.Instructions),
			Tools:             toChatTools(req.Tools),
			ParallelToolCalls: param.NewOpt(true),
			MaxTokens:         param.NewOpt(int64(32768)),
			Temperature:       param.NewOpt(0.6),
			TopP:              param.NewOpt(1.0),
			PresencePenalty:   param.NewOpt(0.0),
			FrequencyPenalty:  param.NewOpt(0.0),
		}

		stream := x.client.Chat.Completions.NewStreaming(
			ctx,
			params,
			option.WithJSONSet("top_k", 40),
			option.WithJSONSet("reasoning_effort", "medium"),
		)
		defer stream.Close()

		calledTools := newCalledTools()

		for stream.Next() {
			chunk := stream.Current()
			for _, choice := range chunk.Choices {
				if reasoningContent := fireworksReasoningContent(choice.Delta.RawJSON()); reasoningContent != "" {
					if !yield(ReasoningDelta{reasoningContent}, nil) {
						return
					}
				}

				if choice.Delta.Content != "" {
					if !yield(AssistantDelta{choice.Delta.Content}, nil) {
						return
					}
				}

				for _, delta := range choice.Delta.ToolCalls {
					itemID := strconv.FormatInt(delta.Index, 10)
					call := calledTools.functionCall(itemID)
					call.ItemID = itemID
					if delta.ID != "" {
						call.CallID = delta.ID
					}
					if delta.Function.Name != "" {
						call.Name = delta.Function.Name
					}
					if delta.Function.Arguments != "" {
						call.Args.WriteString(delta.Function.Arguments)
					}
				}
			}
		}

		if err := stream.Err(); err != nil {
			yield(nil, fmt.Errorf("chat completion stream failed: %w", err))
			return
		}

		for _, itemID := range calledTools.Order {
			call := calledTools.Tools[itemID]
			if !yield(Toolcall{CallID: call.CallID, Name: call.Name, Args: call.Args.String()}, nil) {
				return
			}
		}
	}
}

func fireworksReasoningContent(raw string) string {
	var delta struct {
		ReasoningContent string `json:"reasoning_content"`
	}
	if err := json.Unmarshal([]byte(raw), &delta); err != nil {
		return ""
	}
	return delta.ReasoningContent
}

func toChatMessages(messages []Message, instructions string) []openai.ChatCompletionMessageParamUnion {
	out := make([]openai.ChatCompletionMessageParamUnion, 0, len(messages)+1)
	if instructions != "" {
		out = append(out, openai.SystemMessage(instructions))
	}

	for _, it := range messages {
		if it.Content == "" {
			continue
		}

		switch it.Role {
		case "user_message":
			out = append(out, openai.UserMessage(it.Content))
		case "assistant_message":
			out = append(out, openai.AssistantMessage(it.Content))
		case "system_message":
			out = append(out, openai.SystemMessage(it.Content))
		case "assistant_tool_call":
			assistant := openai.ChatCompletionAssistantMessageParam{
				ToolCalls: []openai.ChatCompletionMessageToolCallUnionParam{
					{
						OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
							ID: it.ToolCallID,
							Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
								Name:      it.ToolName,
								Arguments: it.Content,
							},
						},
					},
				},
			}
			out = append(out, openai.ChatCompletionMessageParamUnion{OfAssistant: &assistant})
		case "tool_output":
			out = append(out, openai.ToolMessage(it.Content, it.ToolCallID))
		}
	}

	return out
}

func toChatTools(tools []Tool) []openai.ChatCompletionToolUnionParam {
	out := make([]openai.ChatCompletionToolUnionParam, 0, len(tools))
	for _, tool := range tools {
		out = append(out, openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
			Name:        tool.Name,
			Description: param.NewOpt(tool.Description),
			Parameters:  shared.FunctionParameters(tool.Parameters),
			Strict:      param.NewOpt(tool.Strict),
		}))
	}
	return out
}
