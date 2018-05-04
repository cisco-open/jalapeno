EXECUTABLE ?= voltron
BINARY ?= bin/$(EXECUTABLE)
REPO=dockerhub.cisco.com/gspie-docker-local
IMAGE=bandwidth-collector
TAG=1.0

all: build

build:
	cp -R ../../framework/api/v1/py-client .
	pip install -r requirements.txt

run:
	python network-utilization.py

container:
	cp -R ../../framework/api/v1/py-client .
	docker run -it --rm --name bandwidth-collector-service -v `pwd`\:/services/bandwidth-collector -w /services/bandwidth-collector python\:2 python network_utilization.py
        docker build -t ${REPO}/${IMAGE}:${TAG} .

