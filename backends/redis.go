package backends

import (
	red "github.com/go-redis/redis/v8"
	c "github.com/mtrense/soil/config"
	"github.com/spf13/viper"
	"github.com/ticker-es/broker-go/backends/redis"
	es "github.com/ticker-es/client-go/eventstream/base"
)

type RedisFactory struct{}

func (s *RedisFactory) Names() []string {
	return []string{"redis"}
}

func (s *RedisFactory) CreateSequenceStore() es.SequenceStore {
	url := viper.GetString("seq_redis_url")
	db := viper.GetInt("seq_redis_db")
	password := viper.GetString("seq_redis_password")
	rdb := red.NewClient(&red.Options{
		Addr:     url,
		DB:       db,
		Password: password,
	})
	return redis.NewSequenceStore(rdb)
}

func (s *RedisFactory) BuildSequenceStoreFlags() c.Applicant {
	return func(b *c.Command) {
		b.Apply(
			c.Flag("seq-redis-url", c.Str("localhost:6379"),
				c.Description("URL for the Redis SequenceStore"),
				c.Persistent(),
				c.EnvName("seq_redis_url")),
			c.Flag("seq-redis-db", c.Int(0),
				c.Description("DB to connect"),
				c.Persistent(),
				c.EnvName("seq_redis_db")),
			c.Flag("seq-redis-password", c.Str(""),
				c.Description("Password for connection"),
				c.Persistent(),
				c.EnvName("seq_redis_password")),
		)
	}
}

func init() {
	RegisterSequenceStore(&RedisFactory{})
}
