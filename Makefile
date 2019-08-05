PKG:=github.com/sapcc/nova-password
APP_NAME:=nova-password
PWD:=$(shell pwd)
UID:=$(shell id -u)
VERSION:=$(shell git describe --tags --always --dirty="-dev")
LDFLAGS:=-X main.Version=$(VERSION)

export GO111MODULE:=off
export GOPATH:=$(PWD)/gopath
export CGO_ENABLED:=0

build: gopath/src/$(PKG) fmt
	GOOS=linux go build -ldflags="$(LDFLAGS)" -o bin/$(APP_NAME) $(PKG)
	GOOS=darwin go build -ldflags="$(LDFLAGS)" -o bin/$(APP_NAME)_darwin $(PKG)
	GOOS=windows go build -ldflags="$(LDFLAGS)" -o bin/$(APP_NAME).exe $(PKG)

docker:
	docker run -ti --rm -e GOCACHE=/tmp -v $(PWD):/$(APP_NAME) -u $(UID):$(UID) --workdir /$(APP_NAME) golang:latest make

fmt:
	gofmt -s -w *.go

mod:
	GO111MODULE=auto go mod download
	GO111MODULE=auto go mod tidy
	GO111MODULE=auto go mod vendor

gopath/src/$(PKG):
	mkdir -p gopath/src/$(shell dirname $(PKG))
	ln -sf ../../../.. gopath/src/$(PKG)
