package storage

import (
	"context"
	"time"

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

func (s *Storage) Set(key string, value any, expiresIn time.Duration) error {
	if err := s.rc.Set(s.ctx, key, value, expiresIn).Err(); err != nil {
		return err
	}
	return nil
}

func (s *Storage) Get(key string) (string, error) {
	data, err := s.rc.Get(s.ctx, key).Result()
	if err != nil {
		return "", err
	}
	return data, nil
}
