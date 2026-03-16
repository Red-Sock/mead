package middleware

import (
	"context"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

func PanicInterceptor() grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(
		func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
			defer func() {
				if r := recover(); r != nil {
					log.Error().
						Any("recover", r).
						Msg("panic in grpc handler")
				}
			}()

			return handler(ctx, req)
		})
}
