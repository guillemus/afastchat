package server

import (
	"fmt"

	"github.com/asaskevich/EventBus"
)

func (x *Server) publishThreadEvent(threadID string) {
	topic := fmt.Sprintf("threads:%s", threadID)
	x.bus.Publish(topic)
}

func (x *Server) subscribeToThread(threadID string, handler func()) (unsub func()) {
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
