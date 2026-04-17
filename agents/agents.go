package agents

import (
	"context"
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
	model  shared.ResponsesModel
	client openai.Client

	exaClient exa.Client
}

var errStopStreaming = errors.New("stop streaming")

var openaiClient = openai.NewClient(option.WithAPIKey(env.Env.OpenAIAPIKey))

func NewAgent() *Agent {

	// var openrouterClient = openai.NewClient(
	// 	option.WithAPIKey(env.Env.OpenRouterAPIKey),
	// 	option.WithBaseURL("https://openrouter.ai/api/v1"),
	// )

	return &Agent{
		client: openaiClient,
		// model:  openai.ChatModelGPT5_4Mini,
		model: openai.ChatModelGPT5_4Nano,

		// client: openrouterClient,
		// model:  "minimax/minimax-m2.7",
	}
}

func MapToResponseInput(dbMessages []db.Message) responses.ResponseInputParam {
	input := make(responses.ResponseInputParam, 0, len(dbMessages))
	for _, it := range dbMessages {
		if it.Content == "" {
			continue
		}
		switch it.Role {
		case "user":
			input = append(input, responses.ResponseInputItemParamOfMessage(
				it.Content,
				responses.EasyInputMessageRoleUser,
			))
		case "assistant":
			input = append(input, responses.ResponseInputItemParamOfMessage(
				it.Content,
				responses.EasyInputMessageRoleAssistant,
			))
		case "system":
			input = append(input, responses.ResponseInputItemParamOfMessage(
				it.Content,
				responses.EasyInputMessageRoleSystem,
			))
		}
	}

	return input
}

func (x *Agent) Stream(ctx context.Context, messages []db.Message) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		input := MapToResponseInput(messages)

		for range 3 {
			state, err := x.streamResponseRound(ctx, input, yield)
			if err != nil {
				if errors.Is(err, errStopStreaming) {
					return
				}
				yield("", err)
				return
			}

			if len(state.Order) == 0 {
				return
			}

			for _, itemID := range state.Order {
				call := state.FunctionCalls[itemID]
				input = append(input, responses.ResponseInputItemParamOfFunctionCall(
					call.Args.String(),
					call.ID,
					call.Name,
				))

				result, err := x.runTool(ctx, call.Name, call.Args.String())
				if err != nil {
					result = fmt.Sprintf(`{"error":%q}`, err.Error())
				}

				input = append(input, responses.ResponseInputItemParamOfFunctionCallOutput(
					call.ID,
					result,
				))
			}
		}

		yield("", fmt.Errorf("too many tool rounds"))
	}
}

type functionCallBuf struct {
	ID   string
	Name string
	Args strings.Builder
}

type responseRoundState struct {
	FunctionCalls map[string]*functionCallBuf
	Order         []string
}

func (x *Agent) streamResponseRound(
	ctx context.Context,
	input responses.ResponseInputParam,
	yield func(string, error) bool,
) (*responseRoundState, error) {
	stream := x.client.Responses.NewStreaming(ctx, responses.ResponseNewParams{
		Model: x.model,
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: input,
		},
		Tools: []responses.ToolUnionParam{
			websearchToolDef,
		},
		ParallelToolCalls: param.NewOpt(true),
		Reasoning: shared.ReasoningParam{
			Effort:  responses.ReasoningEffortMedium,
			Summary: responses.ReasoningSummaryDetailed,
		},
	})
	defer func() {
		if err := stream.Close(); err != nil {
			slog.Error("failed to close response stream", "err", err)
		}
	}()

	state := &responseRoundState{
		FunctionCalls: map[string]*functionCallBuf{},
	}

	for stream.Next() {
		event := stream.Current()
		logResponseEvent(event)

		switch ev := event.AsAny().(type) {
		case responses.ResponseTextDeltaEvent:
			if ev.Delta == "" {
				continue
			}
			if !yield(ev.Delta, nil) {
				return nil, errStopStreaming
			}

		case responses.ResponseFunctionCallArgumentsDeltaEvent:
			call := state.functionCall(ev.ItemID)
			call.Args.WriteString(ev.Delta)

		case responses.ResponseFunctionCallArgumentsDoneEvent:
			call := state.functionCall(ev.ItemID)
			call.Name = ev.Name
			if call.Args.Len() == 0 {
				call.Args.WriteString(ev.Arguments)
			}

		case responses.ResponseCompletedEvent:
			return state, nil

		case responses.ResponseErrorEvent:
			return nil, fmt.Errorf("response error: %s", ev.Message)

		case responses.ResponseFailedEvent:
			return nil, fmt.Errorf("response failed")
		}
	}

	if err := stream.Err(); err != nil {
		return nil, fmt.Errorf("response stream failed: %w", err)
	}

	return state, nil
}

func (x *Agent) runTool(ctx context.Context, name, rawArgs string) (string, error) {
	switch name {
	case "web_search":
		return x.webSearch(ctx, rawArgs)
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

func (x *responseRoundState) functionCall(itemID string) *functionCallBuf {
	call := x.FunctionCalls[itemID]
	if call != nil {
		return call
	}

	call = &functionCallBuf{ID: itemID}
	x.FunctionCalls[itemID] = call
	x.Order = append(x.Order, itemID)

	return call
}

func GenerateTitleFromUserMessage(ctx context.Context, userMessage string) string {
	resp, err := openaiClient.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
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

	if event.ItemID != "" {
		attrs = append(attrs, "item_id", event.ItemID)
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
