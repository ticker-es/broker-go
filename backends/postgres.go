package backends

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"

	c "github.com/mtrense/soil/config"
	"github.com/spf13/viper"
	"github.com/ticker-es/broker-go/backends/postgres"
	"github.com/ticker-es/client-go/eventstream/base"
)

type PostgresFactory struct{}

func (s *PostgresFactory) Names() []string {
	return []string{"postgres", "pg"}
}

func (s *PostgresFactory) CreateEventStore() base.EventStore {
	url := viper.GetString("evt_postgres_url")
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		panic(err)
	}
	db, err := pgxpool.ConnectConfig(context.Background(), cfg)
	//db, err := sqlx.Connect("pgx", url)
	if err != nil {
		panic(err)
	}
	return postgres.NewEventStore(db)
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
	url := viper.GetString("seq_postgres_url")
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		panic(err)
	}
	db, err := pgxpool.ConnectConfig(context.Background(), cfg)
	//db, err := sqlx.Connect("pgx", url)
	if err != nil {
		panic(err)
	}
	return postgres.NewSequenceStore(db)
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
