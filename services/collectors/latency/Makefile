EXECUTABLE ?= voltron
BINARY ?= bin/$(EXECUTABLE)
REPO=dockerhub.cisco.com/gspie-docker-local
IMAGE=latency
TAG=1.5

all: build

build:
	cp -R ../../framework/api/v1/py-client .
	pip install -r requirements.txt

run:
	python create_paths.py

container:
	cp -R ../../framework/api/v1/py-client .
	docker run -it --rm --name latency-service -v `pwd`\:/latency/path -w /latency/path python\:2 python network_latency.py
        docker build -t ${REPO}/${IMAGE}:${TAG} .

