# Merkle File Bank

A CLI tool for running and using a merkle-based file storage service. The project is entirely developped in Go.

- The Merkle tree generation code is inspired from [OpenZeppelin's implementation](https://github.com/OpenZeppelin/merkle-tree).
- Client-Server communications use google RPC (gRPC) and Protobuf.
- Data is stored within a simple filesystem-based directory tree, using Protobuf for serialization.
- Files are encrypted in AES-GCM-128 before upload to server.
- Each filebank is identified by an Ed25519 private key, encrypted and stored in pkcs8 DER format.
- Each bank is protected by a passphrase that is used to decrypt the ed25519 private key, and seeds a PBKDF2 function to generate one distinct AES encryption key for each file in the bank.
- Authentication of banks is based on a simple signature challenge-response scheme.
- All communication is encrypted and authenticated using server-side SSL/TLS.
- CLI is powered by [Cobra](https://github.com/spf13/cobra).
## 1. Compiling

```console
$ go build -o filebankd filebankd/main.go
```

## 2. Usage

### 2.1. Running server

```console
$ filebankd start
Server key not found. Creating new key
Enter password for key: 
Re-enter password for key: 
Starting server with TLS enabled...
2023/08/18 09:47:47 Server listening on [::]:5500
```

### 2.2. Adding server to client

```console
$ filebankd server add --address server1.filebank.fr MyServer1
Server 'MyServer1' was successfully added to known servers
$ filebankd server list
                Name                      Host
===========================================================
           MyServer1            server1.filebank.fr:5500
```

### 2.3. Creating bank on server

```console
$ filebankd bank create -s MyServer1 -b MyBank1 ../test/ ../files/
Scanning ../test
Adding ../test/LICENSE
Scanning ../test/cmd
Adding ../test/cmd/bank.go
Adding ../test/cmd/root.go
Adding ../test/cmd/serveradd.go
Adding ../test/go.mod
Adding ../test/go.sum
Adding ../test/main.go
Scanning ../files
Adding ../files/test1.txt
Adding ../files/test2.docx
Adding ../files/test3.pdf
Adding ../files/test4
Enter password for bank: 
Re-enter password for bank: 
Bank MyServer1:MyBank1 has been succesfully created and uploaded
$ filebankd bank list -s MyServer1
Banks for server 'MyServer1'
=====================================
        MyBank1
$ filebankd bank list -s MyServer1 -b MyBank1
Files for bank 'MyServer1:MyBank1'
=====================================
    1  LICENSE
    2  bank.go
    3  root.go
    4  serveradd.go
    5  go.mod
    6  go.sum
    7  main.go
    8  test1.txt
    9  test2.docx
   10  test3.pdf
   11  test4
```

### 2.4. Pulling file from bank
```console
$ filebankd bank pull -s MyServer1 -b MyBank1 8 
Enter bank password: 
File written to /home/filebankd/.filebankd/downloads/test1.txt
Successfully downloaded, verified and decrypted file 8 from bank MyServer1:MyBank1
$ cat /home/filebankd/.filebankd/downloads/test1.txt 
MyTest1
```

## 3. Deploying

### 3.1. Running containers

A Makefile and docker-compose are provided to deploy a simple testbed containing 6 containers: server1, server2, server3, client1, client2, client3.

Each instance is identified in the network by a network-alias `name.filebank.fr` where name is the instance's name.

Run the following commands in order :
```console
$ make build
$ make start
```

### 3.2. Interactig with containers

Run the following command to interact with an instance. Note that `name` is the name of the instance in question.

```console
$ docker-compose exec -it name bash
```
