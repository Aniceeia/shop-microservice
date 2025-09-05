# Shop Microservice

Микросервис для управления заказами с использованием Kafka, PostgreSQL и кэширования.

![](<misc/Screenshot from 2025-09-04 03-26-34.png>)

## Быстрый старт

```bash
# Клонирование репозитория
git clone <repository-url>
cd shop-microservice

# Запуск сервисов
make up

# Остановка сервисов
make down

# Просмотр логов
make logs
```

## Docker Compose

Сервис разворачивается через Docker Compose со следующими компонентами:

```yaml
version: '3.8'
services:
  app:          # Основное приложение на Go
  postgres:     # База данных PostgreSQL
  kafka:        # Брокер сообщений Apache Kafka
  prometheus:   # Мониторинг метрик
  grafana:      # Визуализация метрик
```
![Пример отчета prometheus при нагрузочном тестировании](<misc/Screenshot from 2025-09-04 03-17-48.png>)
## Переменные окружения

Создайте файл `.env` на основе `.env.example`:

```env
APP_PORT=8080
DB_HOST=postgres
DB_USER=user
DB_PASSWORD=pass
DB_NAME=shop
DB_EXTERNAL_PORT=5432
KAFKA_EXTERNAL=9093
KAFKA_MAX=10485760
```

## API Endpoints

- `GET /order/{id}` - Получение информации о заказе
- Метрики доступны на порту 9090 (Prometheus)
- Дашборды Grafana на порту 3000

## Архитектура 
```
graph TB
    %% External Components
    KAFKA[Kafka Broker]
    POSTGRES[(PostgreSQL)]
    WEB[Web Browser]
    PROMETHEUS[Prometheus]
    
    %% Infrastructure Layer
    subgraph "Infrastructure Layer"
        KAFKA_PROD[Kafka Producer]
        KAFKA_CONS[Kafka Consumer]
        KAFKA_MGR[Kafka Manager]
        KAFKA_ADAPTER[Producer Adapter]
        
        PG_REPO[PostgreSQL Repository]
        PG_MIGRATIONS[Migrations]
        
        CACHE_ADAPTER[Cache Adapter]
        
        METRICS[Metrics Collector]
    end
    
    %% Application Layer
    subgraph "Application Layer"
        ORDER_UC[OrderUseCase]
    end
    
    %% API Layer
    subgraph "API Layer"
        HANDLER[HTTP Handler]
        ROUTER[Router]
        MIDDLEWARE[Middleware]
        VALIDATION[Validation]
        METRICS_MW[Metrics Middleware]
    end
    
    %% Domain Layer
    subgraph "Domain Layer"
        MODEL[Order Model]
        REPO_INTF[Repository Interface]
    end
    
    %% DI Layer
    subgraph "Dependency Injection"
        DI_DEPS[Dependencies]
    end
    
    %% Static Files
    subgraph "Static Content"
        STATIC_HTML[Index.html]
    end
    
    %% External Connections
    WEB --> ROUTER
    KAFKA --> KAFKA_CONS
    KAFKA_PROD --> KAFKA
    KAFKA_ADAPTER --> KAFKA
    PG_REPO --> POSTGRES
    METRICS --> PROMETHEUS
    
    %% Internal Flow
    ROUTER --> MIDDLEWARE
    MIDDLEWARE --> VALIDATION
    MIDDLEWARE --> METRICS_MW
    MIDDLEWARE --> HANDLER
    HANDLER --> ORDER_UC
    ORDER_UC --> REPO_INTF
    ORDER_UC --> KAFKA_ADAPTER
    REPO_INTF --> PG_REPO
    REPO_INTF --> CACHE_ADAPTER
    KAFKA_CONS --> ORDER_UC
    
    %% Infrastructure Internal
    KAFKA_MGR --> KAFKA_PROD
    KAFKA_MGR --> KAFKA_CONS
    PG_MIGRATIONS --> PG_REPO
    
    %% Static Content Flow
    ROUTER --> STATIC_HTML
    
    %% DI Connections
    DI_DEPS --> ROUTER
    DI_DEPS --> ORDER_UC
    DI_DEPS --> KAFKA_PROD
    DI_DEPS --> KAFKA_CONS
    DI_DEPS --> KAFKA_MGR
    DI_DEPS --> KAFKA_ADAPTER
    DI_DEPS --> PG_REPO
    DI_DEPS --> CACHE_ADAPTER
    DI_DEPS --> METRICS
    DI_DEPS --> METRICS_MW
    
    %% Domain Contracts
    REPO_INTF -.-> MODEL
    ORDER_UC -.-> MODEL

    %% Styling
    classDef external fill:#e1f5fe,stroke:#01579b;
    classDef infrastructure fill:#f3e5f5,stroke:#4a148c;
    classDef application fill:#e8f5e8,stroke:#1b5e20;
    classDef api fill:#fff3e0,stroke:#e65100;
    classDef domain fill:#ffebee,stroke:#b71c1c;
    classDef di fill:#f1f8e9,stroke:#33691e;
    classDef static fill:#fce4ec,stroke:#880e4f;
    
    class KAFKA,POSTGRES,WEB,PROMETHEUS external;
    class KAFKA_PROD,KAFKA_CONS,KAFKA_MGR,KAFKA_ADAPTER,PG_REPO,PG_MIGRATIONS,CACHE_ADAPTER,METRICS infrastructure;
    class ORDER_UC application;
    class HANDLER,ROUTER,MIDDLEWARE,VALIDATION,METRICS_MW api;
    class MODEL,REPO_INTF domain;
    class DI_DEPS di;
    class STATIC_HTML static;
```

