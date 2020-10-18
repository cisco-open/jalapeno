IMAGE_REGISTRY=iejalapeno
IMAGE_NAME=gobmp-topology
IMAGE_TAG=1.6.3

.PHONY: all topology topology-container topology-push

all: topology

topology:
	mkdir -p bin
	$(MAKE) -C ./cmd compile-topology

topology-container: topology
	docker build --no-cache -t $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG) -f ./build/Dockerfile .

topology-push: topology-container
	docker push $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)


