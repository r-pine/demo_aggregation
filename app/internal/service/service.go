package service

import (
	"context"
	"time"

	"github.com/r-pine/demo_aggregation/app/internal/storage"
)

type Service struct {
	ctx context.Context
	st  *storage.Storage
}

func NewService(
	ctx context.Context,
	st *storage.Storage,
) *Service {
	return &Service{
		ctx: ctx,
		st:  st,
	}
}

func (s *Service) Get(key string) (string, error) {
	value, err := s.st.Get(key)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (s *Service) Set(key string, value any) error {
	if err := s.st.Set(key, value, time.Duration(0)); err != nil {
		return err
	}
	return nil
}
