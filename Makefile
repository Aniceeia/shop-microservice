.PHONY: test test-unit test-integration

up:
	docker-compose -f docker/docker-compose.yml up --build

down:
	docker-compose -f docker/docker-compose.yml down -v

dev-up:
	docker-compose -f docker/docker-compose.dev.yml up --build

dev-down:
	docker-compose -f docker/docker-compose.dev.yml down -v

logs:
	docker-compose -f docker/docker-compose.dev.yml logs -f app

air-local:
	air -c .air.toml

test:
	env $$(cat docker/.env.test | xargs) \
	docker-compose -f docker/docker-compose.dev.yml up -d postgres && \
	sleep 5 && \
	cd internal/infrastructure/postgresql && \
	env $$(cat ../../../../docker/.env.test | xargs) go test -v . -count=1

test-coverage:
	env $$(cat docker/.env.test | xargs) \
	docker-compose -f docker/docker-compose.dev.yml up -d postgres && \
	sleep 5 && \
	cd internal/infrastructure/postgresql && \
	env $$(cat ../../../../docker/.env.test | xargs) go test -v . -coverprofile=coverage.out && \
	go tool cover -html=coverage.out

test-integration:
	# Запуск интеграционных тестов
	env $$(cat docker/.env.test | xargs) \
	docker-compose -f docker/docker-compose.dev.yml up -d postgres && \
	sleep 5 && \
	cd internal/infrastructure/postgresql && \
	env $$(cat ../../../../docker/.env.test | xargs) go test -tags=integration -v . -count=1

test-unit:
	# Unit тесты (без зависимостей)
	cd internal/infrastructure/postgresql && \
	go test -v . -count=1