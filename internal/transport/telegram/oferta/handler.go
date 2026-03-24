package oferta

import (
	tgapi "github.com/Red-Sock/go_tg/interfaces"
	"github.com/Red-Sock/go_tg/model"

	"go.redsock.ru/mead/internal/transport/telegram/messages"
)

const Command = "/oferta"

type Handler struct {
}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) Handle(in *model.MessageIn, out tgapi.Chat) error {
	return out.SendMessage(messages.OfertaMessage())
}

func (h *Handler) GetCommand() string {
	return Command
}
