all: test build

build:
	@echo '**********' Building binary...
	go build -o ../myapp cmd/local/main.go
	@echo

test:
	@echo '**********' Running tests...
	go test -v ./...
	@echo

