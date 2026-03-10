ROOT_DIR    = $(shell pwd)
NAMESPACE   = "default"
DEPLOY_NAME = "template-single"
DOCKER_NAME = "template-single"

VERSION=v0.0.3

include ./hack/hack-cli.mk
include ./hack/hack.mk

rmbin:
	@rm -rf bin

pack:
	@gf pack resource/public/html/dist internal/packed/data.go -n packed -y
	@rm -rf resource/public/html/dist

dep:
	@go mod tidy

gf: pack dep rmbin
	@mkdir -p bin/gf
	@GOOS=linux GOARCH=amd64  gf build -ew -o  bin/gf/opskvm-linux-amd64-${VERSION} main.go
	@GOOS=linux GOARCH=arm64  gf build -ew -o  bin/gf/opskvm-linux-arm64-${VERSION} main.go
	@chmod +x -R bin/gf/*

go: pack dep rmbin
	@mkdir -p bin/go
	GOOS=linux GOARCH=amd64  go build -ldflags "-s -w" -o  bin/go/opskvm-linux-amd64-${VERSION} main.go
	GOOS=linux GOARCH=arm64  go build -ldflags "-s -w" -o  bin/go/opskvm-linux-arm64-${VERSION} main.go
	@chmod +x -R bin/go/*

install:
	install -m 755 bin/go/opskvm-linux-arm64-${VERSION} /usr/local/bin/opskvm