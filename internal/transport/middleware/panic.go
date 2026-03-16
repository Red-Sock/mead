package middleware

import (
	"context"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func PanicInterceptor() grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(
		func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
			defer func() {
				if r := recover(); r != nil {
					err = r.(error)
					logrus.WithError(err).Error("panic in grpc handler")
				}
			}()

			return handler(ctx, req)
		})
}
