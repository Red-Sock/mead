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
	resp, err := u.q.GetByTelegramName(ctx, username)
	if err != nil {
		return domain.User{}, wrapErr(err)
	}

	return domain.User{
		Username:   resp.Username,
		TelegramId: resp.TelegramID.Int64,
	}, nil
}

func (u *user) GetByTelegramId(ctx context.Context, id int64) (domain.User, error) {
	resp, err := u.q.GetByTelegramId(ctx, sql.NullInt64{Int64: id, Valid: true})
	if err != nil {
		return domain.User{}, wrapErr(err)
	}

	return domain.User{
		Username:   resp.Username,
		TelegramId: resp.TelegramID.Int64,
	}, nil
}

func (u *user) List(ctx context.Context, req domain.ListUsersReq) ([]domain.User, error) {
	q := sq.Select("username", "telegram_id").
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
		var telegramId sql.NullInt64
		err = rows.Scan(&u.Username, &telegramId)
		if err != nil {
			return nil, wrapErr(err)
		}
		u.TelegramId = telegramId.Int64

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

func (u *user) UpdateStatistics(ctx context.Context, arg user_queries.UpdateStatisticsParams) error {
	err := u.q.UpdateStatistics(ctx, arg)
	if err != nil {
		return wrapErr(err)
	}

	return nil
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
