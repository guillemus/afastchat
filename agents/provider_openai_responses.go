package agents

import (
	"context"
	"fmt"
	"iter"
	"log/slog"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"
)

type OpenAIResponsesProvider struct {
	client openai.Client
	model  shared.ResponsesModel
}

func NewOpenAIResponsesProvider(apiKey string, model shared.ResponsesModel) *OpenAIResponsesProvider {
	return &OpenAIResponsesProvider{
		client: openai.NewClient(option.WithAPIKey(apiKey)),
		model:  model,
	}
}

func (x *OpenAIResponsesProvider) Stream(ctx context.Context, req ProviderRequest) iter.Seq2[StreamEvent, error] {
	return func(yield func(StreamEvent, error) bool) {
		params := responses.ResponseNewParams{
			Model: x.model,
			Input: responses.ResponseNewParamsInputUnion{
				OfInputItemList: toResponseInput(req.Messages),
			},
			Instructions:      openai.String(req.Instructions),
			Tools:             toResponseTools(req.Tools),
			ParallelToolCalls: param.NewOpt(true),
			Reasoning: shared.ReasoningParam{
				Effort:  responses.ReasoningEffortXhigh,
				Summary: responses.ReasoningSummaryDetailed,
			},
		}
		if req.State.OpenAIResponseID != "" {
			params.PreviousResponseID = param.NewOpt(req.State.OpenAIResponseID)
		}

		stream := x.client.Responses.NewStreaming(ctx, params)
		defer func() {
			if err := stream.Close(); err != nil {
				slog.Error("failed to close response stream", "err", err)
			}
		}()

		calledTools := newCalledTools()

		for stream.Next() {
			event := stream.Current()
			logResponseEvent(event)

			switch ev := event.AsAny().(type) {
			case responses.ResponseCreatedEvent:
				if !yield(ProviderStateDelta{State: ProviderState{OpenAIResponseID: ev.Response.ID}}, nil) {
					return
				}
			case responses.ResponseTextDeltaEvent:
				if ev.Delta == "" {
					continue
				}
				if !yield(AssistantDelta{ev.Delta}, nil) {
					return
				}
			case responses.ResponseOutputItemAddedEvent:
				if ev.Item.Type != "function_call" {
					continue
				}
				item := ev.Item.AsFunctionCall()
				call := calledTools.functionCall(item.ID)
				call.ItemID = item.ID
				call.CallID = item.CallID
				call.Name = item.Name
				if call.Args.Len() == 0 && item.Arguments != "" {
					call.Args.WriteString(item.Arguments)
				}
			case responses.ResponseFunctionCallArgumentsDeltaEvent:
				call := calledTools.functionCall(ev.ItemID)
				call.ItemID = ev.ItemID
				call.Args.WriteString(ev.Delta)
			case responses.ResponseFunctionCallArgumentsDoneEvent:
				call := calledTools.functionCall(ev.ItemID)
				call.ItemID = ev.ItemID
				if call.Name == "" {
					call.Name = ev.Name
				}
				if call.Args.Len() == 0 {
					call.Args.WriteString(ev.Arguments)
				}
			case responses.ResponseOutputItemDoneEvent:
				if ev.Item.Type != "function_call" {
					continue
				}
				item := ev.Item.AsFunctionCall()
				call := calledTools.functionCall(item.ID)
				call.ItemID = item.ID
				call.CallID = item.CallID
				call.Name = item.Name
				if call.Args.Len() == 0 && item.Arguments != "" {
					call.Args.WriteString(item.Arguments)
				}
				if !yield(Toolcall{CallID: call.CallID, Name: call.Name, Args: call.Args.String()}, nil) {
					return
				}
			case responses.ResponseCompletedEvent:
				return
			case responses.ResponseErrorEvent:
				yield(nil, fmt.Errorf("response error: %s", ev.Message))
				return
			case responses.ResponseFailedEvent:
				yield(nil, fmt.Errorf("response failed"))
				return
			case responses.ResponseReasoningSummaryTextDeltaEvent:
				if ev.Delta == "" {
					continue
				}
				if !yield(ReasoningDelta{ev.Delta}, nil) {
					return
				}
			case responses.ResponseReasoningTextDeltaEvent:
				if ev.Delta == "" {
					continue
				}
				if !yield(ReasoningDelta{ev.Delta}, nil) {
					return
				}
			case responses.ResponseRefusalDeltaEvent:
			case responses.ResponseRefusalDoneEvent:
			case responses.ResponseTextDoneEvent:
			}
		}

		if err := stream.Err(); err != nil {
			yield(nil, fmt.Errorf("response stream failed: %w", err))
			return
		}
	}
}

func toResponseInput(messages []Message) responses.ResponseInputParam {
	input := make(responses.ResponseInputParam, 0, len(messages))
	for _, it := range messages {
		if it.Content == "" {
			continue
		}

		switch it.Role {
		case "user_message":
			input = append(input, responses.ResponseInputItemParamOfMessage(
				it.Content,
				responses.EasyInputMessageRoleUser,
			))
		case "assistant_message":
			input = append(input, responses.ResponseInputItemParamOfMessage(
				it.Content,
				responses.EasyInputMessageRoleAssistant,
			))
		case "system_message":
			input = append(input, responses.ResponseInputItemParamOfMessage(
				it.Content,
				responses.EasyInputMessageRoleSystem,
			))
		case "assistant_tool_call":
			input = append(input, responses.ResponseInputItemParamOfFunctionCall(
				it.Content,
				it.ToolCallID,
				it.ToolName,
			))
		case "tool_output":
			input = append(input, responses.ResponseInputItemParamOfFunctionCallOutput(
				it.ToolCallID,
				it.Content,
			))
		}
	}

	return input
}

func toResponseTools(tools []Tool) []responses.ToolUnionParam {
	out := make([]responses.ToolUnionParam, 0, len(tools))
	for _, tool := range tools {
		out = append(out, responses.ToolParamOfFunction(tool.Name, tool.Parameters, tool.Strict))
	}
	return out
}

func logResponseEvent(event responses.ResponseStreamEventUnion) {
	attrs := []any{
		"type", event.Type,
	}

	if event.Delta != "" {
		attrs = append(attrs, "delta", event.Delta)
	}
	if event.Arguments != "" {
		attrs = append(attrs, "arguments", event.Arguments)
	}
	if event.Name != "" {
		attrs = append(attrs, "name", event.Name)
	}
	if event.Text != "" {
		attrs = append(attrs, "text", event.Text)
	}
	if event.Refusal != "" {
		attrs = append(attrs, "refusal", event.Refusal)
	}

	slog.Debug("openai response event", attrs...)
}
