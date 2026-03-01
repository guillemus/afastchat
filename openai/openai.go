package openai

import (
	"t3chat-datastar/env"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

var Client = openai.NewClient(option.WithAPIKey(env.Env.OpenAIAPIKey))
