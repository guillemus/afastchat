package bench

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"afastchat.com/config"
	"afastchat.com/db"
	"afastchat.com/openrouter"
	"afastchat.com/server"

	"github.com/alexedwards/scs/v2"
	"github.com/andybalholm/brotli"
)

var (
	benchUsers   = flag.String("bench.users", "1,10,50,100", "comma separated concurrent user counts")
	benchTokens  = flag.Int("bench.tokens", 300, "answer tokens streamed per message")
	benchRate    = flag.Int("bench.rate", 100, "tokens per second per stream")
	benchHistory = flag.Int("bench.history", 5, "previous turns seeded per thread")
	benchMath    = flag.Bool("bench.math", false, "include LaTeX math in answers")
)

const sessionCookie = "afastchat_session"

const benchAnswer = `## Overview

Here is what I found after looking at [the official docs](https://example.com/docs) and [a benchmark write-up](https://example.com/bench).

- **Latency** stays low while the working set fits in memory.
- **Throughput** depends on how often you write to disk.
- Connection pooling matters more than raw query speed.

` + "```go\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n```" + `

In short, measure first and only then optimize the hot path.

`

const benchMathAnswer = `The cost grows as $O(n^2)$ because each update rewrites $\sum_{i=1}^{n} i$ bytes.

$$T(n) = \frac{n(n+1)}{2}$$

`

type app struct {
	url      string
	pool     *db.Pool
	sessions *scs.SessionManager
	keys     *openrouter.KeyCipher
}

var (
	appOnce sync.Once
	appEnv  app
	appErr  error
)

func startApp(b *testing.B) app {
	appOnce.Do(func() { appEnv, appErr = bootApp() })
	if appErr != nil {
		b.Fatal(appErr)
	}
	return appEnv
}

func bootApp() (app, error) {
	dir := filepath.Join("..", "tmp", "bench")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return app{}, err
	}

	logFile, err := os.Create(filepath.Join(dir, "server.log"))
	if err != nil {
		return app{}, err
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{Level: slog.LevelWarn})))
	dbPath := filepath.Join(dir, "bench.db")
	for _, suffix := range []string{"", "-wal", "-shm"} {
		_ = os.Remove(dbPath + suffix)
	}

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return app{}, err
	}
	tokenKey := base64.StdEncoding.EncodeToString(key)

	llm := newFakeOpenRouter(*benchTokens, *benchRate)
	if err := os.Setenv("OPENROUTER_BASE_URL", llm.URL); err != nil {
		return app{}, err
	}

	port, err := freePort()
	if err != nil {
		return app{}, err
	}
	go server.StartServer(config.Config{
		DBPath:             dbPath,
		Port:               port,
		OpenRouterTokenKey: tokenKey,
	})

	url := "http://127.0.0.1:" + port
	if err := waitReady(url + "/login"); err != nil {
		return app{}, err
	}

	pool, err := db.Connect(dbPath)
	if err != nil {
		return app{}, err
	}
	keys, err := openrouter.NewKeyCipher(tokenKey)
	if err != nil {
		return app{}, err
	}
	sessions := scs.New()
	sessions.Store = db.NewSCSStore(pool)

	return app{url: url, pool: pool, sessions: sessions, keys: keys}, nil
}

func freePort() (string, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	defer l.Close()
	return strconv.Itoa(l.Addr().(*net.TCPAddr).Port), nil
}

func waitReady(url string) error {
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		res, err := http.Get(url)
		if err == nil {
			_ = res.Body.Close()
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("server at %s never became ready", url)
}

func BenchmarkStreaming(b *testing.B) {
	for _, users := range parseUserCounts(b) {
		b.Run(fmt.Sprintf("users=%d", users), func(b *testing.B) {
			benchmarkStreaming(b, users)
		})
	}
}

func benchmarkStreaming(b *testing.B, userCount int) {
	env := startApp(b)

	var heapBefore runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&heapBefore)

	clients := make([]*client, userCount)
	for i := range clients {
		clients[i] = seedClient(b, env)
	}
	defer func() {
		for _, it := range clients {
			it.close()
		}
	}()
	for _, it := range clients {
		it.connect(b)
	}

	var heapIdle runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&heapIdle)

	sampler := startMemSampler()
	stagger := time.Second / time.Duration(max(userCount, 1))

	for b.Loop() {
		var wg sync.WaitGroup
		for i, it := range clients {
			wg.Go(func() {
				time.Sleep(time.Duration(i) * stagger)
				it.sendAndWait(b)
			})
		}
		wg.Wait()
	}
	b.StopTimer()

	peakHeap, goroutines := sampler.stop()

	var stats streamStats
	for _, it := range clients {
		stats.merge(it.stats)
	}

	ideal := time.Duration(*benchTokens+*benchTokens/10) * time.Second / time.Duration(*benchRate)
	b.ReportMetric(ms(percentileDur(stats.firstPatch, 0.50)), "ttfp-p50-ms")
	b.ReportMetric(ms(percentileDur(stats.firstPatch, 0.99)), "ttfp-p99-ms")
	b.ReportMetric(ms(percentileDur(stats.gaps, 0.50)), "gap-p50-ms")
	b.ReportMetric(ms(percentileDur(stats.gaps, 0.99)), "gap-p99-ms")
	b.ReportMetric(ms(percentileDur(stats.done, 0.50)-ideal), "lag-p50-ms")
	b.ReportMetric(ms(percentileDur(stats.done, 0.99)-ideal), "lag-p99-ms")
	b.ReportMetric(float64(stats.wireBytes)/1024/float64(max(len(stats.done), 1)), "wire-KB/msg")
	b.ReportMetric(max(float64(heapIdle.HeapInuse)-float64(heapBefore.HeapInuse), 0)/1024/float64(userCount), "idle-heap-KB/user")
	b.ReportMetric(float64(peakHeap)/1024/1024, "peak-heap-MB")
	b.ReportMetric(float64(goroutines), "peak-goroutines")
}

