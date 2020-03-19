EXECUTABLE ?= jalapeno
BINARY ?= bin/$(EXECUTABLE)
REPO=iejalapeno
IMAGE=ls-processor
TAG=0.1
PROXY=http://proxy.esl.cisco.com:8080

all: build

build:
	pip3 install -r requirements.txt

run:
	python3.6 create_ls_topology.py

container:
	docker build --build-arg http_proxy=${PROXY} --build-arg https_proxy=${PROXY} -t ${REPO}/${IMAGE}:${TAG} .

