IMAGE_REGISTRY=iejalapeno
IMAGE_NAME=gobmp-topology
IMAGE_TAG=1.0

.PHONY: all topology topology-container topology-push

all: topology

topology:
        mkdir -p bin
        $(MAKE) -C ./cmd/topology compile-topology

topology-container: topology
        docker build -t $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG) -f ./build/Dockerfile .

topology-push: topology-container
        docker push $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
