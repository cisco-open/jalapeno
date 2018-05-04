EXECUTABLE ?= voltron
BINARY ?= bin/$(EXECUTABLE)
REPO=dockerhub.cisco.com/gspie-docker-local
IMAGE=latency-collector
TAG=1.5

all: build

build:
	cp -R ../../framework/api/v1/py-client .
	pip install -r requirements.txt

run:
	python network-latency.py

container:
	cp -R ../../framework/api/v1/py-client .
	docker run -it --rm --name latency-collector-service -v `pwd`\:/services/latency -w /services/latency python\:2 python network_latency.py
        docker build -t ${REPO}/${IMAGE}:${TAG} .

