package redis

import (
	"context"
	"fmt"
	"land/settings"

	"github.com/redis/go-redis/v9"
)

var (
	client *redis.Client
	Nil    = redis.Nil
)

func Init(cfg *settings.RedisConfig) (err error) {
	client = redis.NewClient(
		&redis.Options{
			Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Password:     cfg.PassWord,
			DB:           cfg.DB, // use default DB
			PoolSize:     cfg.PoolSize,
			MinIdleConns: cfg.MinIdleConns,
		},
	)

	// context.Background()为本地测试使用，实际使用可能需要替换
	_, err = client.Ping(context.Background()).Result()
	if err != nil {
		return
	}
	return
}

func Close() {
	if client == nil {
		return
	}
	_ = client.Close()
}
