package mead_api_impl

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"go.redsock.ru/mead/internal/config"
	"go.redsock.ru/mead/pkg/mead_api"
)

type Impl struct {
	mead_api.UnimplementedMeadAPIServer

	version string
}

func New(cfg config.Config) *Impl {
	return &Impl{
		version: cfg.AppInfo.Version,
	}
}

func (impl *Impl) Register(server grpc.ServiceRegistrar) {
	mead_api.RegisterMeadAPIServer(server, impl)
}

func (impl *Impl) Gateway(ctx context.Context, endpoint string, opts ...grpc.DialOption) (route string, handler http.Handler) {
	gwHttpMux := runtime.NewServeMux()

	err := mead_api.RegisterMeadAPIHandlerFromEndpoint(
		ctx,
		gwHttpMux,
		endpoint,
		opts,
	)
	if err != nil {
		log.Error().Err(err).Msg("error registering grpc2http handler")
	}

	return "/api/", gwHttpMux
}
