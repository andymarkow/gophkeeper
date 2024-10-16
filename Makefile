# Usage:
# make        		# run default command

# To check entire script:
# cat -e -t -v Makefile

.EXPORT_ALL_VARIABLES:

LOG_LEVEL=debug

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

.PHONY: run-postgres
run-postgres:
	docker-compose up -d postgres pgadmin

.PHONY: stop-postgres
stop-postgres:
	docker-compose down postgres pgadmin

.PHONY: run-minio
run-minio:
	docker-compose up -d minio

.PHONY: stop-minio
stop-minio:	
	docker-compose down minio
