package start

import (
	tgapi "github.com/Red-Sock/go_tg/interfaces"
	"github.com/Red-Sock/go_tg/model"
	"github.com/Red-Sock/go_tg/model/response"
	"go.redsock.ru/rerrors"

	"go.redsock.ru/mead/internal/service/iservice"
	"go.redsock.ru/mead/internal/service/user_errors"
)

const Command = "/start"

type Handler struct {
	authService iservice.Auth
}

func New(service iservice.Service) *Handler {
	return &Handler{
		authService: service.Auth(),
	}
}

func (h *Handler) Handle(in *model.MessageIn, out tgapi.Chat) error {
	if in.From.IsBot {
		return out.SendMessage(response.NewMessage("Bot's aren't allowed to get proxy"))
	}

	if in.From.ID != in.Chat.ID {
		return out.SendMessage(response.NewMessage("Only direct messages to bot is allowed. No group chatting"))
	}

	auth, err := h.authService.AuthByTelegramUsername(in.Ctx, in.From.UserName)
	if err != nil {
		if rerrors.Is(err, user_errors.ErrNotFound) {
			return out.SendMessage(response.NewMessage("You are not in white list. Ask Administrator to add you"))
		}
		return out.SendMessage(response.NewMessage(err.Error()))
	}

	preText := "Here is your personal proxy link. Click it to setup\n"
	msg := response.NewMessage(preText + auth.ProxyLink)
	return out.SendMessage(msg)
}

func (h *Handler) GetCommand() string {
	return Command
}
