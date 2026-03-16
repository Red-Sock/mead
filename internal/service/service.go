package service

import (
	"github.com/Red-Sock/go_tg"

	"go.redsock.ru/mead/internal/config"
	"go.redsock.ru/mead/internal/service/auth"
	"go.redsock.ru/mead/internal/service/iservice"
	"go.redsock.ru/mead/internal/service/tg"
	"go.redsock.ru/mead/internal/storage"
)

type Service struct {
	auth iservice.Auth
	tg   iservice.Tg
}

func New(cfg config.Config, str storage.Storage, bot *go_tg.Bot) iservice.Service {
	return &Service{
		auth: auth.NewAuth(cfg, str),
		tg:   tg.New(bot),
	}
}

func (s *Service) Auth() iservice.Auth {
	return s.auth
}

func (s *Service) Tg() iservice.Tg {
	return s.tg
}
