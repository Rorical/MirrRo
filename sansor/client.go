package sansor

import (
	"context"
	"github.com/Rorical/MirrRo/sansor/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type SansorClient struct {
	client pb.SansorClient
}

func NewSansorClient(uri string) (*SansorClient, error) {
	conn, err := grpc.Dial(uri, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	cli := pb.NewSansorClient(conn)
	return &SansorClient{
		client: cli,
	}, nil
}

func (cli *SansorClient) TextReview(ctx context.Context, text string) (bool, error) {
	request := &pb.TextReviewRequest{Text: text}
	response, err := cli.client.TextReview(ctx, request, grpc.MaxCallSendMsgSize(10000000))
	if err != nil {
		return false, err
	}
	return response.GetBanned(), nil
}
