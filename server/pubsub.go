package server

import (
	"fmt"

	"github.com/asaskevich/EventBus"
)

type threadEvent struct {
	isFirst    bool
	isFromUser bool
	messageID  string
	message    string
}

func publishThreadMessage(bus EventBus.Bus, threadID string, evt threadEvent) {
	topic := fmt.Sprintf("threads:%s:message", threadID)
	bus.Publish(topic, evt)
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
