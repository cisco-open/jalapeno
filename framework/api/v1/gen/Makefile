.PHONY: .git python out client server

THISDIR = $(shell pwd)

all: client python
client:
	docker run -v $(THISDIR):/swag \
	          jimschubert/swagger-codegen-cli:2.2.2 generate \
	          -i /swag/v1.yaml \
	          -l go \
	          -t /swag/templates/client/ \
	          -c /swag/clientconfig.json \
	          -o /swag/out/api/client
	cd $(THISDIR)/out/api/client && go fmt .

python:
	docker run -v $(THISDIR):/swag \
						jimschubert/swagger-codegen-cli:2.2.2 generate \
	          -i /swag/v1.yaml \
	          -l python \
	          -c /swag/clientconfig.json \
	          -o /swag/out/python/api/client

html:
	docker run -v $(THISDIR):/swag \
          jimschubert/swagger-codegen-cli:2.2.2 generate \
          -i /swag/v1.yaml \
          -l html \
          -o /swag/

server:
	docker run -v $(THISDIR):/swag \
						jimschubert/swagger-codegen-cli:2.2.2 generate \
						-i /swag/v1.yaml \
						-l go-server \
						-t /swag/templates/server/ \
						-c /swag/serverconfig.json \
						-o /swag/out/api/server
	cd $(THISDIR)/out/api/server/go && go fmt .
