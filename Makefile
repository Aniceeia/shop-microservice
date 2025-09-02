.PHONY: test test-unit test-load test-coverage generate-test-data

PACKAGES := $(shell go list ./... | grep -v \/tests | grep -v \/internal\/infrastructure\/kafka | grep -v \/internal\/infrastructure\/postgresql | grep -v \/internal\/di)
COVERPKG := $(shell echo $(PACKAGES) | tr ' ' ',')

generate-test-data:
	go run tests/generate_data.go

test-unit:
	go test ./tests/unit/... -v -coverpkg=$(COVERPKG) -coverprofile=coverage-unit.out

test-load: generate-test-data
	go test ./tests/load/... -v -tags=load -timeout=10m

test-coverage:
	go test ./tests/unit -coverpkg=$(COVERPKG) -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out
	-xdg-open coverage.html >/dev/null 2>&1 || true

test: test-unit  test-coverage

# Frontend test button (add to your index.html)
frontend-test:
	@echo "Add test button to frontend:"
	@echo '<button onclick="runTests()">Run Tests</button>'
	@echo '<script>function runTests() { fetch("/api/test").then(r => r.json()).then(console.log) }</script>'

up:
	docker-compose -f docker/docker-compose.yml up --build

down:
	docker-compose -f docker/docker-compose.yml down -v

logs:
	docker-compose -f docker/docker-compose.yml logs -f app