package backends

import (
	"github.com/mtrense/soil/config"
	"github.com/ticker-es/client-go/eventstream/base"
)

type EventStoreFactory interface {
	Names() []string
	CreateEventStore() base.EventStore
	BuildEventStoreFlags() config.Applicant
}

type SequenceStoreFactory interface {
	Names() []string
	CreateSequenceStore() base.SequenceStore
	BuildSequenceStoreFlags() config.Applicant
}

var (
	eventStores          = make([]EventStoreFactory, 0)
	eventStoresByName    = make(map[string]EventStoreFactory)
	sequenceStores       = make([]SequenceStoreFactory, 0)
	sequenceStoresByName = make(map[string]SequenceStoreFactory)
)

func RegisterEventStore(f EventStoreFactory) {
	for _, name := range f.Names() {
		if _, present := eventStoresByName[name]; present {
			panic("Two EventStores with the same name")
		}
		eventStoresByName[name] = f
	}
	eventStores = append(eventStores, f)
}

func RegisterSequenceStore(f SequenceStoreFactory) {
	for _, name := range f.Names() {
		if _, present := sequenceStoresByName[name]; present {
			panic("Two EventStores with the same name")
		}
		sequenceStoresByName[name] = f
	}
	sequenceStores = append(sequenceStores, f)
}

func GetAllConfiguredFlags() config.Applicant {
	var flags []config.Applicant
	for _, s := range eventStores {
		flags = append(flags, s.BuildEventStoreFlags())
	}
	for _, s := range sequenceStores {
		flags = append(flags, s.BuildSequenceStoreFlags())
	}
	return func(b *config.Command) {
		b.Apply(flags...)
	}
}

func EventStores() []EventStoreFactory {
	return eventStores
}

func SequenceStores() []SequenceStoreFactory {
	return sequenceStores
}

func LookupEventStore(name string) EventStoreFactory {
	return eventStoresByName[name]
}

func LookupSequenceStore(name string) SequenceStoreFactory {
	return sequenceStoresByName[name]
}
