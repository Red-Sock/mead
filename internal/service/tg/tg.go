package tg

import (
	"github.com/Red-Sock/go_tg"

	"go.redsock.ru/mead/internal/service/iservice"
)

type Service struct {
	bot *go_tg.Bot
}

func New(bot *go_tg.Bot) iservice.Tg {
	return &Service{
		bot: bot,
	}
}
