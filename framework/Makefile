.PHONY: vendor

EXECUTABLE ?= voltron
BINARY ?= bin/$(EXECUTABLE)
PACKAGES = $(shell go list ./... | grep -E -v '/bin|/vendor|/docs|/cmd|/log|/api/v1/client|/api/v1/gen|/arango/example')
REPO=dockerhub.cisco.com/gspie-docker
IMAGE=framework
TAG=latest


all: build

test:
	go test -ldflags -s -race -cover $(PACKAGES)

build:
	CGO_ENABLED=0 go build --ldflags '${EXTLDFLAGS}' -o ${BINARY} wwwin-github.cisco.com/spa-ie/voltron-redux/framework

linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ${BINARY} wwwin-github.cisco.com/spa-ie/voltron-redux/framework

fast:
	mkdir -p bin
	GOBIN=$(GOPATH)/src/wwwin-github.cisco.com/spa-ie/voltron-redux/framework/bin go install --ldflags '${EXTLDFLAGS}' wwwin-github.cisco.com/spa-ie/voltron-redux/framework
	mv bin/framework bin/voltron

container:
	docker run -t -w /go/src/wwwin-github.cisco.com/spa-ie/voltron-redux/framework -v `pwd`:/go/src/wwwin-github.cisco.com/spa-ie/voltron-redux/framework golang:1.7.1-wheezy make
	docker build -t ${REPO}/${IMAGE}:${TAG} .
