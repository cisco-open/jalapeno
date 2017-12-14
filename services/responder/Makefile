.PHONY: build test integration fast container debug

EXECUTABLE ?= responder
BINARY ?= bin/$(EXECUTABLE)
BASE ?= dockerhub.cisco.com/gspie-dev-docker
IMAGE ?= responder
TAG ?= 1.0
all: build

build:
	CGO_ENABLED=0 go build --ldflags '${EXTLDFLAGS}' -o ${BINARY} wwwin-github.cisco.com/spa-ie/voltron-redux/services/responder

clean:
	rm ${BINARY}

fast:
	mkdir -p bin
	GOBIN=$(GOPATH)/src/wwwin-github.cisco.com/spa-ie/voltron-redux/services/responder/bin go install --ldflags '${EXTLDFLAGS}' wwwin-github.cisco.com/spa-ie/voltron-redux/services/responder

container:
	docker run -t -w /go/src/wwwin-github.cisco.com/spa-ie/voltron-redux/services/responder -v `pwd`:/go/src/wwwin-github.cisco.com/spa-ie/voltron-redux/services/responder golang:1.7.1-wheezy make
	docker build -t $(BASE)/$(IMAGE):$(TAG) .
