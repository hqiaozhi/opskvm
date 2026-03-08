ROOT_DIR    = $(shell pwd)
NAMESPACE   = "default"
DEPLOY_NAME = "template-single"
DOCKER_NAME = "template-single"

VERSION=v0.0.2

include ./hack/hack-cli.mk
include ./hack/hack.mk

pack:
	gf pack resource/public/html/dist internal/packed/data.go -n packed -y

dep:
	@go mod tidy

gf: dep
	@mkdir -p bin
	GOOS=linux GOARCH=amd64  gf build -ew -o  bin/opskvm-linux-amd64-${VERSION} main.go
	GOOS=linux GOARCH=arm64  gf build -ew -o  bin/opskvm-linux-arm64-${VERSION} main.go
	@chmod +x -R bin/*

go: dep
	@mkdir -p bin
	GOOS=linux GOARCH=amd64  go build -ldflags "-s -w" -o  bin/opskvm-linux-amd64-${VERSION} main.go
	GOOS=linux GOARCH=arm64  go build -ldflags "-s -w" -o  bin/opskvm-linux-arm64-${VERSION} main.go
	@chmod +x -R bin/*

install: build
	install -m 755 bin/opskvm-linux-arm64-${VERSION} /usr/local/bin/opskvm