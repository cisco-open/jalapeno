EXECUTABLE ?= voltron
BINARY ?= bin/$(EXECUTABLE)
REPO=ievoltron
IMAGE=epe-edges
TAG=2.0
PROXY=http://proxy.esl.cisco.com:8080

all: build

build:
	pip install -r requirements.txt

run:
	python create_epe_edges.py

container:
	docker build --build-arg http_proxy=${PROXY} --build-arg https_proxy=${PROXY} -t ${REPO}/${IMAGE}:${TAG} .

