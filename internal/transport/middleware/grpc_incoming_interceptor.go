package middleware

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthHeader - header containing auth token to talk to service
// in order to perform REST call should pre append runtime.MetadataHeaderPrefix (e.g - "Grpc-Metadata-")
const AuthHeader = "Authorization"

func GrpcIncomingInterceptor(keyValidator func(key string) bool) grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(
		func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return nil, status.Error(codes.FailedPrecondition, "error unmarshalling metadata from context")
			}

			auth := md.Get(AuthHeader)
			if len(auth) == 0 {
				return nil, status.Error(codes.PermissionDenied, "no auth header")
			}

			if !keyValidator(auth[0]) {
				return nil, status.Error(codes.PermissionDenied, "invalid auth header")
			}

			return handler(ctx, req)
		})
}
