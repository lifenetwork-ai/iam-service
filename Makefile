.PHONY: human-network-iam-service

build: human-network-iam-service
human-network-iam-service:
	go build -o ./human-network-iam-service ./cmd/main.go
clean:
	rm -i -f human-network-iam-service

restart: stop clean build start
	@echo "human-network-iam-service restarted!"

build-service: clean build
	@echo "Restart service with cmd: 'systemctl restart human-network-iam-service'"
	systemctl restart human-network-iam-service

run: build
	@echo "Starting the human-network-iam-service..."
	@env ./human-network-iam-service &
	@echo "human-network-iam-service running!"

stop:
	@echo "Stopping the human-network-iam-service..."
	@-pkill -SIGTERM -f "human-network-iam-service"
	@echo "Stopped human-network-iam-service"

lint:
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:v1.64.8 golangci-lint run --fix

test:
	go test -v ./...

swagger:
	swag init -g ./cmd/main.go -d ./ -o ./docs

migrate:
	go run cmd/migration/main.go


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

.PHONY: mocks
mocks: clean-mocks
	@echo "Generating mocks..."
	@find . -name "*.go" -not -path "./mocks/*" -not -path "./vendor/*" -exec grep -l "type.*interface" {} \; | \
	while read file; do \
		rel_path=$$(echo $$file | sed 's/\.\///'); \
		dir_path=$$(dirname $$rel_path | sed 's/internal\///'); \
		pkg_name=$$(basename $$dir_path); \
		mock_dir="mocks/$$dir_path"; \
		mkdir -p $$mock_dir; \
		echo "Processing $$file -> $$mock_dir"; \
		mockgen -source="$$file" -package="mock_$$pkg_name" -destination="$$mock_dir/mock_$$(basename $$file)"; \
	done
	@echo "Done generating mocks"

clean-mocks:
	rm -rf mocks/