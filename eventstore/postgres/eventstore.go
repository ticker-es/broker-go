package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/lib/pq"

	es "github.com/ticker-es/client-go/eventstream/base"
)

type EventStore struct {
	db *pgxpool.Pool
}

func NewEventStore(db *pgxpool.Pool) es.EventStore {
	return &EventStore{
		db: db,
	}
}

func (s *EventStore) Store(event *es.Event) (int64, error) {
	var sequence int64
	ctx := context.Background()
	err := s.db.AcquireFunc(ctx, func(c *pgxpool.Conn) error {
		row := c.QueryRow(ctx, "INSERT INTO event_streams (aggregate, type, occurred_at, revision, payload) VALUES ($1, $2, $3, $4, $5) RETURNING sequence",
			event.Aggregate,
			event.Type,
			event.OccurredAt,
			1,
			event.Payload,
		)
		if err := row.Scan(&sequence); err != nil {
			return err
		}
		return nil
	})
	return sequence, err
}

func (s *EventStore) LastKnownSequence() int64 {
	var (
		sequence int64
		isCalled bool
	)
	ctx := context.Background()
	s.db.AcquireFunc(ctx, func(c *pgxpool.Conn) error {
		return c.QueryRow(context.Background(), "SELECT last_value, is_called FROM event_streams_sequence_seq").Scan(&sequence, &isCalled)
	})
	if isCalled {
		return sequence
	} else {
		return 0
	}
}

func (s *EventStore) Read(sequence int64) (*es.Event, error) {
	var result es.Event
	ctx := context.Background()
	return &result, s.db.AcquireFunc(ctx, func(c *pgxpool.Conn) error {
		return c.QueryRow(context.Background(), "SELECT sequence, aggregate, type, occurred_az, payload FROM event_streams WHERE sequence = $1", sequence).
			Scan(&result.Sequence, &result.Aggregate, &result.Type, &result.OccurredAt, &result.Payload)
	})
}

func (s *EventStore) ReadAll(ctx context.Context, sel es.Selector, bracket es.Bracket, handler es.EventHandler) error {
	return s.db.AcquireFunc(ctx, func(c *pgxpool.Conn) error {
		rows, err := c.Query(ctx, "SELECT sequence, aggregate, type, occurred_at, payload FROM event_streams ORDER BY sequence")
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var event es.Event
			if err := rows.Scan(&event.Sequence, pq.Array(&event.Aggregate), &event.Type, &event.OccurredAt, &event.Payload); err != nil {
				return err
			}
			if err := handler(&event); err != nil {
				return err
			}
		}
		return nil
	})
}

type dbEvent struct {
	Sequence   int64                  `json:"sequence,omitempty" yaml:"sequence,omitempty"`
	Aggregate  []string               `json:"aggregate,omitempty" yaml:"aggregate,omitempty"`
	Type       string                 `json:"type,omitempty" yaml:"type,omitempty"`
	OccurredAt time.Time              `json:"occurred_at,omitempty" yaml:"occurred_at,omitempty"`
	Payload    map[string]interface{} `json:"payload,omitempty" yaml:"payload,omitempty"`
}
