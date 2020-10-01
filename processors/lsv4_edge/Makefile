REGISTRY_NAME?=docker.io/sbezverk
IMAGE_VERSION?=0.0.0

.PHONY: all gobmp lsv4_edge container push clean test

ifdef V
TESTARGS = -v -args -alsologtostderr -v 5
else
TESTARGS =
endif

all: lsv4_edge

lsv4_edge:
	mkdir -p bin
	$(MAKE) -C ./cmd/lsv4_edge compile-lsv4_edge

lsv4_edge-container: lsv4_edge
	docker build -t $(REGISTRY_NAME)/lsv4_edge:$(IMAGE_VERSION) -f ./build/Dockerfile.lsv4_edge .

push: lsv4_edge-container
	docker push $(REGISTRY_NAME)/lsv4_edge:$(IMAGE_VERSION)

clean:
	rm -rf bin

test:
	GO111MODULE=on go test `go list ./... | grep -v 'vendor'` $(TESTARGS)
	GO111MODULE=on go vet `go list ./... | grep -v vendor`
