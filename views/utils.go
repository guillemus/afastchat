package views

import (
	"encoding/json"
	"log/slog"

	"github.com/starfederation/datastar-go/datastar"
)

func patchSignals(sse *datastar.ServerSentEventGenerator, v any) {
	bs, err := json.Marshal(v)
	if err != nil {
		slog.Error("failed to marshal", "value", v)
	}

	sse.PatchSignals(bs)
}
