.PHONY: build

EXECUTABLE ?= arangodb
IMAGE ?= bin/$(EXECUTABLE)

all: build

build:
	CGO_ENABLED=0 go build --ldflags '${EXTLDFLAGS}' -o ${IMAGE} wwwin-github.cisco.com/spi-ie/vojltorb/arangodb

clean:
	rm ${IMAGE}

container:
	docker run -t -w /go/src/wwwin-github.cisco.com/spi-ie/vojltorb/arandodb -v `pwd`:/go/src/wwwin-github.cisco.com/spi-ie/vojltorb/arandodb golang:1.7.1-wheezy make
	docker build -t $(VOJLTORB_REPO):$(VOJLTORB_TAG) .
