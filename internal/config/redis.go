package config

import (
	"context"

	log "github.com/Thanga-tamil/logger_v2"

	"github.com/redis/go-redis/v9"
)

func NewRedis(addr string) (*redis.Client, error) {

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password
		DB:       0,  // use default DB
		Protocol: 2,
	})

	ctx := context.Background()

	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	log.Infox("Redis connection established successfully ::", "PING", pong)

	return rdb, err
}
