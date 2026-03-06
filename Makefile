include rscli.mk

.PHONY:
bns: build-linux ship

build-linux:
	GOOS=linux GOARCH=amd64 go build -o mead ./cmd/service/main.go

ship:
	scp ./mead germ:~/mead/mead

# generates folders and installs dependencies
warmup:
	make .prepare-grpc-folders
	make .deps-grpc
	PROTOPACKPATH=proto_deps protopack mod download
# generates code on warm project
codegen:
	PROTOPACKPATH=proto_deps protopack generate
	#cd {{ .NPM_PACKAGE_PATH }} && npm run build

lint:
	golangci-lint run ./...