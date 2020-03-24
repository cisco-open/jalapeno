EXECUTABLE ?= jalapeno
BINARY ?= bin/$(EXECUTABLE)
REPO=iejalapeno
IMAGE=l3vpn-processor
TAG=0.1
# If your cluster is behind and proxy: uncomment edit the following line to allow a proxied connection to dockerhub.
#PROXY=http://proxy.esl.cisco.com:8080

all: build

build:
	pip3 install -r requirements.txt

run:
	python3.6 create_l3vpn_topology.py

container:
	# Default docker build without proxy setting:
	docker build -t ${REPO}/${IMAGE}:${TAG} .	

	# Uncomment this line (and comment out the above line) if you connect to dockerhub via a proxy
	#docker build --build-arg http_proxy=${PROXY} --build-arg https_proxy=${PROXY} -t ${REPO}/${IMAGE}:${TAG} .
	

