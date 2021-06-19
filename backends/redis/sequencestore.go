package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type SequenceStore struct {
	rdb *redis.Client
}

func NewSequenceStore(rdb *redis.Client) *SequenceStore {
	return &SequenceStore{
		rdb: rdb,
	}
}

func (s *SequenceStore) Get(persistentClientID string) (int64, error) {
	key := "sub//" + persistentClientID + "//last-ack-sequence"
	res := s.rdb.Get(context.Background(), key)
	if res.Err() == redis.Nil {
		return 0, nil
	}
	return res.Int64()
}

func (s *SequenceStore) Store(persistentClientID string, sequence int64) error {
	key := "sub//" + persistentClientID + "//last-ack-sequence"
	return s.rdb.Set(context.Background(), key, sequence, 0).Err()
}
