package server

import (
	"bytes"
	"context"
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"datastar-chat/db"
	openaiPkg "datastar-chat/openai"

	"github.com/google/uuid"
	"github.com/openai/openai-go/v3"
	"github.com/starfederation/datastar-go/datastar"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

var md = goldmark.New(goldmark.WithExtensions(extension.GFM))

func renderMarkdown(src string) string {
	var buf bytes.Buffer
	if err := md.Convert([]byte(src), &buf); err != nil {
		return html.EscapeString(src)
	}
	return buf.String()
}

func (s *Server) appRoutes() {
	s.mux.HandleFunc("GET /", s.handleIndex)
	s.mux.HandleFunc("GET /threads/{id}", s.handleThread)
	s.mux.HandleFunc("GET /threads/{id}/stream", s.handleResumeStream)
	s.mux.HandleFunc("GET /threads/search", s.handleSearchThreads)
	s.mux.HandleFunc("POST /chat/send", s.handleSendMessage)
	s.mux.HandleFunc("POST /threads/{id}/archive", s.handleArchiveThread)
}

type indexData struct {
	Threads []db.Thread
}

type threadData struct {
	Threads            []db.Thread
	Thread             db.Thread
	Messages           []db.Message
	Streaming          bool
	StreamingMessageID uuid.UUID
}

type threadItemData struct {
	Thread         db.Thread
	ActiveThreadID uuid.UUID
}

func (s *Server) renderThreadItem(thread db.Thread, activeThreadID uuid.UUID) (string, error) {
	threadItemHTML, err := s.renderer.RenderToString("partials/thread_item", threadItemData{
		Thread:         thread,
		ActiveThreadID: activeThreadID,
	})
	if err != nil {
		return "", fmt.Errorf("render thread item: %w", err)
	}

	return threadItemHTML, nil
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	threads, err := s.db.ListThreads(r.Context())
	if err != nil {
		slog.Error("list threads", "err", err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	if err := s.renderer.Render(w, "index", indexData{Threads: threads}); err != nil {
		slog.Error("render index", "err", err)
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}

func (s *Server) handleThread(w http.ResponseWriter, r *http.Request) {
	threadID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid thread id", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	thread, err := s.db.GetThread(ctx, threadID)
	if err != nil {
		slog.Error("get thread", "err", err)
		http.Error(w, "thread not found", http.StatusNotFound)
		return
	}

	messages, err := s.db.GetMessagesByThread(ctx, threadID)
	if err != nil {
		slog.Error("get messages", "err", err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	for i := range messages {
		if messages[i].Role == "assistant" {
			messages[i].Content = renderMarkdown(messages[i].Content)
		}
	}

	threads, err := s.db.ListThreads(ctx)
	if err != nil {
		slog.Error("list threads", "err", err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	stream, streaming := s.streams.Get(threadID)
	slog.Info("handleThread", "threadId", threadID, "streaming", streaming)

	data := threadData{
		Threads:   threads,
		Thread:    thread,
		Messages:  messages,
		Streaming: streaming,
	}
	if streaming {
		data.StreamingMessageID = stream.messageID
	}

	if err := s.renderer.Render(w, "thread", data); err != nil {
		slog.Error("render thread", "err", err)
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}

func (s *Server) handleArchiveThread(w http.ResponseWriter, r *http.Request) {
	threadID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid thread id", http.StatusBadRequest)
		return
	}

	if err := s.db.ArchiveThread(r.Context(), threadID); err != nil {
		slog.Error("archive thread", "err", err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	sse := datastar.NewSSE(w, r)

	sse.RemoveElementByID(fmt.Sprintf("thread-%s", threadID))
	sse.ExecuteScript(fmt.Sprintf(
		`if (window.location.pathname === '/threads/%s') { window.location.replace('/') }`,
		threadID,
	))
}

type searchThreadsSignals struct {
	ThreadSearch string `json:"threadSearch"`
	ThreadID     string `json:"threadId"`
}

func (s *Server) handleSearchThreads(w http.ResponseWriter, r *http.Request) {
	var sig searchThreadsSignals
	if err := datastar.ReadSignals(r, &sig); err != nil {
		slog.Warn("readSignals failed, using query fallback", "err", err)
		sig.ThreadSearch = r.URL.Query().Get("threadSearch")
		sig.ThreadID = r.URL.Query().Get("threadId")
	}

	threads, err := s.db.ListThreads(r.Context())
	if err != nil {
		slog.Error("list threads", "err", err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	query := strings.ToLower(strings.TrimSpace(sig.ThreadSearch))
	filtered := make([]db.Thread, 0, len(threads))
	for _, t := range threads {
		if query == "" || strings.Contains(strings.ToLower(t.Title), query) {
			filtered = append(filtered, t)
		}
	}

	var activeThreadID uuid.UUID
	if sig.ThreadID != "" {
		parsed, err := uuid.Parse(sig.ThreadID)
		if err != nil {
			slog.Warn("invalid threadId signal", "threadId", sig.ThreadID, "err", err)
		} else {
			activeThreadID = parsed
		}
	}

	var threadItemsHTML strings.Builder
	for _, t := range filtered {
		threadItemHTML, err := s.renderThreadItem(t, activeThreadID)
		if err != nil {
			slog.Error("render thread item", "threadId", t.ID, "err", err)
			http.Error(w, "template error", http.StatusInternalServerError)
			return
		}
		threadItemsHTML.WriteString(threadItemHTML)
	}

	sse := datastar.NewSSE(w, r)
	sse.PatchElements(
		threadItemsHTML.String(),
		datastar.WithSelectorID("thread-items"),
		datastar.WithModeInner(),
	)
}

type sendSignals struct {
	Message  string `json:"message"`
	ThreadID string `json:"threadId"`
}

func (s *Server) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	var sig sendSignals
	if err := datastar.ReadSignals(r, &sig); err != nil {
		slog.Warn("readSignals failed", "err", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	msg := strings.TrimSpace(sig.Message)
	if msg == "" {
		slog.Warn("empty message received")
		http.Error(w, "empty message", http.StatusBadRequest)
		return
	}

	slog.Info("message received", "len", len(msg), "threadId", sig.ThreadID)

	ctx := r.Context()

	// Resolve or create thread
	isNewThread := sig.ThreadID == ""
	var threadID uuid.UUID
	if !isNewThread {
		parsed, err := uuid.Parse(sig.ThreadID)
		if err != nil {
			slog.Warn("invalid thread id", "threadId", sig.ThreadID, "err", err)
			http.Error(w, "invalid thread id", http.StatusBadRequest)
			return
		}
		threadID = parsed
	} else {
		thread, err := s.db.CreateThread(ctx, "New Chat")
		if err != nil {
			slog.Error("create thread", "err", err)
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		threadID = thread.ID
		slog.Info("created thread", "threadId", threadID)
	}

	// Save user message
	userMsg, err := s.db.CreateMessage(ctx, db.CreateMessageParams{
		ThreadID: threadID,
		Role:     "user",
		Content:  msg,
	})
	if err != nil {
		slog.Error("create user message", "err", err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	// Create placeholder assistant message
	assistantMsg, err := s.db.CreateMessage(ctx, db.CreateMessageParams{
		ThreadID: threadID,
		Role:     "assistant",
		Content:  "",
	})
	if err != nil {
		slog.Error("create assistant message", "err", err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	slog.Info("db messages saved", "userMsgId", userMsg.ID, "assistantMsgId", assistantMsg.ID)

	// Open SSE connection (MUST be after ReadSignals)
	sse := datastar.NewSSE(w, r, datastar.WithCompression())

	// Patch signals: set threadId, clear message, set loading
	sse.MarshalAndPatchSignals(map[string]any{
		"threadId": threadID.String(),
		"message":  "",
		"loading":  true,
	})

	// If new thread, prepend to sidebar
	if isNewThread {
		threadItemHTML, err := s.renderThreadItem(db.Thread{ID: threadID, Title: "New Chat"}, threadID)
		if err != nil {
			slog.Error("render thread item", "threadId", threadID, "err", err)
		} else {
			sse.PatchElements(
				threadItemHTML,
				datastar.WithSelectorID("thread-items"),
				datastar.WithModePrepend(),
			)

			sse.ExecuteScript(fmt.Sprintf(
				`history.pushState({}, '', '/threads/%s')`,
				threadID,
			))
		}
	}

	userDivID := fmt.Sprintf("msg-%s", userMsg.ID)
	assistantDivID := fmt.Sprintf("msg-%s", assistantMsg.ID)

	// Append user + assistant message bubbles before scroll anchor
	sse.PatchElements(
		fmt.Sprintf(
			`<div id="%s" class="message user-message is-new">%s</div>
			 <div id="%s" class="message assistant-message is-new"></div>`,
			userDivID, html.EscapeString(msg), assistantDivID,
		),
		datastar.WithSelectorID("scroll-anchor"),
		datastar.WithModeBefore(),
	)

	// Build messages for OpenAI from DB history
	dbMessages, err := s.db.GetMessagesByThread(ctx, threadID)
	if err != nil {
		slog.Error("get messages", "threadId", threadID, "err", err)
		return
	}
	slog.Info("loaded history", "threadId", threadID, "msgCount", len(dbMessages))

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

	// Create active stream and launch background goroutine
	stream := newActiveStream(threadID, assistantMsg.ID)
	s.streams.Set(threadID, stream)

	go s.runOpenAIStream(threadID, assistantMsg.ID, oaiMessages, isNewThread, msg)

	// Relay stream to this SSE connection
	s.relayStream(r.Context(), sse, stream, assistantDivID)

	// Signal loading done
	sse.MarshalAndPatchSignals(map[string]any{
		"loading": false,
	})
}

func (s *Server) runOpenAIStream(
	threadID uuid.UUID, assistantMsgID uuid.UUID,
	oaiMessages []openai.ChatCompletionMessageParamUnion,
	isNewThread bool, userMessage string,
) {
	stream, ok := s.streams.Get(threadID)
	if !ok {
		return
	}

	ctx := context.Background()

	slog.Info("starting openai stream", "threadId", threadID, "model", openai.ChatModelGPT5_2, "oaiMsgCount", len(oaiMessages))
	oaiStream := openaiPkg.Client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Model:    openai.ChatModelGPT5_2,
		Messages: oaiMessages,
	})

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	lastSaved := ""

	for oaiStream.Next() {
		chunk := oaiStream.Current()
		if len(chunk.Choices) == 0 {
			continue
		}
		token := chunk.Choices[0].Delta.Content
		if token == "" {
			continue
		}
		stream.Append(token)

		// Throttled DB save
		select {
		case <-ticker.C:
			content, _ := stream.Snapshot()
			if content != lastSaved {
				if err := s.db.UpdateMessageContent(ctx, db.UpdateMessageContentParams{
					ID:      assistantMsgID,
					Content: content,
				}); err != nil {
					slog.Error("throttled save", "err", err)
				}
				lastSaved = content
			}
		default:
		}
	}

	if err := oaiStream.Err(); err != nil {
		slog.Error("openai stream", "threadId", threadID, "err", err)
	}
	oaiStream.Close()

	// Final DB save
	finalContent, _ := stream.Snapshot()
	slog.Info("stream complete", "threadId", threadID, "contentLen", len(finalContent))
	if err := s.db.UpdateMessageContent(ctx, db.UpdateMessageContentParams{
		ID:      assistantMsgID,
		Content: finalContent,
	}); err != nil {
		slog.Error("update assistant message", "err", err)
	}

	stream.Finish()

	// Keep stream in registry briefly so page refresh can resume.
	// Clean up after 30s if no one picks it up.
	go func() {
		time.Sleep(30 * time.Second)
		s.streams.Delete(threadID)
	}()

	// Auto-generate title for new threads
	if isNewThread {
		title := s.generateThreadTitle(ctx, userMessage)
		if err := s.db.UpdateThreadTitle(ctx, db.UpdateThreadTitleParams{
			ID:    threadID,
			Title: title,
		}); err != nil {
			slog.Error("update thread title", "err", err)
		}
	}
}

func (s *Server) relayStream(ctx context.Context, sse *datastar.ServerSentEventGenerator, stream *ActiveStream, assistantDivID string) {
	ch, content, done := stream.Subscribe()
	defer stream.Unsubscribe(ch)

	if content != "" {
		sse.PatchElements(
			fmt.Sprintf(`<div id="%s" class="message assistant-message">%s</div>`,
				assistantDivID, renderMarkdown(content)),
			datastar.WithSelectorID(assistantDivID),
			datastar.WithModeOuter(),
		)
	}
	if done {
		return
	}

	for {
		select {
		case <-ch:
			content, done = stream.Snapshot()
			sse.PatchElements(
				fmt.Sprintf(`<div id="%s" class="message assistant-message">%s</div>`,
					assistantDivID, renderMarkdown(content)),
				datastar.WithSelectorID(assistantDivID),
				datastar.WithModeOuter(),
			)
			if done {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *Server) handleResumeStream(w http.ResponseWriter, r *http.Request) {
	threadID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid thread id", http.StatusBadRequest)
		return
	}

	stream, ok := s.streams.Get(threadID)
	if !ok {
		slog.Info("resume: no active stream", "threadId", threadID)
		sse := datastar.NewSSE(w, r)
		sse.MarshalAndPatchSignals(map[string]any{
			"loading": false,
		})
		return
	}

	slog.Info("resume: relaying active stream", "threadId", threadID, "messageId", stream.messageID)

	sse := datastar.NewSSE(w, r)
	assistantDivID := fmt.Sprintf("msg-%s", stream.messageID)

	s.relayStream(r.Context(), sse, stream, assistantDivID)

	slog.Info("resume: relay done", "threadId", threadID)
	sse.MarshalAndPatchSignals(map[string]any{
		"loading": false,
	})
}

func (s *Server) generateThreadTitle(ctx context.Context, userMessage string) string {
	resp, err := openaiPkg.Client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openai.ChatModelGPT5Nano,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("Generate a concise title (max 6 words) for a chat thread based on the user's first message. Return only the title, no quotes or punctuation."),
			openai.UserMessage(userMessage),
		},
	})
	if err != nil {
		slog.Error("generate title", "err", err)
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
