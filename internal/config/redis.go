package config

import (
	"context"
	"github.com/redis/go-redis/v9"
	log "github.com/Thanga-tamil/logger_lib"
)

var Redis *redis.Client

func NewRedis() error {

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Protocol: 2,
	})

	ctx := context.Background()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Error("Error while establishing redis connection:", err.Error())
		return err
	}

	Redis = rdb; return nil
}
