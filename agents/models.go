package agents

type ModelProvider string

const (
	ModelProviderOpenAI    ModelProvider = "openai"
	ModelProviderAnthropic ModelProvider = "anthropic"
	ModelProviderGoogle    ModelProvider = "google"
	ModelProviderMeta      ModelProvider = "meta"
	ModelProviderXAI       ModelProvider = "x-ai"
	ModelProviderDeepSeek  ModelProvider = "deepseek"
	ModelProviderQwen      ModelProvider = "qwen"
	ModelProviderMoonshot  ModelProvider = "moonshotai"
	ModelProviderZAI       ModelProvider = "z-ai"
	ModelProviderXiaomi    ModelProvider = "xiaomi"
	ModelProviderMiniMax   ModelProvider = "minimax"
	ModelProviderMistral   ModelProvider = "mistralai"
)

type Provider struct {
	ID    ModelProvider
	Label string
}

var Providers = []Provider{
	{ID: ModelProviderOpenAI, Label: "OpenAI"},
	{ID: ModelProviderAnthropic, Label: "Anthropic"},
	{ID: ModelProviderGoogle, Label: "Google"},
	{ID: ModelProviderMeta, Label: "Meta"},
	{ID: ModelProviderXAI, Label: "xAI"},
	{ID: ModelProviderDeepSeek, Label: "DeepSeek"},
	{ID: ModelProviderQwen, Label: "Qwen"},
	{ID: ModelProviderMoonshot, Label: "Moonshot AI"},
	{ID: ModelProviderZAI, Label: "Z.ai"},
	{ID: ModelProviderXiaomi, Label: "Xiaomi"},
	{ID: ModelProviderMiniMax, Label: "MiniMax"},
	{ID: ModelProviderMistral, Label: "Mistral"},
}

type Model struct {
	ID          string
	Name        string
	Provider    ModelProvider
	Description string
	CostTier    int
}

const DefaultModelID = "anthropic/claude-haiku-5.5"

