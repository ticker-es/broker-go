package memory

import (
	"context"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	es "github.com/ticker-es/client-go/eventstream/base"
)

type testHistory struct{ EventStore }

func newHistory() *testHistory {
	return &testHistory{
		EventStore: EventStore{
			events:         []*es.Event{},
			deadAggregates: make(map[string]bool),
		},
	}
}

func (h *testHistory) event(aggregate []string, tp string, payload map[string]interface{}) *testHistory {
	h.Store(&es.Event{
		Aggregate: aggregate,
		Type:      tp,
		Payload:   payload,
	})
	return h
}

func (h testHistory) store() EventStore {
	return h.EventStore
}

var _ = Describe("memory/eventstore", func() {
	aggregate1 := []string{"de", "customer", "1"}
	aggregate2 := []string{"de", "customer", "2"}

	Context(".Store()", func() {
		Context("general event", func() {
			It("Stores an event", func() {
				event := es.Event{}
				memoryEventStore := NewMemoryEventStore()

				Expect(memoryEventStore.LastKnownSequence()).To(Equal(int64(0)))

				sequence, err := memoryEventStore.Store(&event)

				Expect(err).To(BeNil())
				Expect(sequence).To(Equal(int64(1)))
				Expect(event.Sequence).To(Equal(sequence))
				Expect(memoryEventStore.LastKnownSequence()).To(Equal(int64(1)))
			})
		})

		Context("tombstone event", func() {
			It("Marks aggregate as dead and anonymizes related events", func() {
				store := newHistory().
					event(aggregate1, "registered", map[string]interface{}{}).
					event(aggregate1, "updated", map[string]interface{}{}).
					event(aggregate2, "registered", map[string]interface{}{}).
					event(aggregate2, "updated", map[string]interface{}{}).
					store()

				Expect(store.deadAggregates).To(HaveLen(0))

				_, err := store.Store(&es.Event{Aggregate: aggregate2, Type: TombstoneEventType, Payload: map[string]interface{}{}})
				Expect(err).To(BeNil())

				Expect(store.deadAggregates).To(HaveKey(strings.Join(aggregate2, "/")))
				Expect(store.deadAggregates).NotTo(HaveKey(strings.Join(aggregate1, "/")))

				for _, event := range store.events {
					if stringifyAggregate(event.Aggregate) == stringifyAggregate(aggregate1) {
						Expect(event.Payload).NotTo(BeNil())
					} else {
						if event.Type == TombstoneEventType {
							Expect(event.Payload).NotTo(BeNil())
						} else {
							Expect(event.Payload).To(BeNil())
						}
					}
				}

				sequence, err := store.Store(&es.Event{Aggregate: aggregate2, Type: "updated"})
				Expect(err).To(Equal(ErrAttemptToStoreIntoDeadAggregate))
				Expect(sequence).To(BeZero())

			})
		})
	})

	Context(".ReadAll()", func() {
		It("returns events for live aggregates and tombstone events for dead aggregates", func() {
			store := newHistory().
				event(aggregate1, "registered", map[string]interface{}{}).
				event(aggregate1, "updated", map[string]interface{}{}).
				event(aggregate2, "registered", map[string]interface{}{}).
				event(aggregate2, "updated", map[string]interface{}{}).
				event(aggregate2, TombstoneEventType, map[string]interface{}{}).
				store()

			events := map[string]int{}
			tombstones := map[string]int{}

			store.ReadAll(context.Background(), es.Selector{Aggregate: []string{}}, es.Range(1, store.LastKnownSequence()), func(event *es.Event) error {
				aggregate := stringifyAggregate(event.Aggregate)
				events[aggregate] += 1
				if event.Type == TombstoneEventType {
					tombstones[aggregate] += 1
				}
				return nil
			})

			Expect(events[stringifyAggregate(aggregate1)]).To(Equal(2))
			Expect(events[stringifyAggregate(aggregate2)]).To(Equal(1))
			Expect(tombstones[stringifyAggregate(aggregate2)]).To(Equal(1))
		})
	})

})
