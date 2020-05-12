EXECUTABLE ?= jalapeno
BINARY ?= bin/$(EXECUTABLE)
REPO=iejalapeno
IMAGE=ls-processor
TAG=1.2
# If you connect to dockerhub via a proxy: uncomment and edit the following line (and see below docker build lines)
#PROXY=http://proxy.esl.cisco.com:8080

all: build

build:
	pip3 install -r requirements.txt

run:
	python3.6 create_ls_topology.py

container:
	# Default docker build without proxy setting:
	docker build -t ${REPO}/${IMAGE}:${TAG} .

	# Uncomment this line (and comment out the above line) if you connect to dockerhub via a proxy
	#docker build --build-arg http_proxy=${PROXY} --build-arg https_proxy=${PROXY} -t ${REPO}/${IMAGE}:${TAG} .

