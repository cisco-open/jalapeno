EXECUTABLE ?= voltron
BINARY ?= bin/$(EXECUTABLE)
REPO=dockerhub.cisco.com/gspie-docker-local
IMAGE=telemetry
TAG=2.0

all: build

build:
	pip install -r requirements.txt

run:
	python deploy_telemetry.py

container:
	docker run -it --rm --name telemetry -v `pwd`\:/infra/telemetry -w /infra/telemetry python\:3.6 python deploy_telemetry.py
        docker build -t ${REPO}/${IMAGE}:${TAG} .

