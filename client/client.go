package client

import (
	pb "github.com/oteffahi/merkle-filebank/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	endpoint = "0.0.0.0:5500"
)

type fileBankClient struct {
	pb.FileBankServiceServer
}

func connectToNode(endpoint string) (*grpc.ClientConn, pb.FileBankServiceClient, error) {
	conn, err := grpc.Dial(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		defer conn.Close()
		return nil, nil, err
	}

	client := pb.NewFileBankServiceClient(conn)
	return conn, client, nil
}
