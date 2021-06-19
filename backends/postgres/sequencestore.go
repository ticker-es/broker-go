package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	es "github.com/ticker-es/client-go/eventstream/base"
)

type SequenceStore struct {
	db *pgxpool.Pool
}

func NewSequenceStore(db *pgxpool.Pool) es.SequenceStore {
	return &SequenceStore{
		db: db,
	}
}

func (s *SequenceStore) Get(persistentClientID string) (int64, error) {
	var sequence int64
	ctx := context.Background()
	err := s.db.AcquireFunc(ctx, func(c *pgxpool.Conn) error {
		return c.QueryRow(ctx, "SELECT last_acknowledged_sequence FROM subscribers WHERE id = $1", persistentClientID).Scan(&sequence)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	return sequence, err
}

func (s *SequenceStore) Store(persistentClientID string, sequence int64) error {
	ctx := context.Background()
	return s.db.AcquireFunc(ctx, func(c *pgxpool.Conn) error {
		_, err := c.Exec(ctx, "INSERT INTO subscribers (id, last_acknowledged_sequence) VALUES ($1, $2) ON CONFLICT ON CONSTRAINT subscribers_pkey DO UPDATE SET last_acknowledged_sequence = $2", persistentClientID, sequence)
		return err
	})
}
