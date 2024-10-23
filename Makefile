# Usage:
# make        		# run default command

# To check entire script:
# cat -e -t -v Makefile

.EXPORT_ALL_VARIABLES:

LOG_LEVEL=debug
KEEPER_S3_ACCESS_KEY=minioadmin
KEEPER_S3_SECRET_KEY=minioadmin
KEEPER_DATABASE_DSN=postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable

.PHONY: all
all: fmt tidy vet test lint

.PHONY: fmt
fmt:
	go fmt ./...
	$(HOME)/go/bin/goimports -l -w --local "github.com/andymarkow/gophkeeper" .

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint:
	$(HOME)/go/bin/staticcheck -fail "" ./...
	docker run --rm --name golangci-lint -v `pwd`:/workspace -w /workspace \
		golangci/golangci-lint:latest-alpine golangci-lint run --issues-exit-code 1

.PHONY: test
test:
	go clean -testcache
	go test -race -v ./...

.PHONY: coverage
coverage:
	go clean -testcache
	go test -v -cover -coverprofile=.coverage.cov ./...
	go tool cover -func=.coverage.cov
	go tool cover -html=.coverage.cov
	rm .coverage.cov

.PHONY: run-server
run-server:
	go run ./cmd/server

.PHONY: postgres-start
postgres-start:
	docker-compose up -d postgres

.PHONY: postgres-stop
postgres-stop:
	docker-compose down postgres

.PHONY: minio-start
minio-start:
	docker-compose up -d minio

.PHONY: minio-stop
minio-stop:	
	docker-compose down minio

.PHONY: start-all
start-all: postgres-start minio-start

.PHONY: stop-all
stop-all: postgres-stop minio-stop

.PHONY: clean
clean:
	docker system prune -f
