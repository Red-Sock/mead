package version

import (
	tgapi "github.com/Red-Sock/go_tg/interfaces"
	"github.com/Red-Sock/go_tg/model"
	"github.com/Red-Sock/go_tg/model/response"

	"go.redsock.ru/mead/internal/config"
)

const Command = "/version"

type Handler struct {
	version string
}

func New(cfg config.Config) *Handler {
	return &Handler{
		version: cfg.AppInfo.Version,
	}
}

func (h *Handler) Handle(in *model.MessageIn, out tgapi.Chat) error {
	return out.SendMessage(response.NewMessage(in.Text + ": " + h.version))
}

func (h *Handler) GetCommand() string {
	return Command
}
