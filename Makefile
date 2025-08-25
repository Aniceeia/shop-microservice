.PHONY: test test-unit test-integration

test-unit:
	go test ./internal/... -v -tags=unit

test-integration:
	go test ./internal/... -v -tags=integration

test: test-unit

run:
	go run cmd/main.go

migrate:
	go run cmd/migrate.go