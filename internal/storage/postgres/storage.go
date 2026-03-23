package postgres

import (
	"go.redsock.ru/mead/internal/clients/sqldb"
	"go.redsock.ru/mead/internal/storage"
)

type dataStorage struct {
	user storage.Users
}

func New(db sqldb.DB) storage.Storage {
	return &dataStorage{
		user: NewUser(db),
	}
}

func (s *dataStorage) Users() storage.Users {
	return s.user
}
