package middleware

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	Header = "Authorization"

	passAuth = "Pass"
)

func AuthInterceptor(pass string) grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(
		func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return nil, status.Error(codes.FailedPrecondition, "error unmarshalling metadata from context")
			}

			auth := md.Get(Header)
			if len(auth) == 0 {
				return nil, status.Error(codes.PermissionDenied, "no auth header")
			}

			auth = strings.Split(auth[0], " ")

			if len(auth) != 2 {
				return nil, status.Error(codes.InvalidArgument,
					"in auth header expected to have auth-type and auth-value (key, pass, etc.)")
			}

			switch auth[0] {
			case passAuth:
				if auth[1] != pass {
					return nil, status.Error(codes.PermissionDenied, "invalid auth header")
				}
			default:
				return nil, status.Error(codes.PermissionDenied, "invalid auth-type")
			}

			return handler(ctx, req)
		})
}