var Models = []Model{
	{
		ID:          "openai/gpt-6-astra",
		Name:        "GPT-6 Astra",
		Provider:    ModelProviderOpenAI,
		Description: "OpenAI flagship for deep research and long-horizon work",
		CostTier:    4,
	},
	{
		ID:          "openai/gpt-6.1-sol",
		Name:        "GPT-6.1 Sol",
		Provider:    ModelProviderOpenAI,
		Description: "OpenAI model for complex coding and reasoning",
		CostTier:    2,
	},
	{
		ID:          "openai/gpt-6-luna",
		Name:        "GPT-6 Luna",
		Provider:    ModelProviderOpenAI,
		Description: "Fast OpenAI model for focused tasks",
		CostTier:    1,
	},
	{
		ID:          "anthropic/claude-fable-5.1",
		Name:        "Claude Fable 5.1",
		Provider:    ModelProviderAnthropic,
		Description: "Anthropic's most capable model for long-running work",
		CostTier:    4,
	},
	{
		ID:          "anthropic/claude-opus-5.5",
		Name:        "Claude Opus 5.5",
		Provider:    ModelProviderAnthropic,
		Description: "Anthropic flagship for demanding reasoning and coding",
		CostTier:    3,
	},
	{
		ID:          "anthropic/claude-sonnet-5.5",
		Name:        "Claude Sonnet 5.5",
		Provider:    ModelProviderAnthropic,
		Description: "Balanced Anthropic model for everyday work",
		CostTier:    2,
	},
	{
		ID:          "anthropic/claude-haiku-5.5",
		Name:        "Claude Haiku 5.5",
		Provider:    ModelProviderAnthropic,
		Description: "Small, fast Anthropic model for quick answers",
		CostTier:    1,
	},
	{
		ID:          "google/gemini-4-argon",
		Name:        "Gemini 4 Argon",
		Provider:    ModelProviderGoogle,
		Description: "Google flagship for deep reasoning, coding and agents",
		CostTier:    2,
	},
	{
		ID:          "google/gemini-3.8-flash",
		Name:        "Gemini 3.8 Flash",
		Provider:    ModelProviderGoogle,
		Description: "Fast Google model with strong all-round quality",
		CostTier:    2,
	},
	{
		ID:          "meta/muse-wonder",
		Name:        "Muse Wonder",
		Provider:    ModelProviderMeta,
		Description: "Meta flagship for demanding reasoning",
		CostTier:    3,
	},
	{
		ID:          "meta/muse-spark-1.3",
		Name:        "Muse Spark 1.3",
		Provider:    ModelProviderMeta,
		Description: "Fast Meta model for everyday work",
		CostTier:    2,
	},
	{
		ID:          "x-ai/grok-4.7",
		Name:        "Grok 4.7",
		Provider:    ModelProviderXAI,
		Description: "xAI flagship reasoning model",
		CostTier:    2,
	},
	{
		ID:          "deepseek/deepseek-v4-pro-0813",
		Name:        "DeepSeek V4 Pro",
		Provider:    ModelProviderDeepSeek,
		Description: "DeepSeek's largest open model for hard problems",
		CostTier:    2,
	},
	{
		ID:          "deepseek/deepseek-v4.1-flash",
		Name:        "DeepSeek V4.1 Flash",
		Provider:    ModelProviderDeepSeek,
		Description: "Fast, very cheap open model with strong reasoning",
		CostTier:    1,
	},
	{
		ID:          "qwen/qwen3.8-max-0902",
		Name:        "Qwen3.8 Max",
		Provider:    ModelProviderQwen,
		Description: "Alibaba flagship for complex reasoning",
		CostTier:    2,
	},
	{
		ID:          "qwen/qwen3.8-flash",
		Name:        "Qwen3.8 Flash",
		Provider:    ModelProviderQwen,
		Description: "Fast, low-cost Qwen model",
		CostTier:    1,
	},
	{
		ID:          "moonshotai/kimi-k3",
		Name:        "Kimi K3",
		Provider:    ModelProviderMoonshot,
		Description: "Moonshot's open frontier model for agentic work",
		CostTier:    2,
	},
	{
		ID:          "z-ai/glm-5.3",
		Name:        "GLM-5.3",
		Provider:    ModelProviderZAI,
		Description: "Z.ai open flagship for reasoning and coding",
		CostTier:    2,
	},
	{
		ID:          "z-ai/glm-5.3-flash",
		Name:        "GLM-5.3 Flash",
		Provider:    ModelProviderZAI,
		Description: "Fast, low-cost GLM model",
		CostTier:    1,
	},
	{
		ID:          "xiaomi/mimo-v2.6-pro",
		Name:        "MiMo-V2.6 Pro",
		Provider:    ModelProviderXiaomi,
		Description: "Top open-weights model at a very low price",
		CostTier:    1,
	},
	{
		ID:          "xiaomi/mimo-v2.6-flash",
		Name:        "MiMo-V2.6 Flash",
		Provider:    ModelProviderXiaomi,
		Description: "Fast, very cheap Xiaomi model",
		CostTier:    1,
	},
	{
		ID:          "minimax/minimax-m3",
		Name:        "MiniMax M3",
		Provider:    ModelProviderMiniMax,
		Description: "Efficient open model for coding and agents",
		CostTier:    1,
	},
	{
		ID:          "mistralai/mistral-large-4-0",
		Name:        "Mistral Large 4",
		Provider:    ModelProviderMistral,
		Description: "Mistral's flagship European model",
		CostTier:    2,
	},
}

func LookupModel(id string) (Model, bool) {
	for _, model := range Models {
		if model.ID == id {
			return model, true
		}
	}
	return Model{}, false
}

func ResolveModel(id string) Model {
	if model, ok := LookupModel(id); ok {
		return model
	}
	model, _ := LookupModel(DefaultModelID)
	return model
}
