package middleware

import (
	"context"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

func LogInterceptor() grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(
		func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {

			defer func() {
				log.Info().
					Str("method", info.FullMethod).
					Any("request", req).
					Any("response", resp).
					Err(err).
					Msg("GRPC request")
			}()

			resp, err = handler(ctx, req)

			return resp, err
		})
}
