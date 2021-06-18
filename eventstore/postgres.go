package eventstore

import (
	c "github.com/mtrense/soil/config"
	"github.com/ticker-es/client-go/eventstream/base"
)

type PostgresFactory struct {
}

func (s *PostgresFactory) Names() []string {
	return []string{"postgres", "pg"}
}

func (s *PostgresFactory) CreateEventStore() base.EventStore {

	return nil
}

func (s *PostgresFactory) BuildEventStoreFlags() c.Applicant {
	return func(b *c.Command) {
		b.Apply(
			c.Flag("evt-postgres-url", c.Str("host=localhost port=5432 sslmode=disable"),
				c.Description("Database URL for the Postgres EventStore"),
				c.Persistent(),
				c.EnvName("evt_postgres_url")),
		)
	}
}

func (s *PostgresFactory) CreateSequenceStore() base.SequenceStore {
	return nil
}

func (s *PostgresFactory) BuildSequenceStoreFlags() c.Applicant {
	return func(b *c.Command) {
		b.Apply(
			c.Flag("seq-postgres-url", c.Str("host=localhost port=5432 sslmode=disable"),
				c.Description("Database URL for the Postgres SequenceStore"),
				c.Persistent(),
				c.EnvName("seq_postgres_url")),
		)
	}
}

func init() {
	RegisterEventStore(&PostgresFactory{})
	RegisterSequenceStore(&PostgresFactory{})
}