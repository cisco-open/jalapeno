EXECUTABLE ?= jalapeno
BINARY ?= bin/$(EXECUTABLE)
REPO=iejalapeno
IMAGE=topology
TAG=0.0.3.3
# If you connect to dockerhub via a proxy: uncomment and edit the following line (and see below docker build lines)
#PROXY=http://proxy.esl.cisco.com:8080

all: build

build:
	CGO_ENABLED=0 go build --ldflags '${EXTLDFLAGS}' -o ${BINARY} github.com/cisco-ie/jalapeno/processors/topology

linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ${BINARY} github.com/cisco-ie/jalapeno/processors/topology

container:
	# Default docker build without proxy setting:
	docker build -t ${REPO}/${IMAGE}:${TAG} .	

	# Uncomment this line (and comment out the above line) if you connect to dockerhub via a proxy
	#docker build --build-arg http_proxy=${PROXY} --build-arg https_proxy=${PROXY} -t ${REPO}/${IMAGE}:${TAG} .
