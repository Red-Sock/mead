package iservice

import (
	"context"
)

type Service interface {
	Auth() Auth
}

type Auth interface {
	Authenticate(ctx context.Context, user, password string) error
	Add(ctx context.Context, username, password string) error
}
