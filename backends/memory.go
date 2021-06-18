package backends

import (
	c "github.com/mtrense/soil/config"
	"github.com/ticker-es/broker-go/backends/memory"
	"github.com/ticker-es/client-go/eventstream/base"
)

type MemoryFactory struct {
}

func (s *MemoryFactory) Names() []string {
	return []string{"memory", "mem"}
}

func (s *MemoryFactory) CreateEventStore() base.EventStore {
	return memory.NewMemoryEventStore()
}

func (s *MemoryFactory) BuildEventStoreFlags() c.Applicant {
	return func(b *c.Command) {}
}

func (s *MemoryFactory) CreateSequenceStore() base.SequenceStore {
	return memory.NewMemorySequenceStore()
}

func (s *MemoryFactory) BuildSequenceStoreFlags() c.Applicant {
	return func(b *c.Command) {}
}

func init() {
	RegisterEventStore(&MemoryFactory{})
	RegisterSequenceStore(&MemoryFactory{})
}
