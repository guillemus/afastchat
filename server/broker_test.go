package server

import (
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
)

func newTestBroker(t *testing.T) *Broker {
	broker, err := NewBroker()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, broker.Close()) })
	return broker
}

func receive(t *testing.T, messages chan *nats.Msg) *nats.Msg {
	select {
	case msg := <-messages:
		return msg
	case <-time.After(time.Second):
		require.Fail(t, "event not received")
		return nil
	}
}

func requireNothing(t *testing.T, messages chan *nats.Msg) {
	select {
	case msg := <-messages:
		require.Failf(t, "unexpected event", "subject %s", msg.Subject)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestUserThreadsRouteOnlyToOwner(t *testing.T) {
	broker := newTestBroker(t)

	alice := make(chan *nats.Msg, 1)
	unsubscribeAlice, err := broker.SubscribeSubjects(alice, UserThreadsSubject("alice"))
	require.NoError(t, err)
	t.Cleanup(unsubscribeAlice)

	bob := make(chan *nats.Msg, 1)
	unsubscribeBob, err := broker.SubscribeSubjects(bob, UserThreadsSubject("bob"))
	require.NoError(t, err)
	t.Cleanup(unsubscribeBob)

	broker.PublishUserThread("alice", "thread-1")

	msg := receive(t, alice)
	require.Equal(t, "thread-1", string(msg.Data))
	requireNothing(t, bob)
}

func TestUnsubscribeOnlyRemovesThatSubscriber(t *testing.T) {
	broker := newTestBroker(t)

	stay := make(chan *nats.Msg, 1)
	unsubscribeStay, err := broker.SubscribeSubjects(stay, ThreadSubject("t"))
	require.NoError(t, err)
	t.Cleanup(unsubscribeStay)

	leave := make(chan *nats.Msg, 1)
	unsubscribeLeave, err := broker.SubscribeSubjects(leave, ThreadSubject("t"))
	require.NoError(t, err)
	unsubscribeLeave()

	broker.PublishThread("t", threadEvent{UserMessageCreated: true})

	msg := receive(t, stay)
	require.JSONEq(t, `{"user_message_created":true}`, string(msg.Data))
	requireNothing(t, leave)
}

func TestThreadStopIsSeparateFromThreadEvents(t *testing.T) {
	broker := newTestBroker(t)

	events := make(chan *nats.Msg, 1)
	unsubscribeEvents, err := broker.SubscribeSubjects(events, ThreadSubject("t"))
	require.NoError(t, err)
	t.Cleanup(unsubscribeEvents)

	stops := make(chan *nats.Msg, 1)
	_, unsubscribeStops, err := broker.SubscribeThreadStops("t", stops)
	require.NoError(t, err)
	t.Cleanup(unsubscribeStops)

	broker.PublishThreadStop("t")

	receive(t, stops)
	requireNothing(t, events)
}
