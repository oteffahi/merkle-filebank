package storage

import (
	"fmt"
	"os"
	"strings"

	pb "github.com/oteffahi/merkle-filebank/proto"
	"google.golang.org/protobuf/proto"
)

func IsHomeWellFormed(bankhome string) (bool, error) {
	// root
	if _, err := os.Stat(bankhome); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	// server
	if _, err := os.Stat(bankhome + "/server"); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	// client
	if _, err := os.Stat(bankhome + "/client"); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	// downloads
	if _, err := os.Stat(bankhome + "/downloads"); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func Server_ReadServerKey(bankhome string) ([]byte, error) {
	key, err := os.ReadFile(bankhome + "/server/priv.key")
	if err != nil {
		return nil, err
	}
	return key, nil
}

func Server_ServerKeyExists(bankhome string) (bool, error) {
	if _, err := os.Stat(bankhome + "/server/priv.key"); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func Server_BankExists(bankhome string, pubKeyHashB58 string) (bool, error) {
	// clientPubKey is assumed hashed and b58encoded in exported format
	dirName := pubKeyHashB58
	if _, err := os.Stat(bankhome + "/server/" + dirName); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func Server_ReadBankDescriptor(bankhome string, pubKeyHashB58 string) (*pb.ServerBankDescriptor, error) {
	// clientPubKey is assumed hashed and b58encoded in exported format
	dirName := pubKeyHashB58
	desc, err := os.ReadFile(bankhome + "/server/" + dirName + "/bank.desc")
	if err != nil {
		return nil, err
	}

	descriptor := &pb.ServerBankDescriptor{}
	if err := proto.Unmarshal(desc, descriptor); err != nil {
		return nil, err
	}

	return descriptor, nil
}

func Server_ReadFileFromBank(bankhome string, pubKeyHashB58 string, fileNum int) ([]byte, error) {
	// clientPubKey is assumed hashed and b58encoded in exported format
	dirName := pubKeyHashB58

	file, err := os.ReadFile(fmt.Sprintf("%s/server/%s/%d", bankhome, dirName, fileNum))
	if err != nil {
		return nil, err
	}

	return file, nil
}

func Client_BankExists(bankhome string, serverName string, bankName string) (bool, error) {
	if _, err := os.Stat(fmt.Sprintf("%s/client/srv_%s/bnk_%s.desc", bankhome, serverName, bankName)); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func Client_ServerExists(bankhome string, serverName string) (bool, error) {
	if _, err := os.Stat(fmt.Sprintf("%s/client/srv_%s", bankhome, serverName)); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func Client_ReadBankDescriptor(bankhome string, serverName string, bankName string) (*pb.ClientBankDescriptor, error) {
	desc, err := os.ReadFile(fmt.Sprintf("%s/client/srv_%s/bnk_%s.desc", bankhome, serverName, bankName))
	if err != nil {
		return nil, err
	}
	descriptor := &pb.ClientBankDescriptor{}
	if err := proto.Unmarshal(desc, descriptor); err != nil {
		return nil, err
	}

	return descriptor, nil
}

func Client_ReadServerDescriptor(bankhome string, serverName string) (*pb.ServerDescriptor, error) {
	desc, err := os.ReadFile(fmt.Sprintf("%s/client/srv_%s/server.desc", bankhome, serverName))
	if err != nil {
		return nil, err
	}

	descriptor := &pb.ServerDescriptor{}
	if err := proto.Unmarshal(desc, descriptor); err != nil {
		return nil, err
	}

	return descriptor, nil
}

func Client_ListServers(bankhome string) (serverNames []string, serverHosts []string, err error) {
	dscriptors, err := os.ReadDir(bankhome + "/client")
	if err != nil {
		return nil, nil, err
	}
	for _, descriptor := range dscriptors {
		fileName := descriptor.Name()
		if strings.HasPrefix(fileName, "srv_") {
			serverName, _ := strings.CutPrefix(fileName, "srv_")
			// read descriptor
			descBytes, err := os.ReadFile(bankhome + "/client/" + fileName + "/" + "server.desc")
			if err != nil {
				return nil, nil, err
			}
			// deserialize
			descriptor := &pb.ServerDescriptor{}
			if err := proto.Unmarshal(descBytes, descriptor); err != nil {
				return nil, nil, err
			}
			serverNames = append(serverNames, serverName)
			serverHosts = append(serverHosts, descriptor.Host)
		}
	}
	return serverNames, serverHosts, nil
}

func GetAllFilesPaths(rootPath string) ([]string, error) {
	if rootPath[len(rootPath)-1] == '/' { // remove trailing slash
		rootPath = rootPath[:len(rootPath)-1]
	}
	fileInfo, err := os.Stat(rootPath)
	if err != nil {
		return []string{}, err
	}

	if fileInfo.IsDir() {
		// get all new paths
		files, err := os.ReadDir(rootPath)
		if err != nil {
			return []string{}, err
		}

		var paths []string
		for _, file := range files {
			filePath := rootPath + "/" + file.Name()
			recursivePaths, err := GetAllFilesPaths(filePath)
			if err != nil {
				return []string{}, err
			}
			paths = append(paths, recursivePaths...)
		}
		return paths, nil
	} else {
		return []string{rootPath}, nil
	}
}

func ReadFilesFromPaths(paths []string) (names []string, files [][]byte, err error) {
	for _, path := range paths {
		file, err := os.Stat(path)
		if err != nil {
			return nil, nil, err
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, err
		}
		files = append(files, content)
		names = append(names, file.Name())
	}
	return names, files, nil
}
