package client

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	cr "github.com/oteffahi/merkle-filebank/cryptography"
	"github.com/oteffahi/merkle-filebank/merkle"
	pb "github.com/oteffahi/merkle-filebank/proto"
	"github.com/oteffahi/merkle-filebank/storage"
)

func CallUploadFiles(bankhome, serverName, bankName string, filepaths []string) error {
	if len(filepaths) == 0 {
		return errors.New("Files list is empty")
	}

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
	// import server pubkey
	serverPubKey, err := cr.ImportPublicKey(server.PubKey)
	if err != nil {
		return err
	}

	// verify that bank does not exist
	if bankExist, err := storage.Client_BankExists(bankhome, serverName, bankName); err != nil {
		return err
	} else if bankExist {
		return errors.New(fmt.Sprintf("Bank %v:%v already exists", serverName, bankName))
	}

	// read files
	names, files, err := storage.ReadFilesFromPaths(filepaths)
	if err != nil {
		return err
	}

	// generate key-pair
	privKey, pubKey, passphrase, err := generateNewBankKey()
	if err != nil {
		return err
	}
	// export publicKey
	exportedPubKey, err := cr.ExportPublicKey(pubKey)
	if err != nil {
		return err
	}
	// export private key
	exportedPrivKey, err := cr.SafeExportPrivateKey(privKey, passphrase)
	if err != nil {
		return err
	}

	// encrypt files
	fileDescriptors := []*pb.FileDescriptor{}
	encFiles := [][]byte{}
	for i := 0; i < len(names); i++ {
		encryptedFile, salt, iv, err := cr.EncryptData(files[i], passphrase)
		if err != nil {
			return err
		}
		descriptor := &pb.FileDescriptor{
			Seq:  int32(i + 1),
			Name: names[i],
			Salt: salt,
			Iv:   iv,
		}
		encFiles = append(encFiles, encryptedFile)
		fileDescriptors = append(fileDescriptors, descriptor)
	}

	// generate merkle tree for files
	var tree merkle.MerkleTree
	if err = tree.BuildMerkleTree(encFiles); err != nil {
		return err
	}
	merkleRoot := tree.GetMerkleRoot()

	conn, client, err := connectToNode(server.Host, bankhome)
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

	// sign request
	messageToSign := &pb.SignUploadRequestClient{
		Nonce:   serverNonce,
		PubKey:  exportedPubKey,
		Nbfiles: int32(len(filepaths)),
	}
	sign, err := cr.SignMessage(messageToSign, privKey)
	if err != nil {
		return err
	}

	// send request
	req1 := &pb.UploadFilesRequest{
		Phase: &pb.UploadFilesRequest_SignedResp{
			SignedResp: &pb.ChallengeResponse{
				Nonce:     serverNonce,
				Pubkey:    exportedPubKey,
				Nbfiles:   int32(len(filepaths)),
				Signature: sign,
			},
		},
	}
	if err := stream.Send(req1); err != nil {
		return err
	}

	// Send files
	for i, file := range encFiles {
		req2 := &pb.UploadFilesRequest{
			Phase: &pb.UploadFilesRequest_File{
				File: &pb.FileMessage{
					Seq:     int32(i + 1),
					Content: file,
				},
			},
		}
		if err := stream.Send(req2); err != nil {
			return err
		}
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
	if err := stream.Send(req3); err != nil {
		return err
	}
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
	// verify signature
	if err := verifyMerkleRootSignature(signedResponse, serverPubKey); err != nil {
		return err
	}

	// verify merkle root
	if !bytes.Equal(signedResponse.MerkleRoot, merkleRoot[:]) {
		return errors.New("Server-side merkle tree different from local")
	}

	// write bank descriptor
	bankDescriptor := &pb.ClientBankDescriptor{
		PrivKey:         exportedPrivKey,
		Nbfiles:         int32(len(filepaths)),
		MerkleRoot:      signedResponse.MerkleRoot,
		FileDescriptors: fileDescriptors,
	}
	if err := storage.Client_WriteBankDescriptor(bankhome, bankDescriptor, serverName, bankName); err != nil {
		return err // TODO: maybe try to store somewhere else to save the filebank
	}
	fmt.Printf("Bank %s:%s has been succesfully created and uploaded\n", serverName, bankName)
	return nil
}

func verifyMerkleRootSignature(resp *pb.MerkleRoot, pubKey ed25519.PublicKey) error {
	signedMessage := &pb.SignMerkleRootServer{
		Nonce:      resp.Nonce,
		MerkleRoot: resp.MerkleRoot,
	}
	return cr.VerifySignature(signedMessage, pubKey, resp.Signature)
}

func generateNewBankKey() (ed25519.PrivateKey, ed25519.PublicKey, []byte, error) {
	fmt.Printf("Enter password for bank: ")
	firstPass, err := cr.ReadPassphrase()
	fmt.Println()
	if err != nil {
		return nil, nil, nil, err
	}
	fmt.Printf("Re-enter password for bank: ")
	pass, err := cr.ReadPassphrase()
	fmt.Println()
	if err != nil {
		return nil, nil, nil, err
	}
	if pass != firstPass {
		log.Fatalln("Passwords do not match. Aborting.")
	}
	pubKey, privKey, err := cr.GenerateKeyPair()
	if err != nil {
		return nil, nil, nil, err
	}

	return privKey, pubKey, []byte(pass), nil
}
