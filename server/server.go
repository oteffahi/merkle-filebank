package server

import (
	"log"
	"net"

	pb "github.com/oteffahi/merkle-filebank/proto"
	"google.golang.org/grpc"
)

type fileBankServer struct {
	pb.FileBankServiceServer
}

func RunServer(endpoint string) {
	conn, err := net.Listen("tcp", endpoint)

	if err != nil {
		log.Fatalf("Error occured when starting server: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterFileBankServiceServer(server, &fileBankServer{})

	log.Printf("Server listening on %v", conn.Addr())
	//list is the port, the grpc server needs to start there
	if err := server.Serve(conn); err != nil {
		log.Fatalf("Error occured when starting server: %v", err)
	}
}
