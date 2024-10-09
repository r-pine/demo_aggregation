package service

import (
	"context"
	"github.com/r-pine/demo_aggregation/app/internal/storage"
)

type Service struct {
	ctx context.Context
	st  *storage.Storage
}

func NewService(ctx context.Context, st *storage.Storage) *Service {
	return &Service{
		st: st,
	}
}
