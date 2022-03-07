package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/lib/pq"

	es "github.com/ticker-es/client-go/eventstream/base"
)

const (
	StateAlive = iota
	StateDead
)

var (
	ErrAggregateIsDead = errors.New("aggregate is dead")
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
	if state, err := s.getAggregateState(event.Aggregate); err != nil {
		return 0, err
	} else {
		if state == StateDead {
			return 0, ErrAggregateIsDead
		}
	}

	switch event.Type {
	case "$tombstone":
		err := s.storeAggregateState(event.Aggregate, StateDead)
		if err != nil {
			return 0, err
		}
		return s.storeEvent(event)
	default:
		return s.storeEvent(event)
	}
}

func (s *EventStore) storeAggregateState(aggregate []string, state int) error {
	ctx := context.Background()

	err := s.db.AcquireFunc(ctx, func(c *pgxpool.Conn) error {
		_, err := c.Exec(ctx, "INSERT INTO aggregate_states (aggregate, state) VALUES ($1, $2)",
			aggregate, state,
		)
		return err
	})

	return err
}

func (s *EventStore) getAggregateState(aggregate []string) (int, error) {
	ctx := context.Background()

	var state int
	err := s.db.AcquireFunc(ctx, func(c *pgxpool.Conn) error {
		row := c.QueryRow(ctx, "SELECT state FROM aggregate_states WHERE aggregate = $1", aggregate)

		err := row.Scan(&state)

		if err == nil {
			return nil
			// Dirty hack. I haven't found yet a way to handle pg errors correctly
		} else if err.Error() == "no rows in result set" {
			state = StateAlive
			return nil
		} else {
			return err
		}
	})

	return state, err
}

func (s *EventStore) storeEvent(event *es.Event) (int64, error) {
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
	var arguments []interface{}
	var predicates []string
	for pos, agg := range sel.Aggregate {
		if agg != "" {
			arguments = append(arguments, agg)
			predicates = append(predicates, fmt.Sprintf("e.aggregate[%d] = $%d", pos, len(arguments)))
		}
	}
	if sel.Type != "" {
		arguments = append(arguments, sel.Type)
		predicates = append(predicates, fmt.Sprintf("e.type = $%d", len(arguments)))
	}
	arguments = append(arguments, bracket.NextSequence, bracket.LastSequence)
	predicates = append(predicates, fmt.Sprintf("e.sequence BETWEEN $%d AND $%d", len(arguments)-1, len(arguments)))

	query := `
	SELECT e.sequence, e.aggregate, e.type, e.occurred_at, e.payload
	FROM event_streams e
	LEFT JOIN aggregate_states a ON (a.aggregate = e.aggregate)
	WHERE
	`
	query = query + "(" + strings.Join(predicates, " OR ") + ") "

	query = query + "AND ((coalesce(a.state,0) = " + fmt.Sprintf("%v", StateDead) + " AND e.type = '$tombstone') OR (coalesce(a.state,0) != " + fmt.Sprintf("%v", StateDead) + ")) "

	query = query + "ORDER BY e.sequence "

	return s.db.AcquireFunc(ctx, func(c *pgxpool.Conn) error {
		rows, err := c.Query(ctx, query, arguments...)
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
