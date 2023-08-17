package storage

import (
	"errors"
	"fmt"
	"os"

	cr "github.com/oteffahi/merkle-filebank/cryptography"
	pb "github.com/oteffahi/merkle-filebank/proto"
	"google.golang.org/protobuf/proto"
)

func InitHome(bankhome string) error {
	if _, err := os.Stat(bankhome); os.IsNotExist(err) {
		if err := os.MkdirAll(bankhome+"/server", os.ModeDir); err != nil {
			return err
		}
		if err := os.MkdirAll(bankhome+"/client", os.ModeDir); err != nil {
			return err
		}
		if err := os.MkdirAll(bankhome+"/downloads", os.ModeDir); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func Server_WriteServerKey(bankhome string, key []byte) error {
	err := os.WriteFile(bankhome+"/server/priv.key", key, 0400)
	if err != nil {
		return err
	}
	return nil
}

func Server_WriteBankDescriptor(bankhome string, descriptor *pb.ServerBankDescriptor) error {
	pubKey := descriptor.PubKey
	keyHash := cr.HashOnce(pubKey)
	dirName := cr.Base58Encode(keyHash[:])

	if _, err := os.Stat(bankhome + "/server/" + dirName); !os.IsNotExist(err) {
		return errors.New("Client key already has bank")
	}

	data, err := proto.Marshal(descriptor)
	if err != nil {
		return err
	}

	if err := os.Mkdir(bankhome+"/server/"+dirName, os.ModeDir); err != nil {
		return err
	}
	if err := os.WriteFile(bankhome+"/server/"+dirName+"/bank.desc", data, 0400); err != nil {
		return err
	}
	return nil
}

func Server_WriteFileToBank(bankhome string, clientPubKey []byte, file []byte, fileNum int) error {
	keyHash := cr.HashOnce(clientPubKey)
	dirName := cr.Base58Encode(keyHash[:])

	if err := os.WriteFile(fmt.Sprintf("%s/server/%s/%d", bankhome, dirName, fileNum), file, 0444); err != nil {
		return err
	}
	return nil
}

func Client_WriteBankDescriptor(bankhome string, descriptor *pb.ClientBankDescriptor, serverName string, bankName string) error {
	serverPath := fmt.Sprintf("%s/client/srv_%s", bankhome, serverName)
	bankPath := fmt.Sprintf("%s/bnk_%s.desc", serverPath, bankName)
	// serverPath must exist
	if _, err := os.Stat(serverPath); os.IsNotExist(err) {
		return errors.New("Server " + serverName + " does not exist")
	}
	// bank must not exist
	if _, err := os.Stat(bankPath); !os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("Bank %s:%s already exists", serverName, bankName))
	}

	data, err := proto.Marshal(descriptor)
	if err != nil {
		return err
	}
	if err := os.WriteFile(bankPath, data, 0400); err != nil {
		return err
	}
	return nil
}

func Client_WriteServerDescriptor(bankhome string, descriptor *pb.ServerDescriptor, serverName string) error {
	serverPath := fmt.Sprintf("%s/client/srv_%s", bankhome, serverName)
	// serverPath must not exist
	if _, err := os.Stat(serverPath); !os.IsNotExist(err) {
		return errors.New("Server " + serverName + " already exists")
	}

	// create server directory
	if err := os.Mkdir(serverPath, os.ModeDir); err != nil {
		return err
	}

	data, err := proto.Marshal(descriptor)
	if err != nil {
		return err
	}
	if err := os.WriteFile(serverPath+"/server.desc", data, 0444); err != nil {
		return err
	}
	return nil
}
