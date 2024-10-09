package storage

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type Storage struct {
	ctx context.Context
	rc  *redis.Client
}

func NewStorage(
	ctx context.Context,
	rc *redis.Client,
) *Storage {
	return &Storage{
		ctx: ctx,
		rc:  rc,
	}
}
