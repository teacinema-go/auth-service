package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/teacinema-go/auth-service/internal/config"
)

func NewRedisClient(ctx context.Context, cfg *config.Redis) (*redis.Client, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Password,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		fmt.Printf("failed to connect to redis server: %s\n", err.Error())
		return nil, err
	}

	return rdb, nil
}
