package openai

import (
	"datastar-chat/env"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

var Client = openai.NewClient(option.WithAPIKey(env.Env.OpenAIAPIKey))
