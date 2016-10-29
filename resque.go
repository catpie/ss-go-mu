package main

import (
	"github.com/kavu/go-resque"         // Import this package
	_ "github.com/kavu/go-resque/godis" // Use godis driver
	"github.com/simonz05/godis/redis"   // Redis client from godis package
)

var (
	enqueuer *resque.RedisEnqueuer
)

func InitQueue() error {
	var err error

	client := redis.New("tcp:"+config.Redis.Host, int(config.Redis.Db), config.Redis.Pass) // Create new Redis client to use for enqueuing
	enqueuer = resque.NewRedisEnqueuer("godis", client)                                    // Create enqueuer instance

	return err
}
