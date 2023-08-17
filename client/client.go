package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	pb "github.com/oteffahi/merkle-filebank/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	endpoint = "0.0.0.0:5500"
)

type fileBankClient struct {
	pb.FileBankServiceServer
}

func connectToNode(endpoint string, bankhome string) (*grpc.ClientConn, pb.FileBankServiceClient, error) {
	var conn *grpc.ClientConn
	// try to load cert
	serverCA, err := os.ReadFile(bankhome + "/cert/filebank-ca-cert.pem") // read /etc/ssl/cert.pem in production
	if err == nil {
		certPool := x509.NewCertPool()
		if certPool.AppendCertsFromPEM(serverCA) {
			config := &tls.Config{
				RootCAs: certPool,
			}
			conn, err = grpc.Dial(endpoint, grpc.WithTransportCredentials(credentials.NewTLS(config)))
			client := pb.NewFileBankServiceClient(conn)
			return conn, client, nil
		}
	}

	fmt.Println("!!! Could not load certificate. Connecting without TLS...")
	conn, err = grpc.Dial(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		defer conn.Close()
		return nil, nil, err
	}

	client := pb.NewFileBankServiceClient(conn)
	return conn, client, nil
}