func parseUserCounts(b *testing.B) []int {
	var out []int
	for it := range strings.SplitSeq(*benchUsers, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(it))
		if err != nil {
			b.Fatalf("parse -bench.users: %v", err)
		}
		out = append(out, n)
	}
	return out
}

var clientSeq atomic.Int64

type client struct {
	baseURL  string
	cookie   string
	threadID string

	events io.ReadCloser
	wire   *countingReader
	stream chan sseEvent

	stats streamStats
}

type sseEvent struct {
	at   time.Time
	data string
}

func seedClient(b *testing.B, env app) *client {
	ctx := context.Background()
	n := clientSeq.Add(1)

	user, err := env.pool.LoginWithIdentity(ctx, db.LoginWithIdentityParams{
		Provider:    db.AuthProviderGoogle,
		Subject:     fmt.Sprintf("bench-%d", n),
		Email:       fmt.Sprintf("bench-%d@example.com", n),
		DisplayName: fmt.Sprintf("Bench %d", n),
	})
	if err != nil {
		b.Fatal(err)
	}

	if err := env.pool.SaveOpenRouterConnection(ctx, user.ID, env.keys.Seal("sk-bench")); err != nil {
		b.Fatal(err)
	}

	threadID, err := env.pool.InitThread(ctx, user.ID, "Bench thread", "You are a helpful assistant")
	if err != nil {
		b.Fatal(err)
	}
	for range *benchHistory {
		if _, err := env.pool.CreateUserMessage(ctx, user.ID, threadID, "What should I know about database performance?"); err != nil {
			b.Fatal(err)
		}
		if _, err := env.pool.CreateMessage(ctx, user.ID, threadID, db.AssistantMessage, answerText()); err != nil {
			b.Fatal(err)
		}
	}

	sessionCtx, err := env.sessions.Load(ctx, "")
	if err != nil {
		b.Fatal(err)
	}
	env.sessions.Put(sessionCtx, "user_id", user.ID)
	token, _, err := env.sessions.Commit(sessionCtx)
	if err != nil {
		b.Fatal(err)
	}

	return &client{
		baseURL:  env.url,
		cookie:   sessionCookie + "=" + token,
		threadID: threadID,
	}
}

func (x *client) connect(b *testing.B) {
	req, err := http.NewRequest(http.MethodGet, x.baseURL+"/threads/"+x.threadID+"/events", nil)
	if err != nil {
		b.Fatal(err)
	}
	req.Header.Set("Cookie", x.cookie)
	req.Header.Set("Accept-Encoding", "br")
	req.Header.Set("Datastar-Request", "true")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		b.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		_ = res.Body.Close()
		b.Fatalf("events: status %d", res.StatusCode)
	}
	x.events = res.Body

	x.wire = &countingReader{r: res.Body}
	x.stream = make(chan sseEvent, 4096)

	var body io.Reader = x.wire
	if res.Header.Get("Content-Encoding") == "br" {
		body = brotli.NewReader(x.wire)
	}
	go readSSE(body, x.stream)

	if !x.waitFor(30*time.Second, func(evt sseEvent) bool {
		return strings.Contains(evt.data, `id="messages"`)
	}) {
		b.Fatal("events: initial page never arrived")
	}
}

func (x *client) sendAndWait(b *testing.B) {
	body, _ := json.Marshal(map[string]string{"message": "Tell me more", "threadId": x.threadID})
	req, err := http.NewRequest(http.MethodPost, x.baseURL+"/chat/send", bytes.NewReader(body))
	if err != nil {
		b.Error(err)
		return
	}
	req.Header.Set("Cookie", x.cookie)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Datastar-Request", "true")

	wireBefore := x.wire.n.Load()
	sentAt := time.Now()
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		b.Error(err)
		return
	}
	_, _ = io.Copy(io.Discard, res.Body)
	_ = res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b.Errorf("send: status %d", res.StatusCode)
		return
	}

	timeout := time.Duration(*benchTokens)*time.Second/time.Duration(*benchRate) + time.Minute
	var lastPatch time.Time
	if !x.waitFor(timeout, func(evt sseEvent) bool {
		if strings.Contains(evt.data, `id="messages"`) {
			if lastPatch.IsZero() {
				x.stats.firstPatch = append(x.stats.firstPatch, evt.at.Sub(sentAt))
			} else {
				x.stats.gaps = append(x.stats.gaps, evt.at.Sub(lastPatch))
			}
			lastPatch = evt.at
		}
		return strings.Contains(evt.data, `"_isStreaming":false`)
	}) {
		b.Errorf("thread %s: stream did not finish within %s", x.threadID, timeout)
		return
	}

	x.stats.done = append(x.stats.done, time.Since(sentAt))
	x.stats.wireBytes += x.wire.n.Load() - wireBefore
}

