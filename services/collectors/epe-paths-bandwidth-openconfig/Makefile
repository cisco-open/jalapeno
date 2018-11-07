EXECUTABLE ?= voltron
BINARY ?= bin/$(EXECUTABLE)
REPO=ievoltron
IMAGE=epe-paths-bandwidth-openconfig-collector
TAG=1.0
PROXY=http://proxy.esl.cisco.com:8080

all: build

build:
	pip install -r requirements.txt

run:
	python calculate_epe_paths_bandwidth_openconfig.py

container:
	docker build --build-arg http_proxy=${PROXY} --build-arg https_proxy=${PROXY} -t ${REPO}/${IMAGE}:${TAG} .

