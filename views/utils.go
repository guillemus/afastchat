package views

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/starfederation/datastar-go/datastar"

	. "maragu.dev/gomponents"
)

func patchSignals(sse *datastar.ServerSentEventGenerator, v any) {
	bs, err := json.Marshal(v)
	if err != nil {
		slog.Error("failed to marshal", "value", v)
	}

	sse.PatchSignals(bs)
}

func renderNode(w io.Writer, node Node) {
	if err := node.Render(w); err != nil {
		slog.Error("view render error", "err", err)
	}
}

func ssePatch(sse *datastar.ServerSentEventGenerator, node Node) {
	var sb strings.Builder
	if err := node.Render(&sb); err != nil {
		slog.Error("view render error", "err", err)
		return
	}

	if err := sse.PatchElements(sb.String()); err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		slog.Error("sse.PatchElements error", "err", err)
		return
	}
}

func sseExecJS(sse *datastar.ServerSentEventGenerator, script string) {
	// Add block scope so that declared variables are not on global scope.
	s := fmt.Sprintf("{ %s }", script)
	if err := sse.ExecuteScript(s); err != nil {
		slog.Error("sse.ExecuteScript", "err", err)
	}
}
