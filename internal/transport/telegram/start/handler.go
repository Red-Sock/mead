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
		return out.SendMessage(response.NewMessage("Ботам не разрешено получать прокси"))
	}

	if in.From.ID != in.Chat.ID {
		return out.SendMessage(response.NewMessage("Разрешены только личные сообщения боту. Групповые чаты не поддерживаются"))
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
				contactButton := tgbotapi.NewKeyboardButtonContact("Поделиться контактом 📇")
				contactBtnWrapper := keyboard.Button{
					InternalButton: &contactButton,
				}
				kb.AddButton(contactBtnWrapper)
				kb.SetIsReplyKeyboard(true)

				msg := response.NewMessage("Добро пожаловать! Чтобы воспользоваться сервисом, вам необходимо зарегистрироваться.\n\n" +
					"Нажимая «Поделиться контактом», вы соглашаетесь на отслеживание ваших данных сервисом. " +
					"Мы не логируем содержание сообщений, но мы фиксируем, кто, откуда и куда отправляет сообщения " +
					"на случай запроса от правоохранительных органов. Не делайте ничего плохо используя наш прокси. Спасибо!")
				msg.Keys = &kb

				return out.SendMessage(msg)
			}
			return out.SendMessage(response.NewMessage(err.Error()))
		}
	}

	preText := "Вот ваша персональная ссылка на прокси. Нажмите на нее для настройки\n"
	msg := response.NewMessage(preText + auth.ProxyLink)
	return out.SendMessage(msg)
}

func (h *Handler) GetCommand() string {
	return Command
}
