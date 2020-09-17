REGISTRY_NAME?=docker.io/sbezverk
IMAGE_VERSION?=0.0.0

.PHONY: all gobmp topology container push clean test

ifdef V
TESTARGS = -v -args -alsologtostderr -v 5
else
TESTARGS =
endif

all: topology

topology:
	mkdir -p bin
	$(MAKE) -C ./cmd/topology compile-topology

topology-container: topology
	docker build -t $(REGISTRY_NAME)/topology:$(IMAGE_VERSION) -f ./build/Dockerfile.topology .

push: topology-container
	docker push $(REGISTRY_NAME)/topology:$(IMAGE_VERSION)

clean:
	rm -rf bin

test:
	GO111MODULE=on go test `go list ./... | grep -v 'vendor'` $(TESTARGS)
	GO111MODULE=on go vet `go list ./... | grep -v vendor`
