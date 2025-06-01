package kvs

import (
	"context"
	"fmt"

	"github.com/dev-pt-bai/cataloging/configs"
	"github.com/redis/go-redis/v9"
)

func NewClient(config *configs.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: config.Database.Redis.Addr,
	})

	res, err := rdb.Ping(context.Background()).Result()
	if res != "PONG" || err != nil {
		return nil, fmt.Errorf("failed to verify connection to redis database: response: %s, cause: %w", res, err)
	}

	return rdb, nil
}
