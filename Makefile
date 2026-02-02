.PHONY: run, test

run:
	@export CONFIG_PATH="./config/local.yaml" && \
		go run ./cmd/url-shortener/main.go

test:
	@go test ./... -v