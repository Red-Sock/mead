.PHONY:
bns: build-linux ship

build-linux:
	GOOS=linux GOARCH=amd64 go build -o mead ./cmd/service/main.go

ship:
	scp ./mead germ:~/mead/mead