package start

import (
	"strings"

	tgapi "github.com/Red-Sock/go_tg/interfaces"
	"github.com/Red-Sock/go_tg/model"
	"github.com/Red-Sock/go_tg/model/response"
	"go.redsock.ru/rerrors"

	"go.redsock.ru/mead/internal/service/iservice"
	"go.redsock.ru/mead/internal/service/user_errors"
	"go.redsock.ru/mead/internal/transport/telegram/messages"
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
		return out.SendMessage(response.NewMessage("Ботам не разрешено получать прокси"))
	}

	if in.From.ID != in.Chat.ID {
		return out.SendMessage(response.NewMessage("Разрешены только личные сообщения боту. Групповые чаты не поддерживаются"))
	}

	auth, err := h.authService.AuthByTelegramUsername(in.Ctx, in.From.UserName)
	if err != nil {
		if !rerrors.Is(err, user_errors.ErrNotFound) {
			err = out.SendMessage(response.NewMessage(err.Error()))
			if err != nil {
				return rerrors.Wrap(err, "")
			}
		}

		msg := messages.OfertaMessage()
		err = out.SendMessage(msg)
		if err != nil {
			return rerrors.Wrap(err, "")
		}

		auth, err = h.authService.Add(in.Ctx, in.From.UserName)
		if err != nil {
			return rerrors.Wrap(err, "")
		}
	}

	msg := messages.ProxyLinkMessage(auth.ProxyLink)

	err = out.SendMessage(msg)
	if err != nil {
		return rerrors.Wrap(err, "")
	}

	p, err := h.authService.ReserveProxies(in.Ctx)
	if err != nil {
		return rerrors.Wrap(err, "")
	}

	if len(p) == 0 {
		return nil
	}

	extra := "Так же рекомендую добавить запасной(ые) прокси:\n" + strings.Join(p, "\n\n")

	msg = response.NewMessage(extra)
	return out.SendMessage(msg)
}

func (h *Handler) GetCommand() string {
	return Command
}
