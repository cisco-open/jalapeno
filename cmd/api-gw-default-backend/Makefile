compile-default-backend:
	docker build -t build-default-backend .
	docker container create --name temp build-default-backend
	docker container cp temp:/build/api-gw-default-backend ../../bin
	docker container rm temp
