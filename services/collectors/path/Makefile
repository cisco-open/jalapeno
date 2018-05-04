EXECUTABLE ?= voltron
BINARY ?= bin/$(EXECUTABLE)
REPO=dockerhub.cisco.com/gspie-docker-local
IMAGE=path
TAG=1.0

all: build

build:
	cp -R ../../framework/api/v1/py-client .
	pip install -r requirements.txt

run:
	python create_paths.py

container:
	cp -R ../../framework/api/v1/py-client .
	docker run -it --rm --name paths-service -v `pwd`\:/services/path -w /services/path python\:3 python create_paths.py
        docker build -t ${REPO}/${IMAGE}:${TAG} .

