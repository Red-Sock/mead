package sqlite

import (
	"context"

	"go.redsock.ru/mead/internal/clients/sqldb"
	"go.redsock.ru/mead/internal/storage"
	"go.redsock.ru/mead/internal/storage/sqlite/user_queries"
)

type user struct {
	db sqldb.DB

	q user_queries.Querier
}

func NewUser(db sqldb.DB) storage.Users {
	return &user{
		q: user_queries.New(db),
	}
}

func (u *user) Add(ctx context.Context, arg user_queries.AddParams) error {
	err := u.q.Add(ctx, arg)
	if err != nil {
		return wrapErr(err)
	}

	return nil
}

func wrapErr(err error) error {
	return err
}
