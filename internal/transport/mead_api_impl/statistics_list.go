package mead_api_impl

import (
	"context"

	pb "go.redsock.ru/mead/pkg/mead_api"
)

func (impl *Impl) ListStatistics(ctx context.Context, request *pb.ListStatistics_Request) (*pb.ListStatistics_Response, error) {
	stats, err := impl.authService.ListStatistics(ctx)
	if err != nil {
		return nil, err
	}

	resp := &pb.ListStatistics_Response{
		Stas: make([]*pb.UserStatistic, 0, len(stats)),
	}

	for _, s := range stats {
		resp.Stas = append(resp.Stas, &pb.UserStatistic{
			Username:    s.Username,
			LastConnect: s.LastConnect,
			BytesPassed: s.BytesPassed,
		})
	}

	return resp, nil
}
