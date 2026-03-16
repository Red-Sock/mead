package service

import (
	"go.redsock.ru/mead/internal/config"
	"go.redsock.ru/mead/internal/service/auth"
	"go.redsock.ru/mead/internal/service/iservice"
	"go.redsock.ru/mead/internal/storage"
)

type Service struct {
	auth iservice.Auth
}

func New(cfg config.Config, str storage.Storage) iservice.Service {
	return &Service{
		auth: auth.NewAuth(cfg, str),
	}
}

func (s *Service) Auth() iservice.Auth {
	return s.auth
}
