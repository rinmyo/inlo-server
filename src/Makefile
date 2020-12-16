gen:
	protoc --proto_path=proto proto/*.proto --proto_path=/Users/satanya/api-common-protos  --go_out=plugins=grpc:pb

clean:
	rm -rf pb/*.pb.go

run:
	go run . -port 8080