## Тестирование

```bash
# Unit-тесты
make test-unit

# Нагрузочное тестирование
make test-load

# Покрытие кода
make test-coverage

# Все тесты
make test
```

## Нагрузочное тестирование

![](<misc/Screenshot from 2025-09-04 03-19-13.png>)

Нагрузочное тестирование реализовано через команду `make test-load`, которая:

1. **Генерирует тестовые данные** - создает фикстуры для имитации реальной нагрузки
2. **Запускает нагрузочные тесты** - тестирует систему под высокой нагрузкой
3. **Таймаут 10 минут** - достаточно времени для проведения полноценного тестирования

```bash
# Запуск нагрузочного тестирования
make test-load

# Генерация тестовых данных (отдельно)
make generate-test-data

# Удаление тестовых данных (делайте обязательно!)
make clean
```

**Что тестируется:**
- Обработка большого количества сообщений из Kafka
- Производительность записи в PostgreSQL
- Эффективность кэширования
- Стабильность работы под нагрузкой
- Потребление памяти и CPU

## Покрытие кода тестами

Анализ покрытия кода выполняется через `make test-coverage`:

```bash
# Запуск тестов с анализом покрытия
make test-coverage

# Только unit-тесты с покрытием
make test-unit

# Полный цикл тестирования
make test
```

**Функциональность покрытия:**

1. **Генерация отчетов**:
   - `coverage.out` - сырые данные покрытия
   - `coverage.html` - визуальный HTML-отчет
   - Автоматическое открытие отчета в браузере

2. **Анализ конкретных пакетов**:
   ```makefile
   COVERPKG=shop-microservice/internal/infrastructure/cache,shop-microservice/internal/infrastructure/kafka
   ```
   - Фокус на ключевых компонентах инфраструктуры
   - Кэширование и работа с Kafka

3. **Метрики покрытия**:
   - Процент покрытия по функциям
   - Детализация по строкам кода
   - Визуализация непокрытых участков

**Пример вывода:**
```
coverage: 78.5% of statements
shop-microservice/internal/infrastructure/cache 85.2%
shop-microservice/internal/infrastructure/kafka 72.1%
```

## Структура тестов

```
tests/
├── unit/           # Юнит-тесты
├── load/           # Нагрузочные тесты
├── generate_data.go # Генератор тестовых данных
└── fixtures/       # Тестовые фикстуры
```

**Особенности реализации:**

- **Теги компиляции**: `-tags=load` для нагрузочных тестов
- **Изоляция**: Тесты не влияют на продакшен базу
- **Воспроизводимость**: Генерация данных гарантирует повторяемость тестов
- **CI-дружественность**: Все тесты могут запускаться в пайплайнах


## Трудности реализации

1. **Проблемы с Kafka**: InvalidReceiveException при больших сообщениях
   - Решение: Сброс volume и согласование лимитов между клиентом и брокером через docker-compose и непосредственно в коде
   - Разбивка больших сообщений на чанки

2. **Архитектурные решения**: Выбор между интерфейсами в domain/usecase слоях
   - Решение: Интерфейсы в domain слое для соблюдения Clean Architecture

3. **Производительность БД**: N+1 запрос при получении заказов с товарами
   - Решение: Оптимизация через batch-запросы и джойны

4. **Утечки горутин**: Риск при высоких нагрузках в Producer
   - Решение: Реализация worker pool и буферизованных каналов

5. **Безопасность**: Инкапсуляция переменных окружения в .env файл

6. **Мониторинг**: Интеграция Prometheus+Grafana для грамотного сбора метрик