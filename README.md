# Shop Microservice

Микросервис для управления заказами с использованием Kafka, PostgreSQL и кеширования.

## Архитектура

Сервис построен по принципам чистой архитектуры:

```
internal/
├── api/           # HTTP API слой
├── application/   # Бизнес-логика (usecases)
├── domain/        # Доменные модели и интерфейсы
├── infrastructure/# Внешние зависимости (БД, Kafka, кеш)
└── di/           # Dependency injection контейнер
```

### Основные компоненты

- **Kafka Producer/Consumer**: Асинхронная обработка заказов
- **PostgreSQL Repository**: Хранение данных с retry механизмом
- **In-Memory Cache**: Кеш с TTL и ограничением размера
- **HTTP API**: REST endpoints для работы с заказами

## Возможности

- ✅ Создание заказов через API
- ✅ Получение заказа по ID
- ✅ Получение всех заказов
- ✅ Кеширование с TTL и ограничением размера
- ✅ Retry механизм для БД и Kafka
- ✅ Graceful shutdown горутин
- ✅ Swagger документация
- ✅ Unit тесты с покрытием >80%

## Запуск

### Предварительные требования

- Docker и Docker Compose
- Go 1.24+

### Быстрый старт

```bash
# Запуск сервисов
make up
# Очистка контейнеров
make down
```


## API Endpoints

- `POST /api/orders` - Создание заказа
- `GET /api/orders/:id` - Получение заказа по ID
- `GET /api/orders` - Получение всех заказов
- `GET /api/health` - Проверка здоровья сервиса
- `GET /api/test` - Системные тесты

## Тестирование

### Unit тесты
```bash
make test-unit
```

### Покрытие кода
```bash
make test-coverage
```
Откроется HTML страница с детальным покрытием.

### Нагрузочное тестирование
```bash
make test-load
```

## Swagger

После установки зависимостей:
```bash
make swagger
```

Swagger UI будет доступен по адресу: `http://localhost:8081/swagger/`

## Оптимизации

### Кеш
- TTL: 30 минут
- Максимальный размер: 1000 записей
- Автоматическая очистка устаревших записей
- LRU eviction при переполнении

### Retry механизмы
- **PostgreSQL**: 3 попытки с exponential backoff
- **Kafka**: 3 попытки с exponential backoff
- **Base delay**: 100ms

### Горутины
- Graceful shutdown с таймаутом
- Корректное завершение worker'ов
- Тестирование корректности завершения

## Мониторинг

- Health check endpoint
- Логирование ошибок с retry
- Метрики размера кеша
- Статус подключений к внешним сервисам
