package iservice

import (
	"context"

	"go.redsock.ru/mead/internal/domain"
)

type Service interface {
	Auth() Auth
}

type Auth interface {
	Authenticate(ctx context.Context, user, password string) error
	Add(ctx context.Context, username, password string) error

	ListUsers(ctx context.Context, req domain.ListUsersReq) ([]domain.User, error)
	AuthByTelegramUsername(ctx context.Context, username string) (domain.UserAuth, error)
}
