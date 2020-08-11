IMAGE_REGISTRY=iejalapeno
IMAGE_NAME=l3vpn_rt
IMAGE_TAG=1.0

.PHONY: all l3vpn_rt l3vpn_rt-container l3vpn_rt-push

all: l3vpn_rt

l3vpn_rt:
	mkdir -p bin
	$(MAKE) -C ./cmd compile-l3vpn_rt

l3vpn_rt-container: l3vpn_rt
	docker build --no-cache -t $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG) -f ./build/Dockerfile .

l3vpn_rt-push: l3vpn_rt-container
	docker push $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)


