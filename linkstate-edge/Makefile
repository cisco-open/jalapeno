REGISTRY_NAME?=docker.io/sbezverk
IMAGE_VERSION?=latest

.PHONY: all lslinknode-edge container push clean test

ifdef V
TESTARGS = -v -args -alsologtostderr -v 5
else
TESTARGS =
endif

all: lslinknode-edge

lslinknode-edge:
	mkdir -p bin
	$(MAKE) -C ./cmd compile-lslinknode-edge

lslinknode-edge-container: lslinknode-edge
	docker build -t $(REGISTRY_NAME)/lslinknode-edge:$(IMAGE_VERSION) -f ./build/Dockerfile.lslinknode-edge .

push: lslinknode-edge-container
	docker push $(REGISTRY_NAME)/lslinknode-edge:$(IMAGE_VERSION)

clean:
	rm -rf bin

test:
	GO111MODULE=on go test `go list ./... | grep -v 'vendor'` $(TESTARGS)
	GO111MODULE=on go vet `go list ./... | grep -v vendor`
