.PHONY: human-network-auth-service

build: human-network-auth-service
human-network-auth-service:
	go build -o ./human-network-auth-service ./cmd/main.go
clean:
	rm -i -f human-network-auth-service

run-test:
	go test -v ./internal/infra/caching/test
	go test -v ./internal/util/test
	go test -v ./test

restart: stop clean build start
	@echo "human-network-auth-service restarted!"

build-service: clean build
	@echo "Restart service with cmd: 'systemctl restart human-network-auth-service'"
	systemctl restart human-network-auth-service

run: build
	@echo "Starting the human-network-auth-service..."
	@env DB_PASSWORD=${DB_PASSWORD} ./human-network-auth-service &
	@echo "human-network-auth-service running!"

stop:
	@echo "Stopping the human-network-auth-service..."
	@-pkill -SIGTERM -f "human-network-auth-service"
	@echo "Stopped human-network-auth-service"

lint:
	golangci-lint run --fix

swagger:
	swag init -g cmd/main.go

wiring: 
	wire ./wire


DB_NAME=human-network-auth
DB_USER=postgres
DB_PASSWORD=postgres
DB_PORT=5432
DB_HOST=localhost


docker-db-up:
	docker run --name secure-genom-db \
		-e POSTGRES_DB=$(DB_NAME) \
		-e POSTGRES_USER=$(DB_USER) \
		-e POSTGRES_PASSWORD=$(DB_PASSWORD) \
		-p $(DB_PORT):5432 \
		-d postgres:12

docker-db-down:
	docker stop secure-genom-db
	docker rm secure-genom-db
