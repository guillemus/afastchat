package server

import (
	"fmt"

	"github.com/asaskevich/EventBus"
	"github.com/google/uuid"
)

type threadEvent struct {
	isFirst    bool
	isFromUser bool
	messageID  uuid.UUID
	message    string
}

func publishThreadMessage(bus EventBus.Bus, threadID uuid.UUID, evt threadEvent) {
	topic := fmt.Sprintf("threads:%s:message", threadID.String())
	bus.Publish(topic, evt)
}

func subscribeToThread(
	bus EventBus.Bus, threadID uuid.UUID, handler func(threadEvent),
) (unsub func()) {
	topic := fmt.Sprintf("threads:%s:message", threadID.String())
	bus.Subscribe(topic, handler)

	return func() {
		bus.Unsubscribe(topic, handler)
	}
}
