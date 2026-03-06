package main

import (
	"github.com/rs/zerolog/log"

	"go.redsock.ru/mead/internal/app"
)

func main() {
	a, err := app.New()
	if err != nil {
		log.Fatal().Err(err).Msg("app initialization failed")
	}

	err = a.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("app start failed")
	}
}
