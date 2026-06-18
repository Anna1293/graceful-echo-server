.PHONY: run test test-race vet build docker docker-down certs certs-sh lint

run:
	go run .

test:
	go test ./...

test-race:
	go test -race ./...

vet:
	go vet ./...

build:
	go build -o bin/graceful-echo-server .

docker:
	docker compose up --build

docker-down:
	docker compose down

certs:
	powershell -ExecutionPolicy Bypass -File scripts/generate-certs.ps1

certs-sh:
	sh scripts/generate-certs.sh

lint:
	golangci-lint run ./...
