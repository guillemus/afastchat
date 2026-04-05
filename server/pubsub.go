package server

import (
	"fmt"

	"github.com/asaskevich/EventBus"
	"github.com/google/uuid"
)

func publishThreadMessage(bus EventBus.Bus, threadID uuid.UUID, msg string) {
	topic := fmt.Sprintf("threads:%s:message", threadID.String())
	bus.Publish(topic, msg)
}

func subscribeToThread(bus EventBus.Bus, threadID uuid.UUID, handler func(string)) (unsub func()) {
	topic := fmt.Sprintf("threads:%s:message", threadID.String())
	bus.Subscribe(topic, handler)

	return func() {
		bus.Unsubscribe(topic, handler)
	}
}
