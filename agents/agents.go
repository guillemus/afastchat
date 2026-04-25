package agents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"strings"

	"afastchat.com/db"
	"afastchat.com/env"
	"afastchat.com/exa"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"
)

type Agent struct {
	model        shared.ResponsesModel
	openaiClient openai.Client
	exaClient    *exa.Client
}

var errStopStreaming = errors.New("stop streaming")

func NewAgent(env env.Env) *Agent {
	return &Agent{
		exaClient:    exa.NewClient(env.ExaAPIKey),
		openaiClient: openai.NewClient(option.WithAPIKey(env.OpenAIAPIKey)),
		model:        openai.ChatModelGPT5_4Nano,
	}
}

func MapToResponseInput(dbMessages []db.Message) responses.ResponseInputParam {
	input := make(responses.ResponseInputParam, 0, len(dbMessages))
	for _, it := range dbMessages {
		if it.Content == "" {
			continue
		}

		switch it.Type {
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

type (
	StreamEvent interface {
		isStreamEvent()
	}
	AssistantDelta struct {
		Delta string
	}
	ReasoningDelta struct {
		Delta string
	}
	Toolcall struct {
		CallID string
		Name   string
		Args   string
	}
	ToolOutput struct {
		CallID string
		Name   string
		Output string
	}
)

func (AssistantDelta) isStreamEvent() {}
func (ReasoningDelta) isStreamEvent() {}
func (Toolcall) isStreamEvent()       {}
func (ToolOutput) isStreamEvent()     {}

func (x *Agent) Stream(ctx context.Context, messages []db.Message) iter.Seq2[StreamEvent, error] {
	return func(yield func(StreamEvent, error) bool) {
		input := MapToResponseInput(messages)

		var currResponseID string

		for range 10 {
			calledTools, responseID, err := x.streamResponse(ctx, input, currResponseID, yield)
			if err != nil {
				if errors.Is(err, errStopStreaming) {
					return
				}
				yield(nil, err)
				return
			}

			currResponseID = responseID

			if len(calledTools.Order) == 0 {
				return
			}

			for _, itemID := range calledTools.Order {
				call := calledTools.Tools[itemID]
				input = append(input, responses.ResponseInputItemParamOfFunctionCall(
					call.Args.String(),
					call.CallID,
					call.Name,
				))

				result, err := x.runTool(ctx, call.Name, call.Args.String())
				if err != nil {
					result = fmt.Sprintf(`{"error":%q}`, err.Error())
				}

				if !yield(ToolOutput{CallID: call.CallID, Name: call.Name, Output: result}, nil) {
					return
				}

				input = append(input, responses.ResponseInputItemParamOfFunctionCallOutput(
					call.CallID,
					result,
				))
			}
		}

		yield(nil, fmt.Errorf("too many tool rounds"))
	}
}

type ToolcallBuf struct {
	ItemID string
	CallID string
	Name   string
	Args   strings.Builder
}

type CalledTools struct {
	Tools map[string]*ToolcallBuf
	Order []string
}

func (x *Agent) streamResponse(
	ctx context.Context,
	input responses.ResponseInputParam,
	previousResponseID string,
	yield func(StreamEvent, error) bool,
) (*CalledTools, string, error) {
	params := responses.ResponseNewParams{
		Model: x.model,
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: input,
		},
		Instructions: openai.String("You are a helpful assistant. Before each tool call, briefly tell the user what you are about to do in one short sentence."),
		Tools: []responses.ToolUnionParam{
			websearchToolDef,
		},
		ParallelToolCalls: param.NewOpt(true),
		Reasoning: shared.ReasoningParam{
			Effort:  responses.ReasoningEffortXhigh,
			Summary: responses.ReasoningSummaryDetailed,
		},
	}
	if previousResponseID != "" {
		params.PreviousResponseID = param.NewOpt(previousResponseID)
	}

	stream := x.openaiClient.Responses.NewStreaming(ctx, params)
	defer func() {
		if err := stream.Close(); err != nil {
			slog.Error("failed to close response stream", "err", err)
		}
	}()

	calledTools := &CalledTools{
		Tools: map[string]*ToolcallBuf{},
	}
	var responseID string

	for stream.Next() {
		event := stream.Current()
		logResponseEvent(event)

		switch ev := event.AsAny().(type) {
		case responses.ResponseCreatedEvent:
			responseID = ev.Response.ID
		case responses.ResponseTextDeltaEvent:
			if ev.Delta == "" {
				continue
			}

			delta := AssistantDelta{ev.Delta}
			if !yield(delta, nil) {
				return nil, responseID, errStopStreaming
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

			toolcall := Toolcall{CallID: call.CallID, Name: call.Name, Args: call.Args.String()}
			if !yield(toolcall, nil) {
				return nil, responseID, errStopStreaming
			}

		case responses.ResponseCompletedEvent:
			return calledTools, responseID, nil

		case responses.ResponseErrorEvent:
			return nil, responseID, fmt.Errorf("response error: %s", ev.Message)

		case responses.ResponseFailedEvent:
			return nil, responseID, fmt.Errorf("response failed")

		case responses.ResponseReasoningTextDeltaEvent:
			if ev.Delta == "" {
				continue
			}

			delta := ReasoningDelta{ev.Delta}
			if !yield(delta, nil) {
				return nil, responseID, errStopStreaming
			}

		case responses.ResponseRefusalDeltaEvent:
		case responses.ResponseRefusalDoneEvent:
		case responses.ResponseTextDoneEvent:
		}
	}

	if err := stream.Err(); err != nil {
		return nil, responseID, fmt.Errorf("response stream failed: %w", err)
	}

	return calledTools, responseID, nil
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

func (x *CalledTools) functionCall(itemID string) *ToolcallBuf {
	call := x.Tools[itemID]
	if call != nil {
		return call
	}

	call = &ToolcallBuf{ItemID: itemID}
	x.Tools[itemID] = call
	x.Order = append(x.Order, itemID)

	return call
}

func (x *Agent) GenerateTitleFromUserMessage(ctx context.Context, userMessage string) string {
	resp, err := x.openaiClient.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
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
