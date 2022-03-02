package memory

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	es "github.com/ticker-es/client-go/eventstream/base"
)

var _ = Describe("memory/eventstore", func() {
	Context("general", func() {
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

	Context("tombstone", func() {
		aggregate1 := []string{"de", "customer", "1"}
		aggregate2 := []string{"de", "customer", "2"}

		When("Event is stored", func() {
			store := EventStore{
				events:         make([]*es.Event, 0),
				deadAggregates: make(map[string]bool),
			}

			type record struct {
				Aggregate []string
				Type      string
				Payload   map[string]interface{}
			}

			history := []record{
				{aggregate1, "registered", map[string]interface{}{}},
				{aggregate1, "updated", map[string]interface{}{}},
				{aggregate2, "registered", map[string]interface{}{}},
				{aggregate2, "updated", map[string]interface{}{}},
			}

			for _, r := range history {
				store.Store(&es.Event{
					Aggregate: r.Aggregate, Type: r.Type, Payload: r.Payload,
				})
			}

			It("Marks aggregate as dead and anonymizes related events", func() {
				Expect(store.deadAggregates).To(HaveLen(0))

				_, err := store.Store(&es.Event{Aggregate: aggregate2, Type: "$tombstone", Payload: map[string]interface{}{}})
				Expect(err).To(BeNil())

				Expect(store.deadAggregates).To(HaveKey(strings.Join(aggregate2, "/")))
				Expect(store.deadAggregates).NotTo(HaveKey(strings.Join(aggregate1, "/")))

				for _, event := range store.events {
					if stringifyAggregate(event.Aggregate) == stringifyAggregate(aggregate1) {
						Expect(event.Payload).NotTo(BeNil())
					} else {
						if event.Type == "$tombstone" {
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
})
