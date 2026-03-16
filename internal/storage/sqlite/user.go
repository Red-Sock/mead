package sqlite

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"
	"modernc.org/sqlite"

	"go.redsock.ru/mead/internal/clients/sqldb"
	"go.redsock.ru/mead/internal/domain"
	"go.redsock.ru/mead/internal/service/user_errors"
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

func (u *user) GetByTelegramName(ctx context.Context, username string) (domain.User, error) {
	userNameDb, err := u.q.GetByTelegramName(ctx, username)
	if err != nil {
		return domain.User{}, wrapErr(err)
	}

	return domain.User{
		Username: userNameDb,
	}, nil
}

func (u *user) List(ctx context.Context, req domain.ListUsersReq) ([]domain.User, error) {
	q := sq.Select("username").
		From("users")

	query, args, err := q.ToSql()
	if err != nil {
		return nil, rerrors.Wrap(err, "list users")
	}

	rows, err := u.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, wrapErr(err)
	}
	defer wrapClose(rows)

	users := []domain.User{}

	for rows.Next() {
		u := domain.User{}
		err = rows.Scan(&u.Username)
		if err != nil {
			return nil, wrapErr(err)
		}

		users = append(users, u)
	}

	return users, nil

}

func (u *user) GetPassByUsername(ctx context.Context, username string) (string, error) {
	pass, err := u.q.GetPassByUsername(ctx, username)
	if err != nil {
		return "", wrapErr(err)
	}

	return pass, nil
}

func wrapErr(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return user_errors.ErrNotFound
	}

	var e *sqlite.Error
	if !errors.As(err, &e) {
		return err
	}

	switch e.Code() {
	case 1555:
		return user_errors.ErrAlreadyExists
	}

	return err
}

func wrapClose(rows *sql.Rows) {
	err := rows.Close()
	if err != nil {
		log.Error().
			Err(err).
			Msg("close rows error")
	}
}
