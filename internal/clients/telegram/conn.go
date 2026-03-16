package telegram

import (
	"github.com/Red-Sock/go_tg"
	errors "github.com/Red-Sock/trace-errors"
	"go.vervstack.ru/matreshka/pkg/matreshka/resources"
)

func New(cfg *resources.Telegram) (*go_tg.Bot, error) {
	bot, err := go_tg.NewBot(cfg.ApiKey)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	return bot, nil
}
