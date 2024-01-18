all: test build

build:
	@echo '**********' Building binary...
	go build -o myapp cmd/local/main.go
	@echo

test:
	@echo '**********' Running tests...
	go test -v ./...
	@echo

gen-swagger:
	swagger generate spec -o ./resources/swagger.json --scan-models

swagger: gen-swagger
	swagger serve --port=8081 -F=swagger resources/swagger.json
