package server

import (
	"bytes"
	"errors"
	"io"

	cr "github.com/oteffahi/merkle-filebank/cryptography"
	"github.com/oteffahi/merkle-filebank/merkle"
	pb "github.com/oteffahi/merkle-filebank/proto"
	"google.golang.org/protobuf/proto"
)

func (c *fileBankServer) UploadFiles(stream pb.FileBankService_UploadFilesServer) error {
	serverNonce, err := cr.Random12BytesNonce()
	if err != nil {
		return err
	}
	stream.Send(&pb.UploadFilesResponse{
		Phase: &pb.UploadFilesResponse_Nonce{
			Nonce: serverNonce,
		},
	})

	req1, err := stream.Recv()
	if err == io.EOF {
		return errors.New("Connexion closed by client")
	}
	if err != nil {
		return err
	}

	var signedResp *pb.ChallengeResponse

	switch phase := req1.Phase.(type) {
	case *pb.UploadFilesRequest_SignedResp:
		signedResp = phase.SignedResp
	default:
		return errors.New("Invalid message type")
	}

	// verify nonce matches
	if !bytes.Equal(signedResp.Nonce, serverNonce) {
		return errors.New("Invalid challenge response nonce")
	}
	// TODO: verify signature
	validSignature, err := verifySignature(signedResp)
	if err != nil {
		return err
	}
	if !validSignature {
		return errors.New("Invalid challenge signature")
	}
	// TODO: verify key existence
	exists, err := verifyKeyExistence(signedResp.Pubkey)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("Key already has a bank")
	}

	// read files
	var files [][]byte
	for i := 0; i < int(signedResp.Nbfiles); i++ {
		req2, err := stream.Recv()
		if err == io.EOF {
			return errors.New("Connexion closed by client")
		}
		if err != nil {
			return err
		}

		var file *pb.FileMessage

		// verify type of message
		switch phase := req2.Phase.(type) {
		case *pb.UploadFilesRequest_File:
			file = phase.File
		default:
			return errors.New("Invalid message type")
		}
		// verify messages are in order
		if int(file.Seq) != i+1 {
			return errors.New("Invalid file order")
		}
		files = append(files, file.Content)
	}

	// generate merkle tree for files
	tree, err := generateMerkleTreeForFiles(files)
	if err != nil {
		return err
	}
	merkleRoot := tree.GetMerkleRoot()

	// read nonce
	req3, err := stream.Recv()
	if err == io.EOF {
		return errors.New("Connexion closed by client")
	}
	if err != nil {
		return err
	}

	var clientNonce []byte
	switch phase := req3.Phase.(type) {
	case *pb.UploadFilesRequest_Nonce:
		clientNonce = phase.Nonce
	default:
		return errors.New("Invalid message type")
	}

	// TODO: write everything to disk

	// TODO: sign response
	sign, err := proto.Marshal(req3)
	if err != nil {
		return err
	}

	resp := &pb.UploadFilesResponse{
		Phase: &pb.UploadFilesResponse_MerkleResponse{
			MerkleResponse: &pb.MerkleRoot{
				Nonce:      clientNonce,
				MerkleRoot: merkleRoot[:],
				Signature:  sign,
			},
		},
	}

	// only send when successfuly written to disk
	stream.Send(resp)

	return nil
}

func verifySignature(m *pb.ChallengeResponse) (bool, error) {
	return true, nil // TODO
}

func verifyKeyExistence(key []byte) (bool, error) {
	return false, nil // TODO
}

func generateMerkleTreeForFiles(files [][]byte) (*merkle.MerkleTree, error) {
	var tree merkle.MerkleTree
	err := tree.BuildMerkeTree(files)
	return &tree, err
}
