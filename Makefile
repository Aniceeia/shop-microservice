.PHONY: test test-unit test-integration


up:
	docker-compose -f docker/docker-compose.yml up --build

down:
	docker-compose -f docker/docker-compose.yml down -v

logs:
	docker-compose -f docker/docker-compose.yml logs -f app

air-local:
	air -c .air.toml

test:
	docker-compose -f docker/docker-compose.yml up -d postgres && \
	sleep 10 && \
	cd internal/infrastructure/postgresql && \
	DB_HOST=localhost DB_PORT=5433 DB_USER=orders_user DB_PASSWORD=orders_password DB_NAME=orders_db go test -v . -count=1

test-coverage:
	docker-compose -f docker/docker-compose.yml up -d postgres && \
	sleep 5 && \
	cd internal/infrastructure/postgresql && \
	env $$(cat ../../../../docker/.env | xargs) go test -v . -coverprofile=coverage.out && \
	go tool cover -html=coverage.out

test-integration:
	# Запуск интеграционных тестов
	env $$(cat docker/.env | xargs) \
	docker-compose -f docker/docker-compose.yml up -d postgres && \
	sleep 5 && \
	cd internal/infrastructure/postgresql && \
	env $$(cat ../../../docker/.env | xargs) go test -tags=integration -v . -count=1

test-unit:
	# Unit тесты (без зависимостей)
	cd internal/infrastructure/postgresql && \
	go test -v . -count=1