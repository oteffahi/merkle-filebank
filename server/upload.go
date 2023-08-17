package server

import (
	"bytes"
	"crypto/ed25519"
	"errors"
	"io"
	"log"

	cr "github.com/oteffahi/merkle-filebank/cryptography"
	"github.com/oteffahi/merkle-filebank/merkle"
	pb "github.com/oteffahi/merkle-filebank/proto"
	"github.com/oteffahi/merkle-filebank/storage"
)

func (c *fileBankServer) UploadFiles(stream pb.FileBankService_UploadFilesServer) error {
	log.Printf("Received call: UploadFiles")
	serverNonce, err := cr.Random12BytesNonce()
	if err != nil {
		return err
	}
	if err := stream.Send(&pb.UploadFilesResponse{
		Phase: &pb.UploadFilesResponse_Nonce{
			Nonce: serverNonce,
		},
	}); err != nil {
		return err
	}

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
	// import pubKey
	clientPubKey, err := cr.ImportPublicKey(signedResp.Pubkey)
	if err != nil {
		return err
	}

	if err := verifyUploadChallengeResponseSignature(signedResp, clientPubKey); err != nil {
		return err
	}

	// check bank existence
	if exists, err := verifyBankExistence(signedResp.Pubkey); err != nil {
		return err
	} else if exists {
		return errors.New("Bank already exists")
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

	// convert tree.Hashes from slice of arrays to slice of slices
	merkleHashes := [][]byte{}
	for _, hash := range tree.Hashes {
		merkleHashes = append(merkleHashes, hash[:])
	}
	// write bank descriptor
	bankDescriptor := &pb.ServerBankDescriptor{
		PubKey:       signedResp.Pubkey,
		Nbfiles:      signedResp.Nbfiles,
		MerkleHashes: merkleHashes,
	}
	if err := storage.Server_WriteBankDescriptor(bankhome, bankDescriptor); err != nil {
		return err
	}

	// write files
	for i := 0; i < len(files); i++ {
		if err := storage.Server_WriteFileToBank(bankhome, signedResp.Pubkey, files[i], i+1); err != nil {
			return err
		}
	}

	// files stored correctly. Sign response
	msgToSign := &pb.SignMerkleRootServer{
		Nonce:      clientNonce,
		MerkleRoot: merkleRoot[:],
	}
	sign, err := cr.SignMessage(msgToSign, ServerKeys.privKey)
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
	if err := stream.Send(resp); err != nil {
		return err
	}
	return nil
}

func verifyUploadChallengeResponseSignature(resp *pb.ChallengeResponse, pubKey ed25519.PublicKey) error {
	clientSignedMsg := &pb.SignUploadRequestClient{
		Nonce:   resp.Nonce,
		PubKey:  resp.Pubkey,
		Nbfiles: resp.Nbfiles,
	}
	return cr.VerifySignature(clientSignedMsg, pubKey, resp.Signature)
}

func verifyBankExistence(clientPubKey []byte) (bool, error) {
	keyHash := cr.HashOnce(clientPubKey)
	keyHashB58 := cr.Base58Encode(keyHash[:])
	if exists, err := storage.Server_BankExists(bankhome, keyHashB58); err != nil {
		return false, err
	} else if exists {
		return true, nil
	}
	return false, nil
}

func generateMerkleTreeForFiles(files [][]byte) (*merkle.MerkleTree, error) {
	var tree merkle.MerkleTree
	err := tree.BuildMerkeTree(files)
	return &tree, err
}
