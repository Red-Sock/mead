package storage

import (
	user_queries "go.redsock.ru/mead/internal/storage/sqlite/user_queries"
)

type Storage interface {
	Users() Users
}

type Users interface {
	user_queries.Querier
}
