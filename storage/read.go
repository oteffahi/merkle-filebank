package storage

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"os"

	cr "github.com/oteffahi/merkle-filebank/cryptography"
	pb "github.com/oteffahi/merkle-filebank/proto"
	"google.golang.org/protobuf/proto"
)

func Server_ReadServerKey(bankhome string) ([]byte, error) {
	key, err := os.ReadFile(bankhome + "/server/priv.key")
	if err != nil {
		return nil, err
	}
	return key, nil
}

func Server_ReadBankDescriptor(bankhome string, clientPubKey ed25519.PublicKey) (*pb.ServerBankDescriptor, error) {
	keyHash := cr.HashOnce(clientPubKey)
	dirName, err := cr.Base58Encode(keyHash[:])
	if err != nil {
		return nil, err
	}
	desc, err := os.ReadFile(bankhome + "/server/" + dirName + "/bank.desc")
	if err != nil {
		return nil, err
	}

	var deserialized proto.Message
	if err := proto.Unmarshal(desc, deserialized); err != nil {
		return nil, err
	}

	descriptor, ok := deserialized.(*pb.ServerBankDescriptor)
	if !ok {
		return nil, errors.New(fmt.Sprintf("Bank descriptor for %v is malformed", dirName))
	}
	return descriptor, nil
}

func Client_ReadBankDescriptor(bankhome string, serverName string, bankName string) (*pb.ClientBankDescriptor, error) {
	desc, err := os.ReadFile(fmt.Sprintf("%s/client/srv_%s/bnk_%s.desc", bankhome, serverName, bankName))
	if err != nil {
		return nil, err
	}
	var deserialized proto.Message
	if err := proto.Unmarshal(desc, deserialized); err != nil {
		return nil, err
	}

	descriptor, ok := deserialized.(*pb.ClientBankDescriptor)
	if !ok {
		return nil, errors.New(fmt.Sprintf("Bank descriptor for %s:%s is malformed", serverName, bankName))
	}
	return descriptor, nil
}

func Client_ReadServerDescriptor(bankhome string, serverName string) (*pb.ServerDescriptor, error) {
	desc, err := os.ReadFile(fmt.Sprintf("%s/client/srv_%s/server.desc", bankhome, serverName))
	if err != nil {
		return nil, err
	}
	var deserialized proto.Message
	if err := proto.Unmarshal(desc, deserialized); err != nil {
		return nil, err
	}

	descriptor, ok := deserialized.(*pb.ServerDescriptor)
	if !ok {
		return nil, errors.New(fmt.Sprintf("Server descriptor for %s is malformed", serverName))
	}
	return descriptor, nil
}
