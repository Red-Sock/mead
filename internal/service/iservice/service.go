package iservice

import (
	"context"

	"go.redsock.ru/mead/internal/domain"
)

type Service interface {
	Auth() Auth
	Tg() Tg
}

type Auth interface {
	Authenticate(ctx context.Context, user, password string) error
	Add(ctx context.Context, username string) (domain.UserAuth, error)

	ListUsers(ctx context.Context, req domain.ListUsersReq) ([]domain.User, error)
	AuthByTelegramUsername(ctx context.Context, username string) (domain.UserAuth, error)

	UpdateStatistics(ctx context.Context, username string, bytesPassed int64) error
	ListStatistics(ctx context.Context) ([]domain.UserStatistic, error)
}

type Tg interface {
}
