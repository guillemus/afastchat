package agents

type UsageCost struct {
	InputCostMicros       int64
	CachedInputCostMicros int64
	OutputCostMicros      int64
	TotalCostMicros       int64
}

type modelPrice struct {
	InputMicrosPerMTok       int64
	CachedInputMicrosPerMTok int64
	OutputMicrosPerMTok      int64
}

var modelPrices = map[string]modelPrice{
	"fireworks:accounts/fireworks/models/kimi-k2p6": {
		InputMicrosPerMTok:       950_000,
		CachedInputMicrosPerMTok: 160_000,
		OutputMicrosPerMTok:      4_000_000,
	},
}

func CalculateUsageCost(usage Usage) UsageCost {
	price := modelPrices[usage.Provider+":"+usage.Model]
	billableInputTokens := usage.InputTokens - usage.CachedInputTokens

	inputCost := billableInputTokens * price.InputMicrosPerMTok / 1_000_000
	cachedInputCost := usage.CachedInputTokens * price.CachedInputMicrosPerMTok / 1_000_000
	outputCost := usage.OutputTokens * price.OutputMicrosPerMTok / 1_000_000

	return UsageCost{
		InputCostMicros:       inputCost,
		CachedInputCostMicros: cachedInputCost,
		OutputCostMicros:      outputCost,
		TotalCostMicros:       inputCost + cachedInputCost + outputCost,
	}
}
