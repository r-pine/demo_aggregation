package redis

import (
	"context"
	"fmt"
	"github.com/r-pine/demo_aggregation/app/pkg/config"
	"github.com/r-pine/demo_aggregation/app/pkg/logging"

	"github.com/redis/go-redis/v9"
)

type RcClient struct {
	ctx context.Context
	cfg *config.Config
	log logging.Logger
}

func NewRedisClient(ctx context.Context, cfg *config.Config, log logging.Logger) *RcClient {
	return &RcClient{
		ctx: ctx,
		cfg: cfg,
		log: log,
	}
}

func (rc *RcClient) ConnectToRedis() (*redis.Client, error) {
	addr := fmt.Sprintf("%s:%s", rc.cfg.Redis.Host, rc.cfg.Redis.Port)
	client := redis.NewClient(&redis.Options{
		MaxIdleConns: 6,
		Addr:         addr,
		Password:     rc.cfg.Redis.Password,
		DB:           0,
	})

	_, err := client.Ping(rc.ctx).Result()
	if err != nil {
		return nil, err
	}

	return client, nil
}
