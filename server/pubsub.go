package server

import (
	"fmt"

	"github.com/asaskevich/EventBus"
)

type (
	threadEvent       interface{ isThreadEvent() }
	evtAssistantStart struct {
		userMessageID      string
		userMessage        string
		assistantMessageID string
	}
	evtAssistantChunk struct {
		messageID string
		content   string
	}
	evtFinishedStreaming struct{}
	evtTitleUpdate       struct {
		newTitle string
	}
)

func (evtAssistantStart) isThreadEvent()    {}
func (evtAssistantChunk) isThreadEvent()    {}
func (evtFinishedStreaming) isThreadEvent() {}
func (evtTitleUpdate) isThreadEvent()       {}

func (x *Server) publishThreadEvent(threadID string, evt threadEvent) {
	topic := fmt.Sprintf("threads:%s", threadID)
	x.bus.Publish(topic, evt)
}

func subscribeToThread(
	bus EventBus.Bus, threadID string, handler func(threadEvent),
) (unsub func()) {
	topic := fmt.Sprintf("threads:%s", threadID)
	bus.Subscribe(topic, handler)

	return func() {
		bus.Unsubscribe(topic, handler)
	}
}

// We keep a separate topic so that it's simple to communicate the stop between a request and
// the agent streaming goroutine.

func publishThreadStop(bus EventBus.Bus, threadID string) {
	topic := fmt.Sprintf("threads:%s:stop", threadID)
	bus.Publish(topic)
}

func subscribeToStops(bus EventBus.Bus, threadID string, handler func()) (unsub func()) {
	topic := fmt.Sprintf("threads:%s:stop", threadID)
	bus.Subscribe(topic, handler)

	return func() {
		bus.Unsubscribe(topic, handler)
	}
}
