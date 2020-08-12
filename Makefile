serve:
	go run .

protoc:
	protoc --go_out=plugins=grpc:./proto -I./proto --go_opt=paths=source_relative ./proto/*.proto
