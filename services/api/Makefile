EXECUTABLE ?= voltron
BINARY ?= bin/$(EXECUTABLE)
REPO=ievoltron
IMAGE=api
TAG=0.0.1.3
PROXY=http://proxy.esl.cisco.com:8080

all: build

build:
	cd src/ && pip3 install -r requirements.txt

run:
	cd src/ && python3 -m swagger_server

container:
	docker build --build-arg http_proxy=${PROXY} --build-arg https_proxy=${PROXY} -t ${REPO}/${IMAGE}:${TAG} src/
