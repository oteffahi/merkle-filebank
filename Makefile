protos:
	protoc --go_out=. --go-grpc_out=. proto/filebank.proto proto/signed.proto proto/storage.proto

gencerts:
	rm ./certs/*/*.pem || rm ./certs/*/*.srl || true

	openssl req -x509 -newkey rsa:4096 -days 365 -nodes -keyout ./certs/ca/filebank-ca-key.pem -out ./certs/ca/filebank-ca-cert.pem -subj "/C=FR/ST=Ile de France/L=Paris/O=MerkleFileBank/OU=MerkleFileBank/CN=*.filebank.fr/emailAddress=ca@filebank.fr"

	openssl req -newkey rsa:4096 -nodes -keyout ./certs/server/filebank-server-key.pem -out ./certs/server/filebank-server-req.pem -subj "/C=FR/ST=Ile de France/L=Paris/O=MerkleFileBank/OU=Server/CN=*.filebank.fr/emailAddress=servers@filebank.fr/"

	openssl x509 -req -in ./certs/server/filebank-server-req.pem -days 60 -CA ./certs/ca/filebank-ca-cert.pem -CAkey ./certs/ca/filebank-ca-key.pem -CAcreateserial -out ./certs/server/filebank-server-cert.pem -extfile certs/ext.cnf