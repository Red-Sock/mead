package mead_api_impl

import (
	"context"

	errors "github.com/Red-Sock/trace-errors"

	"go.redsock.ru/mead/internal/service/user_errors"
	pb "go.redsock.ru/mead/pkg/mead_api"
)

func (impl *Impl) CreateUser(ctx context.Context, request *pb.CreateUser_Request) (*pb.CreateUser_Response, error) {
	username := request.GetUsername()
	if username != "" {
		_, err := impl.authService.Add(ctx, username)
		if err != nil {
			return nil, err
		}
		return &pb.CreateUser_Response{}, nil

	}

	for _, username = range request.GetUsernames().GetUsernames() {
		_, err := impl.authService.Add(ctx, username)
		if err != nil {
			if errors.Is(err, user_errors.ErrAlreadyExists) {
				continue
			}

			return nil, err
		}
	}

	return &pb.CreateUser_Response{}, nil
}
