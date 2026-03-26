package postgres

import (
	"context"

	"go.redsock.ru/mead/internal/clients/sqldb"
	"go.redsock.ru/mead/internal/domain"
	"go.redsock.ru/mead/internal/storage"
	"go.redsock.ru/mead/internal/storage/postgres/extra_proxies_queries"
)

type extraProxies struct {
	q *extra_proxies_queries.Queries
}

func NewExtraProxies(db sqldb.DB) storage.ExtraProxies {
	return &extraProxies{
		q: extra_proxies_queries.New(db),
	}
}

func (e *extraProxies) ListExtraProxies(ctx context.Context, req domain.ListExtraProxies) ([]string, error) {
	p, err := e.q.ListExtraProxies(ctx)
	if err != nil {
		return nil, wrapErr(err)
	}

	return p, nil
}
