REGISTRY_NAME?=docker.io/iejalapeno
IMAGE_VERSION?=latest

.PHONY: all igp-graph container push clean test

ifdef V
TESTARGS = -v -args -alsologtostderr -v 5
else
TESTARGS =
endif

all: igp-graph

igp-graph:
	mkdir -p bin
	$(MAKE) -C ./../../ -f ./Makefile.igp-graph compile-igp-graph

igp-graph-container: igp-graph
	docker build -t $(REGISTRY_NAME)/igp-graph:$(IMAGE_VERSION) -f ./../../build/Dockerfile.igp-graph .

container: igp-graph-container

push: container
	docker push $(REGISTRY_NAME)/igp-graph:$(IMAGE_VERSION)

clean:
	rm -rf bin

test:
	GO111MODULE=on go test `go list ./... | grep -v 'vendor'` $(TESTARGS)

testrace:
	GO111MODULE=on go test -race `go list ./... | grep -v 'vendor'` $(TESTARGS)
