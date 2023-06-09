CONTAINER_NAME = mongodb
MONGODB_VERSION = 6.0.6
MONGODB_PORT = 27017

COMMIT_HASH := $(shell git rev-parse --short HEAD)
LDFLAGS := -X 'main.commit=${COMMIT_HASH}'

init:
	asdf install
	mkdir -p ./mongo_data/data ./mongo_data/log

up:
	docker run -d \
		--name $(CONTAINER_NAME) \
		-p $(MONGODB_PORT):$(MONGODB_PORT) \
		-v "$(shell pwd)/mongo_data/data/:/data/mongodb" \
		-v "$(shell pwd)/mongo_data/log:/var/log" \
		mongo:$(MONGODB_VERSION) \
		mongod \
		--dbpath /data/mongodb \
		--logpath /var/log/mongodb.log \
		--logappend \
		--vv

down:
	docker stop $(CONTAINER_NAME)
	docker rm $(CONTAINER_NAME)

logs:
	docker logs $(CONTAINER_NAME)

exec:
	docker exec -it $(CONTAINER_NAME) mongosh

fmt:
	go fmt ./...

tidy:
	go mod tidy

update:
	go get -u ./...

install:
	go install -ldflags "$(LDFLAGS)"

test: up
	go test -v -cover ./...

