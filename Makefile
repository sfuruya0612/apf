CONTAINER_NAME = mongodb
MONGODB_VERSION = 4.4.6
MONGODB_PORT = 27017

init:
	asdf install
	mkdir -p ./mongo/data ./mongo/log

up:
	docker run -d \
		--name $(CONTAINER_NAME) \
		-p $(MONGODB_PORT):$(MONGODB_PORT) \
		-v "$(shell pwd)/mongo/data/:/data/mongodb" \
		-v "$(shell pwd)/mongo/log:/var/log" \
		mongo:$(MONGODB_VERSION) \
		mongod \
		--dbpath /data/mongodb \
		--logpath /var/log/mongodb.log \
		--logappend \
		--vv

down:
	docker stop $(CONTAINER_NAME)

logs:
	docker logs $(CONTAINER_NAME)

exec:
	docker exec -it $(CONTAINER_NAME) mongo

fmt:
	go fmt ./...

tidy:
	go mod tidy

install:
	go install ./...

test: up
	go test ./...

