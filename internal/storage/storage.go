package storage

import (
	"context"

	"go.redsock.ru/mead/internal/domain"
	"go.redsock.ru/mead/internal/storage/postgres/user_queries"
)

type Storage interface {
	Users() Users
}

type Users interface {
	Add(ctx context.Context, arg user_queries.AddParams) error

	GetByTelegramName(ctx context.Context, username string) (domain.User, error)
	GetByTelegramId(ctx context.Context, id int64) (domain.User, error)
	GetPassByUsername(ctx context.Context, username string) (string, error)

	List(ctx context.Context, req domain.ListUsersReq) ([]domain.User, error)
	UpdateStatistics(ctx context.Context, arg user_queries.UpdateStatisticsParams) error
	ListStatistics(ctx context.Context) ([]domain.UserStatistic, error)
}
