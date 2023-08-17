package server

import (
	"crypto/ed25519"
	"crypto/tls"
	"fmt"
	"log"
	"net"

	cr "github.com/oteffahi/merkle-filebank/cryptography"
	pb "github.com/oteffahi/merkle-filebank/proto"
	"github.com/oteffahi/merkle-filebank/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type fileBankServer struct {
	pb.FileBankServiceServer
}

func handleError(err error) {
	log.Fatalf("Error occured when starting server: %v", err)
}

func RunServer(endpoint string) {
	var privKey ed25519.PrivateKey
	if homeWellFormed, err := storage.IsHomeWellFormed(bankhome); err != nil {
		handleError(err)
	} else if !homeWellFormed {
		handleError(fmt.Errorf("Bankhome %v is not initialized or is malformed\n", bankhome))
	}
	if keyExists, err := storage.Server_ServerKeyExists(bankhome); err != nil {
		handleError(err)
	} else if !keyExists {
		// create new key
		privKey, err = generateNewServerKey()
		if err != nil {
			handleError(err)
		}
	} else {
		// read key
		encryptedKey, err := storage.Server_ReadServerKey(bankhome)
		if err != nil {
			handleError(err)
		}
		fmt.Printf("Enter password for server key: ")
		pass, err := cr.ReadPassphrase()
		fmt.Println()
		if err != nil {
			handleError(err)
		}
		privKey, err = cr.SafeImportPrivateKey(encryptedKey, []byte(pass))
		if err != nil {
			handleError(err)
		}
	}
	if err := LoadKeyPair(privKey); err != nil {
		handleError(err)
	}

	conn, err := net.Listen("tcp", endpoint)

	if err != nil {
		handleError(err)
	}

	var server *grpc.Server
	// try to load cert
	serverCert, err := tls.LoadX509KeyPair(bankhome+"/cert/filebank-server-cert.pem", bankhome+"/cert/filebank-server-key.pem")
	if err != nil {
		fmt.Println("Could not load certificate and key. Starting without TLS...")
		server = grpc.NewServer()
	} else {
		fmt.Println("Starting server with TLS enabled...")
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{serverCert},
			ClientAuth:   tls.NoClientCert,
		}
		server = grpc.NewServer(
			grpc.Creds(credentials.NewTLS(tlsConfig)),
		)
	}

	pb.RegisterFileBankServiceServer(server, &fileBankServer{})

	log.Printf("Server listening on %v", conn.Addr())

	if err := server.Serve(conn); err != nil {
		handleError(err)
	}
}

func generateNewServerKey() (ed25519.PrivateKey, error) {
	fmt.Println("Server key not found. Creating new key")
	fmt.Printf("Enter password for key: ")
	firstPass, err := cr.ReadPassphrase()
	fmt.Println()
	if err != nil {
		return nil, err
	}
	fmt.Printf("Re-enter password for key: ")
	pass, err := cr.ReadPassphrase()
	fmt.Println()
	if err != nil {
		return nil, err
	}
	if pass != firstPass {
		log.Fatalln("Passwords do not match. Aborting.")
	}
	_, privKey, err := cr.GenerateKeyPair()
	if err != nil {
		return nil, err
	}
	exportedKey, err := cr.SafeExportPrivateKey(privKey, []byte(pass))
	if err != nil {
		return nil, err
	}
	if err := storage.Server_WriteServerKey(bankhome, exportedKey); err != nil {
		return nil, err
	}
	return privKey, nil
}
