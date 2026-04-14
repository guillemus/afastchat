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
	evtStreamFinishedStreaming struct{}
)

func (evtAssistantStart) isThreadEvent()          {}
func (evtAssistantChunk) isThreadEvent()          {}
func (evtStreamFinishedStreaming) isThreadEvent() {}

func (x *Server) publishThreadEvent(threadID string, evt threadEvent) {
	topic := fmt.Sprintf("threads:%s", threadID)
	x.bus.Publish(topic, evt)
}

func (x *Server) subscribeToThread(threadID string, handler func(threadEvent)) (unsub func()) {
	topic := fmt.Sprintf("threads:%s", threadID)
	x.bus.Subscribe(topic, handler)

	return func() {
		x.bus.Unsubscribe(topic, handler)
	}
}

type (
	sidebarEvent    interface{ isSidebarEvent() }
	evtThreadUpdate struct {
		threadID string
	}
)

func (evtThreadUpdate) isSidebarEvent() {}

func (x *Server) subscribeToSidebar(handler func(sidebarEvent)) (unsub func()) {
	topic := "sidebar"
	x.bus.Subscribe(topic, handler)

	return func() {
		x.bus.Unsubscribe(topic, handler)
	}
}

func (x *Server) publishSidebarEvent(evt sidebarEvent) {
	topic := "sidebar"
	x.bus.Publish(topic, evt)
}

// We keep a separate topic so that flow between request to stop streaming and stream is separate from stream -> /events flow.

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
