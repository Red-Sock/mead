package start

import (
	tgapi "github.com/Red-Sock/go_tg/interfaces"
	"github.com/Red-Sock/go_tg/model"
	"github.com/Red-Sock/go_tg/model/keyboard"
	"github.com/Red-Sock/go_tg/model/response"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

	auth, err := h.authService.AuthByTelegramId(in.Ctx, in.From.ID)
	if err != nil {
		if !rerrors.Is(err, user_errors.ErrNotFound) {
			return out.SendMessage(response.NewMessage(err.Error()))
		}

		auth, err = h.authService.AuthByTelegramUsername(in.Ctx, in.From.UserName)
		if err != nil {
			if rerrors.Is(err, user_errors.ErrNotFound) {
				kb := keyboard.GridKeyboard{
					Columns: 1,
					Rows:    1,
				}
				contactButton := tgbotapi.NewKeyboardButtonContact("Share contact 📇")
				contactBtnWrapper := keyboard.Button{
					InternalButton: &contactButton,
				}
				kb.AddButton(contactBtnWrapper)
				kb.SetIsReplyKeyboard(true)

				msg := response.NewMessage("Welcome! To use this service, you need to register.\n\n" +
					"By clicking \"Share contact\", you consent to the service tracking your data. " +
					"We do not log message content, but we log who, from where, and where to send messages " +
					"in case authorities ask.")
				msg.Keys = &kb

				return out.SendMessage(msg)
			}
			return out.SendMessage(response.NewMessage(err.Error()))
		}
	}

	preText := "Here is your personal proxy link. Click it to setup\n"
	msg := response.NewMessage(preText + auth.ProxyLink)
	return out.SendMessage(msg)
}

func (h *Handler) GetCommand() string {
	return Command
}
