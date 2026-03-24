package contact

import (
	tgapi "github.com/Red-Sock/go_tg/interfaces"
	"github.com/Red-Sock/go_tg/model"
	"github.com/Red-Sock/go_tg/model/response"

	"go.redsock.ru/mead/internal/service/iservice"
)

type Handler struct {
	authService iservice.Auth
}

func New(service iservice.Service) *Handler {
	return &Handler{
		authService: service.Auth(),
	}
}

func (h *Handler) Handle(in *model.MessageIn, out tgapi.Chat) error {
	return out.SendMessage(response.NewMessage("Не могу это обработать"))
}
