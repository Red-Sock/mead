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
	Add(ctx context.Context, username string) error

	ListUsers(ctx context.Context, req domain.ListUsersReq) ([]domain.User, error)
	AuthByTelegramUsername(ctx context.Context, username string) (domain.UserAuth, error)
	AuthByTelegramId(ctx context.Context, id int64) (domain.UserAuth, error)

	Register(ctx context.Context, username string, telegramId int64) (domain.UserAuth, error)
	UpdateStatistics(ctx context.Context, username string, bytesPassed int64) error
}

type Tg interface {
}
