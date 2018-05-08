.PHONY: vendor

EXECUTABLE ?= voltron
BINARY ?= bin/$(EXECUTABLE)
PACKAGES = $(shell go list ./... | grep -E -v '/bin|/vendor|/docs|/cmd|/log|/api/v1/client|/api/v1/gen|/arango/example')
REPO=dockerhub.cisco.com/gspie-docker-local
IMAGE=topology
TAG=1.7


all: build

test:
	go test -ldflags -s -race -cover $(PACKAGES)

build:
	CGO_ENABLED=0 go build --ldflags '${EXTLDFLAGS}' -o ${BINARY} wwwin-github.cisco.com/spa-ie/voltron/services/collectors/topology

linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ${BINARY} wwwin-github.cisco.com/spa-ie/voltron/services/collectors/topology

fast:
	mkdir -p bin
	GOBIN=$(GOPATH)/src/wwwin-github.cisco.com/spa-ie/voltron/services/collectors/topology/bin go install --ldflags '${EXTLDFLAGS}' wwwin-github.cisco.com/spa-ie/voltron/services/collectors/topology
	mv bin/services/collectors/topology bin/voltron

container:
	docker run -t -w /go/src/wwwin-github.cisco.com/spa-ie/voltron/services/collectors/topology -v `pwd`:/go/src/wwwin-github.cisco.com/spa-ie/voltron/services/collectors/topology golang:1.7.1-wheezy make
	docker build -t ${REPO}/${IMAGE}:${TAG} .
