package agents

import (
	"context"
	"datastar-chat/db"
	"datastar-chat/env"
	"iter"
	"log/slog"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type Agent struct {
	model  openai.ChatModel
	client openai.Client
}

func NewAgent() *Agent {
	return &Agent{
		client: openai.NewClient(option.WithAPIKey(env.Env.OpenAIAPIKey)),
		model:  openai.ChatModelGPT5_4Mini,
	}
}

func MapToOpenaiMessages(dbMessages []db.Message) []openai.ChatCompletionMessageParamUnion {
	oaiMessages := make([]openai.ChatCompletionMessageParamUnion, 0, len(dbMessages))
	for _, m := range dbMessages {
		if m.Content == "" {
			continue
		}
		switch m.Role {
		case "user":
			oaiMessages = append(oaiMessages, openai.UserMessage(m.Content))
		case "assistant":
			oaiMessages = append(oaiMessages, openai.AssistantMessage(m.Content))
		case "system":
			oaiMessages = append(oaiMessages, openai.SystemMessage(m.Content))
		}
	}

	return oaiMessages
}

func (a *Agent) Stream(ctx context.Context, messages []db.Message) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		openaiMsgs := MapToOpenaiMessages(messages)

		stream := a.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
			Model:    a.model,
			Messages: openaiMsgs,
		})
		defer func() {
			if err := stream.Close(); err != nil {
				slog.Error("failed to close stream", "error", err)
			}
		}()

		for stream.Next() {
			chunk := stream.Current()
			if len(chunk.Choices) == 0 {
				continue
			}
			token := chunk.Choices[0].Delta.Content
			if token == "" {
				continue
			}

			if !yield(token, nil) {
				return
			}
		}

		if err := stream.Err(); err != nil {
			yield("", err)
		}
	}
}

func (a *Agent) GenerateTitleFromUserMessage(ctx context.Context, userMessage string) string {
	resp, err := a.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openai.ChatModelGPT5Nano,
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
