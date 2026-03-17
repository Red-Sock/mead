package contact

import (
	tgapi "github.com/Red-Sock/go_tg/interfaces"
	"github.com/Red-Sock/go_tg/model"
	"github.com/Red-Sock/go_tg/model/response"
	"github.com/rs/zerolog/log"

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
	if in.Contact != nil {
		return h.HandleContactSharing(in, out)
	}

	return out.SendMessage(response.NewMessage("Can't handle that"))

}

func (h *Handler) HandleContactSharing(in *model.MessageIn, out tgapi.Chat) error {
	log.Info().
		Int64("user_id", in.From.ID).
		Str("username", in.From.UserName).
		Str("phone_number", in.Contact.PhoneNumber).
		Msg("User shared contact")

	auth, err := h.authService.Register(in.Ctx, in.From.UserName, in.From.ID)
	if err != nil {
		return out.SendMessage(response.NewMessage("Failed to register: " + err.Error()))
	}

	preText := "Registration successful! Here is your personal proxy link. Click it to setup\n"
	msg := response.NewMessage(preText + auth.ProxyLink)

	msg.RemoveKeyboard = true

	return out.SendMessage(msg)
}
