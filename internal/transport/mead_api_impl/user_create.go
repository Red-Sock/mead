package mead_api_impl

import (
	"context"

	pb "go.redsock.ru/mead/pkg/mead_api"
)

func (impl *Impl) CreateUser(ctx context.Context, request *pb.CreateUser_Request) (*pb.CreateUser_Response, error) {
	err := impl.authService.Add(ctx, request.Username, request.Password)
	if err != nil {
		return nil, err
	}

	return &pb.CreateUser_Response{}, nil
}
