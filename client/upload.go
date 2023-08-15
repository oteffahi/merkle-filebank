package client

import (
	"bytes"
	"context"
	"errors"
	"io"
	"time"

	cr "github.com/oteffahi/merkle-filebank/cryptography"
	"github.com/oteffahi/merkle-filebank/merkle"
	pb "github.com/oteffahi/merkle-filebank/proto"
	"google.golang.org/protobuf/proto"
)

func CallUploadFiles(endpoint string, files [][]byte) error {
	if len(files) == 0 {
		return errors.New("Files list is empty")
	}

	// generate merkle tree for files
	var tree merkle.MerkleTree
	err := tree.BuildMerkeTree(files)
	if err != nil {
		return err
	}
	merkleRoot := tree.GetMerkleRoot()

	conn, client, err := connectToNode(endpoint)
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // 30s timeout because of file manipulations on server
	defer cancel()

	stream, err := client.UploadFiles(ctx)
	if err != nil {
		return err
	}

	resp1, err := stream.Recv()
	if err == io.EOF {
		return errors.New("Connexion closed by server")
	}

	var serverNonce []byte
	switch phase := resp1.Phase.(type) {
	case *pb.UploadFilesResponse_Nonce:
		serverNonce = phase.Nonce
	default:
		return errors.New("Invalid message type")
	}

	// TODO: sign request
	pubkey := []byte("TODO")
	sign, err := proto.Marshal(resp1)

	// generate request instance
	req1 := &pb.UploadFilesRequest{
		Phase: &pb.UploadFilesRequest_SignedResp{
			SignedResp: &pb.ChallengeResponse{
				Nonce:     serverNonce,
				Pubkey:    pubkey,
				Nbfiles:   int32(len(files)),
				Signature: sign,
			},
		},
	}
	stream.Send(req1)

	// Send files
	for i, file := range files {
		req2 := &pb.UploadFilesRequest{
			Phase: &pb.UploadFilesRequest_File{
				File: &pb.FileMessage{
					Seq:     int32(i + 1),
					Content: file,
				},
			},
		}
		stream.Send(req2)
	}
	// send Nonce
	clientNonce, err := cr.Random12BytesNonce()
	if err != nil {
		return err
	}
	req3 := &pb.UploadFilesRequest{
		Phase: &pb.UploadFilesRequest_Nonce{
			Nonce: clientNonce,
		},
	}
	stream.Send(req3)

	// receive signed response
	resp2, err := stream.Recv()
	if err == io.EOF {
		return errors.New("Connexion closed by server")
	}
	if err != nil {
		return err
	}
	var signedResponse *pb.MerkleRoot
	switch phase := resp2.Phase.(type) {
	case *pb.UploadFilesResponse_MerkleResponse:
		signedResponse = phase.MerkleResponse
	default:
		return errors.New("Invalid message type")
	}

	// verify nonce
	if !bytes.Equal(signedResponse.Nonce, clientNonce) {
		return errors.New("Invalid challenge response nonce")
	}
	// TODO: verify signature
	validSignature, err := verifyMerkleRootSignature(signedResponse)
	if err != nil {
		return err
	}
	if !validSignature {
		return errors.New("Invalid challenge signature")
	}

	// verify merkle root
	if !bytes.Equal(signedResponse.MerkleRoot, merkleRoot[:]) {
		return errors.New("Server-side merkle tree different from local")
	}

	// TODO: write everything to disk

	return nil
}

func verifyMerkleRootSignature(msg *pb.MerkleRoot) (bool, error) {
	return true, nil // TODO
}
