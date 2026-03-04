package server

import (
	"strings"
	"sync"

	"github.com/google/uuid"
)

type StreamRegistry struct {
	mu      sync.RWMutex
	streams map[uuid.UUID]*ActiveStream
}

func newStreamRegistry() *StreamRegistry {
	return &StreamRegistry{
		streams: make(map[uuid.UUID]*ActiveStream),
	}
}

func (r *StreamRegistry) Get(threadID uuid.UUID) (*ActiveStream, bool) {
	r.mu.RLock()
	s, ok := r.streams[threadID]
	r.mu.RUnlock()
	return s, ok
}

func (r *StreamRegistry) Set(threadID uuid.UUID, s *ActiveStream) {
	r.mu.Lock()
	r.streams[threadID] = s
	r.mu.Unlock()
}

func (r *StreamRegistry) Delete(threadID uuid.UUID) {
	r.mu.Lock()
	delete(r.streams, threadID)
	r.mu.Unlock()
}

type ActiveStream struct {
	mu        sync.RWMutex
	threadID  uuid.UUID
	messageID uuid.UUID
	content   strings.Builder
	done      bool
	listeners map[chan struct{}]struct{}
}

func newActiveStream(threadID, messageID uuid.UUID) *ActiveStream {
	return &ActiveStream{
		threadID:  threadID,
		messageID: messageID,
		listeners: make(map[chan struct{}]struct{}),
	}
}

func (s *ActiveStream) Append(token string) {
	s.mu.Lock()
	s.content.WriteString(token)
	listeners := make([]chan struct{}, 0, len(s.listeners))
	for ch := range s.listeners {
		listeners = append(listeners, ch)
	}
	s.mu.Unlock()

	for _, ch := range listeners {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

func (s *ActiveStream) Finish() {
	s.mu.Lock()
	s.done = true
	listeners := make([]chan struct{}, 0, len(s.listeners))
	for ch := range s.listeners {
		listeners = append(listeners, ch)
	}
	s.mu.Unlock()

	for _, ch := range listeners {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

func (s *ActiveStream) Subscribe() (ch chan struct{}, content string, done bool) {
	ch = make(chan struct{}, 1)
	s.mu.Lock()
	s.listeners[ch] = struct{}{}
	content = s.content.String()
	done = s.done
	s.mu.Unlock()
	return
}

func (s *ActiveStream) Unsubscribe(ch chan struct{}) {
	s.mu.Lock()
	delete(s.listeners, ch)
	s.mu.Unlock()
}

func (s *ActiveStream) Snapshot() (content string, done bool) {
	s.mu.RLock()
	content = s.content.String()
	done = s.done
	s.mu.RUnlock()
	return
}
