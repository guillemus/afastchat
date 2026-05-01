package server

import (
	"fmt"

	"github.com/asaskevich/EventBus"
)

type ThreadEvent struct {
	UserMessageCreated bool
	SidebarChanged     bool
}

func (x *Server) publishThreadEvent(threadID string, evt ThreadEvent) {
	topic := fmt.Sprintf("threads:%s", threadID)
	x.bus.Publish(topic, evt)
}

func (x *Server) publishThreadDidUpdateEventEvent(threadID string) {
	x.publishThreadEvent(threadID, ThreadEvent{UserMessageCreated: false})
}

func (x *Server) subscribeToThread(threadID string, handler func(evt ThreadEvent)) (unsub func()) {
	topic := fmt.Sprintf("threads:%s", threadID)
	x.bus.Subscribe(topic, handler)

	return func() {
		x.bus.Unsubscribe(topic, handler)
	}
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
