package memory

import (
	"context"
	"errors"

	es "github.com/ticker-es/client-go/eventstream/base"
)

type EventStore struct {
	events []*es.Event
}

func NewMemoryEventStore() es.EventStore {
	return &EventStore{
		events: make([]*es.Event, 0),
	}
}

func (s *EventStore) Store(event *es.Event) (int64, error) {
	// Sequence starts at 1
	seq := int64(len(s.events) + 1)
	event.Sequence = seq
	s.events = append(s.events, event)
	return seq, nil
}

func (s *EventStore) LastKnownSequence() int64 {
	return int64(len(s.events))
}

func (s *EventStore) Read(sequence int64) (*es.Event, error) {
	return s.events[sequence-1], nil
}

func (s *EventStore) ReadAll(ctx context.Context, sel es.Selector, bracket es.Bracket, handler es.EventHandler) error {
	for _, event := range s.events[bracket.NextSequence-1 : bracket.LastSequence] {
		if err := ctx.Err(); errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil
		}
		if sel.Matches(event) {
			if err := handler(event); err != nil {
				return err
			}
		}
	}
	return nil
}
