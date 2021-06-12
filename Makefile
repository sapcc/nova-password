PKG:=github.com/sapcc/nova-password
APP_NAME:=nova-password
PWD:=$(shell pwd)
UID:=$(shell id -u)
VERSION:=$(shell git describe --tags --always --dirty="-dev")
LDFLAGS:=-X main.Version=$(VERSION) -w -s

export CGO_ENABLED:=0

build: fmt vet
	GOOS=linux go build -mod=vendor -ldflags="$(LDFLAGS)" -o bin/$(APP_NAME) $(PKG)
	GOOS=darwin go build -mod=vendor -ldflags="$(LDFLAGS)" -o bin/$(APP_NAME)_darwin $(PKG)
	GOOS=windows go build -mod=vendor -ldflags="$(LDFLAGS)" -o bin/$(APP_NAME).exe $(PKG)

docker:
	docker pull golang:latest
	docker run -ti --rm -e GOCACHE=/tmp -v $(PWD):/$(APP_NAME) -u $(UID):$(UID) --workdir /$(APP_NAME) golang:latest make

fmt:
	gofmt -s -w *.go

vet:
	go vet -mod=vendor ./

static:
	staticcheck ./

mod:
	go mod vendor
