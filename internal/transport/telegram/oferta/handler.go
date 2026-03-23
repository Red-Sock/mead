package oferta

import (
	tgapi "github.com/Red-Sock/go_tg/interfaces"
	"github.com/Red-Sock/go_tg/model"
	"github.com/Red-Sock/go_tg/model/response"

	"go.redsock.ru/mead/internal/domain"
)

const Command = "/oferta"

type Handler struct {
}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) Handle(in *model.MessageIn, out tgapi.Chat) error {
	return out.SendMessage(response.NewMessage(domain.OfertaText))
}

func (h *Handler) GetCommand() string {
	return Command
}
