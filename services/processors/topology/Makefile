EXECUTABLE ?= jalapeno
BINARY ?= bin/$(EXECUTABLE)
REPO=iejalapeno
IMAGE=topology
TAG=2.1.2
PROXY=http://proxy.esl.cisco.com:8080

all: build

build:
	CGO_ENABLED=0 go build --ldflags '${EXTLDFLAGS}' -o ${BINARY} wwwin-github.cisco.com/spa-ie/jalapeno/services/collectors/topology

linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ${BINARY} wwwin-github.cisco.com/spa-ie/jalapeno/services/collectors/topology

container:
	docker build --build-arg http_proxy=${PROXY} --build-arg https_proxy=${PROXY} -t ${REPO}/${IMAGE}:${TAG} .
