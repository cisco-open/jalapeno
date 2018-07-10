EXECUTABLE ?= voltron
BINARY ?= bin/$(EXECUTABLE)
REPO=dockerhub.cisco.com/gspie-docker-local
IMAGE=bandwidth-collector
TAG=1.1
PROXY=http://proxy.esl.cisco.com:8080

all: build

build:
	cp -R ../../framework/api/v1/py-client .
	pip install -r requirements.txt

run:
	python network-utilization.py

container:
	cp -R ../../framework/api/v1/py-client .
	docker build --build-arg http_proxy=${PROXY} --build-arg https_proxy=${PROXY} -t ${REPO}/${IMAGE}:${TAG} .

