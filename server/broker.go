package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

type Broker struct {
	server *natsserver.Server
	conn   *nats.Conn
}

func NewBroker() (*Broker, error) {
	ns, err := natsserver.NewServer(&natsserver.Options{
		Host:   "127.0.0.1",
		Port:   -1,
		NoSigs: true,
		NoLog:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("create embedded NATS server: %w", err)
	}

	go ns.Start()
	if !ns.ReadyForConnections(5 * time.Second) {
		ns.Shutdown()
		return nil, fmt.Errorf("embedded NATS server did not become ready")
	}

	conn, err := nats.Connect(
		ns.ClientURL(), nats.Name("afastchat.com"),
		nats.ErrorHandler(func(_ *nats.Conn, subscription *nats.Subscription, err error) {
			subject := ""
			if subscription != nil {
				subject = subscription.Subject
			}
			slog.Error("NATS async error", "subject", subject, "err", err)
		}),
	)
	if err != nil {
		ns.Shutdown()
		ns.WaitForShutdown()
		return nil, fmt.Errorf("connect to embedded NATS server: %w", err)
	}

	return &Broker{server: ns, conn: conn}, nil
}

func ThreadSubject(threadID string) string {
	return fmt.Sprintf("threads.%s", threadID)
}

func ThreadStopSubject(threadID string) string {
	return fmt.Sprintf("threads.%s.stop", threadID)
}

func UserThreadsSubject(userID string) string {
	return fmt.Sprintf("users.%s.threads", userID)
}

type threadEvent struct {
	UserMessageCreated bool `json:"user_message_created,omitempty"`
}

func (b *Broker) PublishThread(threadID string, event threadEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		slog.Error("encode thread event", "thread_id", threadID, "err", err)
		return
	}

	b.publish(ThreadSubject(threadID), data)
}

func (b *Broker) PublishThreadUpdated(threadID string) {
	b.PublishThread(threadID, threadEvent{})
}

func (b *Broker) PublishThreadStop(threadID string) {
	b.publish(ThreadStopSubject(threadID), nil)
}

func (b *Broker) PublishUserThread(userID, threadID string) {
	b.publish(UserThreadsSubject(userID), []byte(threadID))
}

func (b *Broker) SubscribeThreadStops(
	threadID string, messages chan *nats.Msg,
) (*nats.Subscription, func(), error) {
	return b.subscribe(ThreadStopSubject(threadID), messages)
}

func (b *Broker) SubscribeSubjects(messages chan *nats.Msg, subjects ...string) (func(), error) {
	unsubscribes := make([]func(), 0, len(subjects))
	unsubscribe := func() {
		for _, unsubscribe := range unsubscribes {
			unsubscribe()
		}
	}

	for _, subject := range subjects {
		_, unsubscribeSubject, err := b.subscribe(subject, messages)
		if err != nil {
			unsubscribe()
			return nil, err
		}

		unsubscribes = append(unsubscribes, unsubscribeSubject)
	}

	return unsubscribe, nil
}

func (b *Broker) publish(subject string, data []byte) {
	slog.Debug("publish NATS event", "subject", subject)

	if err := b.conn.Publish(subject, data); err != nil {
		slog.Error("publish", "subject", subject, "err", err)
	}
}

func (b *Broker) subscribe(
	subject string,
	messages chan *nats.Msg,
) (*nats.Subscription, func(), error) {
	subscription, err := b.conn.ChanSubscribe(subject, messages)
	if err != nil {
		return nil, nil, fmt.Errorf("subscribe to %s: %w", subject, err)
	}

	if err := b.conn.Flush(); err != nil {
		if unsubscribeErr := subscription.Unsubscribe(); unsubscribeErr != nil {
			slog.Debug("unsubscribe after flush failure", "subject", subject, "err", unsubscribeErr)
		}
		return nil, nil, fmt.Errorf("flush subscription to %s: %w", subject, err)
	}

	unsubscribe := func() {
		if err := subscription.Unsubscribe(); err != nil {
			slog.Debug("unsubscribe", "subject", subject, "err", err)
		}
	}

	return subscription, unsubscribe, nil
}

func (b *Broker) Close() error {
	var drainErr error
	if b.conn != nil {
		drainErr = b.conn.Drain()
		b.conn.Close()
	}

	if b.server != nil {
		b.server.Shutdown()
		b.server.WaitForShutdown()
	}

	if drainErr != nil {
		return fmt.Errorf("drain NATS connection: %w", drainErr)
	}

	return nil
}

func (x *Server) cancelOnThreadStop(ctx context.Context, threadID string, cancel context.CancelFunc) error {
	stops := make(chan *nats.Msg, 1)
	_, unsubscribe, err := x.broker.SubscribeThreadStops(threadID, stops)
	if err != nil {
		return err
	}

	GoRecover(func() {
		defer unsubscribe()
		select {
		case <-stops:
			cancel()
		case <-ctx.Done():
		}
	})

	return nil
}
