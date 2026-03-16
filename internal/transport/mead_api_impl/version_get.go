package mead_api_impl

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	pb "go.redsock.ru/mead/pkg/mead_api"
)

func (impl *Impl) Version(ctx context.Context, request *pb.Version_Request) (*pb.Version_Response, error) {
	return &pb.Version_Response{
		Version:         impl.version,
		ClientTimestamp: timestamppb.Now(),
	}, nil
}
