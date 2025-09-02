REGISTRY_NAME?=docker.io/iejalapeno
IMAGE_VERSION?=latest

.PHONY: all ip-graph container push clean test

ifdef V
TESTARGS = -v -args -alsologtostderr -v 5
else
TESTARGS =
endif

all: ip-graph

ip-graph:
	mkdir -p bin
	$(MAKE) compile-ip-graph

compile-ip-graph:
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o bin/ip-graph .

test:
	GO111MODULE=on go test `go list ./... | grep -v vendor/` $(TESTARGS)

container: ip-graph
	docker build -t $(REGISTRY_NAME)/ip-graph:$(IMAGE_VERSION) -f ./build/Dockerfile.ip-graph .

push: container
	docker push $(REGISTRY_NAME)/ip-graph:$(IMAGE_VERSION)

clean:
	rm -rf bin

.PHONY: install-deps
install-deps:
	GO111MODULE=on go mod download
