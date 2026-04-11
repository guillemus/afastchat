package server

import (
	"fmt"

	"github.com/asaskevich/EventBus"
)

type (
	threadEvent struct {
		assistantStart    *evtAssistantStart
		assistantChunk    *evtAssistantChunk
		finishedStreaming *evtFinishedStreaming
	}
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
)

func publishThreadMessage(bus EventBus.Bus, threadID string, evt threadEvent) {
	topic := fmt.Sprintf("threads:%s:message", threadID)
	bus.Publish(topic, evt)
}

func publishThreadFinishedStreaming(bus EventBus.Bus, threadID string) {
	publishThreadMessage(bus, threadID, threadEvent{finishedStreaming: &evtFinishedStreaming{}})
}

func subscribeToThread(
	bus EventBus.Bus, threadID string, handler func(threadEvent),
) (unsub func()) {
	topic := fmt.Sprintf("threads:%s:message", threadID)
	bus.Subscribe(topic, handler)

	return func() {
		bus.Unsubscribe(topic, handler)
	}
}

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