func (x *client) waitFor(timeout time.Duration, match func(sseEvent) bool) bool {
	deadline := time.After(timeout)
	for {
		select {
		case evt, ok := <-x.stream:
			if !ok {
				return false
			}
			if match(evt) {
				return true
			}
		case <-deadline:
			return false
		}
	}
}

func (x *client) close() {
	if x.events != nil {
		_ = x.events.Close()
	}
}

func readSSE(r io.Reader, out chan<- sseEvent) {
	defer close(out)

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 4096), 16*1024*1024)
	var data strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case line == "":
			if data.Len() > 0 {
				out <- sseEvent{at: time.Now(), data: data.String()}
			}
			data.Reset()
		case strings.HasPrefix(line, "data: "):
			data.WriteString(strings.TrimPrefix(line, "data: "))
			data.WriteByte('\n')
		}
	}
}

type countingReader struct {
	r io.Reader
	n atomic.Int64
}

func (x *countingReader) Read(p []byte) (int, error) {
	n, err := x.r.Read(p)
	x.n.Add(int64(n))
	return n, err
}

type streamStats struct {
	firstPatch []time.Duration
	gaps       []time.Duration
	done       []time.Duration
	wireBytes  int64
}

func (x *streamStats) merge(other streamStats) {
	x.firstPatch = append(x.firstPatch, other.firstPatch...)
	x.gaps = append(x.gaps, other.gaps...)
	x.done = append(x.done, other.done...)
	x.wireBytes += other.wireBytes
}

type memSampler struct {
	stopC      chan struct{}
	doneC      chan struct{}
	peakHeap   uint64
	goroutines int
}

func startMemSampler() *memSampler {
	x := &memSampler{stopC: make(chan struct{}), doneC: make(chan struct{})}
	go func() {
		defer close(x.doneC)
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()
		for {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			x.peakHeap = max(x.peakHeap, m.HeapInuse)
			x.goroutines = max(x.goroutines, runtime.NumGoroutine())
			select {
			case <-ticker.C:
			case <-x.stopC:
				return
			}
		}
	}()
	return x
}

func (x *memSampler) stop() (peakHeap uint64, goroutines int) {
	close(x.stopC)
	<-x.doneC
	return x.peakHeap, x.goroutines
}

func newFakeOpenRouter(tokens, rate int) *httptest.Server {
	words := strings.SplitAfter(answerText(), " ")
	interval := time.Second / time.Duration(rate)

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Stream bool `json:"stream"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		if !req.Stream {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": "gen-bench", "object": "chat.completion", "model": "bench",
				"choices": []map[string]any{{
					"index":         0,
					"message":       map[string]any{"role": "assistant", "content": "Bench thread"},
					"finish_reason": "stop",
				}},
				"usage": map[string]any{"prompt_tokens": 10, "completion_tokens": 2, "total_tokens": 12},
			})
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		flusher := w.(http.Flusher)
		send := func(v any) bool {
			payload, _ := json.Marshal(v)
			if _, err := fmt.Fprintf(w, "data: %s\n\n", payload); err != nil {
				return false
			}
			flusher.Flush()
			return true
		}
		delta := func(d map[string]any, finish any) map[string]any {
			return map[string]any{
				"id": "gen-bench", "object": "chat.completion.chunk", "model": "bench",
				"choices": []map[string]any{{"index": 0, "delta": d, "finish_reason": finish}},
			}
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		reasoningTokens := tokens / 10
		for i := range tokens + reasoningTokens {
			select {
			case <-ticker.C:
			case <-r.Context().Done():
				return
			}
			d := map[string]any{"content": words[i%len(words)]}
			if i < reasoningTokens {
				d = map[string]any{"reasoning": words[i%len(words)]}
			}
			if !send(delta(d, nil)) {
				return
			}
		}

		send(delta(map[string]any{}, "stop"))
		send(map[string]any{
			"id": "gen-bench", "object": "chat.completion.chunk", "model": "bench",
			"choices": []any{},
			"usage":   map[string]any{"prompt_tokens": 1000, "completion_tokens": tokens, "total_tokens": 1000 + tokens},
		})
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
		flusher.Flush()
	}))
}

func answerText() string {
	if *benchMath {
		return benchAnswer + benchMathAnswer
	}
	return benchAnswer
}

func percentileDur(samples []time.Duration, p float64) time.Duration {
	if len(samples) == 0 {
		return 0
	}
	sorted := slices.Clone(samples)
	slices.Sort(sorted)
	return sorted[int(math.Round(float64(len(sorted)-1)*p))]
}

func ms(d time.Duration) float64 {
	return float64(d) / float64(time.Millisecond)
}
