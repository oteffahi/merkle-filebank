package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/oteffahi/merkle-filebank/merkle"
	pb "github.com/oteffahi/merkle-filebank/proto"
	"github.com/oteffahi/merkle-filebank/storage"
	"google.golang.org/protobuf/proto"
)

func CallDownloadFiles(bankhome, serverName, bankName string, fileNumber int) error {
	// verify that server exists locally
	if serverExists, err := storage.Client_ServerExists(bankhome, serverName); err != nil {
		return err
	} else if !serverExists {
		return errors.New(fmt.Sprintf("Server %v does not exist locally", serverName))
	}
	server, err := storage.Client_ReadServerDescriptor(bankhome, serverName)
	if err != nil {
		return err
	}

	// verify that bank exists
	if bankExist, err := storage.Client_BankExists(bankhome, serverName, bankName); err != nil {
		return err
	} else if !bankExist {
		return errors.New(fmt.Sprintf("Bank %v:%v does not exist", serverName, bankName))
	}
	bank, err := storage.Client_ReadBankDescriptor(bankhome, serverName, bankName)
	if err != nil {
		return err
	}

	// verify fileNumber exists in bank
	if fileNumber < int(bank.Nbfiles) || fileNumber > int(bank.Nbfiles) {
		return errors.New(fmt.Sprintf("No file identified by %v. Bank %v:%v has files between 1-%v", fileNumber, serverName, bankName, bank.Nbfiles))
	}

	conn, client, err := connectToNode(server.Host)
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // 30s timeout because of file manipulations on server
	defer cancel()

	stream, err := client.DownloadFiles(ctx)
	if err != nil {
		return err
	}

	resp1, err := stream.Recv()
	if err == io.EOF {
		return errors.New("Connexion closed by server")
	}
	if err != nil {
		return err
	}

	// verify msg type is nonce
	var serverNonce []byte
	switch phase := resp1.Phase.(type) {
	case *pb.DownloadFilesResponse_Nonce:
		serverNonce = phase.Nonce
	default:
		return errors.New("Invalid message type")
	}

	// TODO: get pubkey
	pubkey := []byte("key")

	// TODO: sign nonce
	sign, err := proto.Marshal(resp1)
	if err != nil {
		return err
	}

	// generate and send request message
	if err := stream.Send(&pb.DownloadFilesRequest{
		Nonce:     serverNonce,
		Pubkey:    pubkey,
		FileNum:   int32(fileNumber),
		Signature: sign,
	}); err != nil {
		return err
	}

	resp2, err := stream.Recv()
	if err == io.EOF {
		return errors.New("Connexion closed by server")
	}
	if err != nil {
		return err
	}

	var fileAndProof *pb.FileAndProof
	switch phase := resp2.Phase.(type) {
	case *pb.DownloadFilesResponse_Fp:
		fileAndProof = phase.Fp
	default:
		return errors.New("Invalid message type")
	}

	// unlinearize merkle proof
	var serverProof [][32]byte
	if len(fileAndProof.Proof)%32 != 0 {
		return errors.New("Invalid merkle proof format")
	}
	for i := 0; i < len(fileAndProof.Proof); i += 32 {
		var buff [32]byte
		copy(buff[:], fileAndProof.Proof[i:i+32])
		serverProof = append(serverProof, buff)
	}

	// verify proof
	merkleProof := merkle.MerkleProof{
		Hashes: serverProof,
	}
	validProof, err := merkleProof.VerifyFileProof(fileAndProof.File, [32]byte(bank.MerkleRoot))
	if err != nil {
		return err
	}
	if !validProof { //TODO: flip boolean when verification is implemented
		return errors.New("Invalid merkle proof")
	}

	// TODO: write file to disk

	return nil
}
