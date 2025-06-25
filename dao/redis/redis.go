package redis

import (
	"context"
	"fmt"
	"land/settings"
	"time"

	"github.com/go-redis/redis/v8"
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

	fmt.Println("Redis init success")
	return
}

func Close() {
	if client == nil {
		return
	}
	_ = client.Close()
}

// SetJWTToken 将JWT存入Redis，过期时间为24小时
func SetJWTToken(userID int64, token string, expire time.Duration) error {
	ctx := context.Background()
	key := getRedisKey(KeyJWTTokenPF + fmt.Sprintf("%d", userID))
	return client.Set(ctx, key, token, expire).Err()
}

// GetJWTToken 获取Redis中存储的JWT
func GetJWTToken(userID int64) (string, error) {
	ctx := context.Background()
	key := getRedisKey(KeyJWTTokenPF + fmt.Sprintf("%d", userID))
	return client.Get(ctx, key).Result()
}

// DelJWTToken 删除Redis中的JWT
func DelJWTToken(userID int64) error {
	ctx := context.Background()
	key := getRedisKey(KeyJWTTokenPF + fmt.Sprintf("%d", userID))
	return client.Del(ctx, key).Err()
}
