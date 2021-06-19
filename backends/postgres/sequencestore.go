package postgres

import (
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
	panic("implement me")
}

func (s *SequenceStore) Store(persistentClientID string, sequence int64) error {
	panic("implement me")
}
