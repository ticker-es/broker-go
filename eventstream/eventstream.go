package eventstream

import (
	"context"
	"sync"
	"time"

	es "github.com/ticker-es/client-go/eventstream/base"
)

type EventStream struct {
	eventStore        es.EventStore
	writeLock         sync.Mutex
	sequenceStore     es.SequenceStore
	subscriptions     map[string]*Subscription
	defaultBufferSize int
}

type Option = func(s *EventStream)

func NewEventStream(eventStore es.EventStore, sequenceStore es.SequenceStore, opts ...Option) *EventStream {
	s := &EventStream{
		eventStore:        eventStore,
		defaultBufferSize: 100,
		sequenceStore:     sequenceStore,
		subscriptions:     make(map[string]*Subscription),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *EventStream) Emit(event *es.Event) (int64, error) {
	s.writeLock.Lock()
	defer s.writeLock.Unlock()
	seq, err := s.eventStore.Store(event)
	if err != nil {
		return seq, err
	}
	event.Sequence = seq
	for _, sub := range s.subscriptions {
		if sub.active {
			sub.publishEvent(event)
		}
	}
	return event.Sequence, nil
}

func (s *EventStream) LastSequence() int64 {
	return s.eventStore.LastKnownSequence()
}

func (s *EventStream) Get(sequence int64) (*es.Event, error) {
	return s.eventStore.Read(sequence)
}

func (s *EventStream) Stream(ctx context.Context, sel es.Selector, bracket es.Bracket, handler es.EventHandler) error {
	bracket.Sanitize(s.LastSequence())
	return s.eventStore.ReadAll(ctx, sel, bracket, handler)
}

func (s *EventStream) Listen(ctx context.Context, sel es.Selector, handler es.EventHandler) error {
	return nil
}

func (s *EventStream) Subscribe(ctx context.Context, persistentClientID string, sel es.Selector, handler es.EventHandler) (es.Subscription, error) {
	sub := s.getOrCreateSubscription(persistentClientID, sel)
	err := sub.handleSubscription(ctx, handler)
	return sub, err
}

func (s *EventStream) getOrCreateSubscription(persistentClientID string, sel es.Selector) *Subscription {
	s.writeLock.Lock()
	defer s.writeLock.Unlock()
	sub, present := s.subscriptions[persistentClientID]
	if !present {
		sub = newSubscription(s, persistentClientID, sel)
		s.subscriptions[persistentClientID] = sub
	}
	return sub
}

func (s *EventStream) attachSubscription(sub *Subscription) (int64, error) {
	s.writeLock.Lock()
	defer s.writeLock.Unlock()
	sub.buffer = make(chan *es.Event, s.defaultBufferSize)
	sub.active = true
	sub.live = true
	return s.LastSequence(), nil
}

func (s *EventStream) unsubscribe(sub *Subscription) {
	s.writeLock.Lock()
	defer s.writeLock.Unlock()
	sub.active = false
	sub.inactiveSince = time.Now()
	close(sub.buffer)
}

func (s *EventStream) Acknowledge(persistentClientID string, sequence int64) error {
	return s.sequenceStore.Store(persistentClientID, sequence)
}

func (s *EventStream) Subscriptions() []es.Subscription {
	result := make([]es.Subscription, 0, len(s.subscriptions))
	for _, sub := range s.subscriptions {
		result = append(result, sub)
	}
	return result
}
