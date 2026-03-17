package telegram

import (
	client "github.com/Red-Sock/go_tg"

	"go.redsock.ru/mead/internal/config"
	"go.redsock.ru/mead/internal/service/iservice"
	"go.redsock.ru/mead/internal/transport/telegram/contact"
	"go.redsock.ru/mead/internal/transport/telegram/start"
	"go.redsock.ru/mead/internal/transport/telegram/version"
)

type Server struct {
	bot *client.Bot
}

func NewServer(cfg config.Config, bot *client.Bot, srv iservice.Service) (s *Server) {
	s = &Server{
		bot: bot,
	}

	{
		// Add handlers here
		s.bot.MustAddCommandHandler(version.New(cfg))
		s.bot.MustAddCommandHandler(start.New(srv))

		s.bot.SetDefaultCommandHandler(contact.New(srv))
	}

	return s
}

func (s *Server) Start() error {
	return s.bot.Start()
}

func (s *Server) Stop() error {
	s.bot.Stop()
	return nil
}
