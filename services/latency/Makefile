all: build


build:
	cp -R ../../framework/api/v1/py-client .
	pip install -r requirements.txt

container:
	cp -R ../../framework/api/v1/py-client .
	docker build -t dockerhub.cisco.com/gspie-dev-docker/latency:latest .
