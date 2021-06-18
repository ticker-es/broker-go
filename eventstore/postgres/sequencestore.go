package postgres

import (
	"github.com/jmoiron/sqlx"
	es "github.com/ticker-es/client-go/eventstream/base"
)

type SequenceStore struct {
	db *sqlx.DB
}

func (s *SequenceStore) Get(persistentClientID string) (int64, error) {
	panic("implement me")
}

func (s *SequenceStore) Store(persistentClientID string, sequence int64) error {
	panic("implement me")
}

func NewSequenceStore() es.SequenceStore {
	return &SequenceStore{}
}
