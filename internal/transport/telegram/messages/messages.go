package messages

import (
	"github.com/Red-Sock/go_tg/model/response"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const ofertaText = `Привет! Это Прокси для восстановления работы Телеграм.
Пользуясь сервисом, вы автоматически соглашаетесь со следующими условиями:

Сервис логирует данные о подключении в формате

<ip откуда идёт запрос>-<login>-<ip куда идёт запрос>

Мы не видим ваших сообщений и не логируем их. Только факты подключения к нашему прокси на случай обращения
правоохранительных органов
Пока сервис в Бэта тестировании - он бесплатный. В будущем это может изменится. Вы соглашаетесь с данной политикой

Не делайте ничего плохо используя наш прокси. Спасибо! ☺️
`

func OfertaMessage() *response.MessageOut {
	msg := response.NewMessage(ofertaText)
	msg.Entities = append(msg.Entities, tgbotapi.MessageEntity{
		Type:   "italic",
		Offset: 179,
		Length: 53,
	})

	return msg
}

func ProxyLinkMessage(link string) *response.MessageOut {
	msg := response.NewMessage(
		`Вот ваша персональная ссылка на прокси. Нажмите на нее для настройки
` + link +
			`
Не делитесь ей ни с кем!`)
	msg.Entities = append(msg.Entities, tgbotapi.MessageEntity{
		Type:   "text_link",
		URL:    link,
		Offset: 69,
		Length: len(link),
	})

	return msg
}
