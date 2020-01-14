EXECUTABLE ?= voltron
BINARY ?= bin/$(EXECUTABLE)
REPO=ievoltron
IMAGE=internal-links-performance-collector
TAG=3.0
PROXY=http://proxy.esl.cisco.com:8080

all: build

build:
	pip install -r requirements.txt

run:
	python calculate_internal_links_performance.py

container:
	docker build --build-arg http_proxy=${PROXY} --build-arg https_proxy=${PROXY} -t ${REPO}/${IMAGE}:${TAG} .

