package memory

import (
	"context"
	"errors"
	"strings"

	es "github.com/ticker-es/client-go/eventstream/base"
)

const (
	TombstoneEventType = "$tombstone"
)

var (
	ErrAttemptToStoreIntoDeadAggregate = errors.New("attempt to store into dead aggregate")
)

type EventStore struct {
	events         []*es.Event
	deadAggregates map[string]bool
}

func NewMemoryEventStore() es.EventStore {
	return &EventStore{
		events:         make([]*es.Event, 0),
		deadAggregates: make(map[string]bool),
	}
}

func (s *EventStore) Store(event *es.Event) (int64, error) {
	if s.IsAggredateDead(event.Aggregate) {
		return 0, ErrAttemptToStoreIntoDeadAggregate
	}

	// Sequence starts at 1
	seq := int64(len(s.events) + 1)
	event.Sequence = seq
	s.events = append(s.events, event)

	if event.Type == TombstoneEventType {
		s.markAggregareAsDead(event)
		s.anonymizeAggregate(event.Aggregate)
	}

	return seq, nil
}

func (s *EventStore) LastKnownSequence() int64 {
	return int64(len(s.events))
}

func (s *EventStore) Read(sequence int64) (*es.Event, error) {
	return s.events[sequence-1], nil
}

func (s *EventStore) ReadAll(ctx context.Context, sel es.Selector, bracket es.Bracket, handler es.EventHandler) error {
	// TODO: es.All() doesn't work here: runtime error: slice bounds out of range [:9223372036854775807] with capacity 8
	for _, event := range s.events[bracket.NextSequence-1 : bracket.LastSequence] {
		if err := ctx.Err(); errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil
		}

		if sel.Matches(event) {
			var err error
			if s.IsAggredateDead(event.Aggregate) {
				if event.Type == TombstoneEventType {
					err = handler(event)
				}
			} else {
				err = handler(event)
			}

			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s EventStore) IsAggredateDead(aggregate []string) (result bool) {
	return s.deadAggregates[stringifyAggregate(aggregate)]
}

func (s *EventStore) markAggregareAsDead(event *es.Event) {
	s.deadAggregates[stringifyAggregate(event.Aggregate)] = true
}

func (s *EventStore) anonymizeAggregate(aggregate []string) {
	for _, event := range s.events {
		if (stringifyAggregate(event.Aggregate) == stringifyAggregate(aggregate)) && (event.Type != TombstoneEventType) {
			event.Payload = nil
		}
	}
}

func stringifyAggregate(aggregate []string) string {
	return strings.Join(aggregate, "/")
}
