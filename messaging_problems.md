Предложенный вариант с интерфейсом в domain слое - это **отличное решение** и соответствует лучшим практикам чистой архитектуры. Но есть и альтернативные варианты. Давайте рассмотрим их все:

## Вариант 1: Интерфейс в domain слое (рекомендуемый)

```go
// internal/domain/ports/message_producer.go
package ports

import (
	"context"
	"shop-microservice/internal/domain/model"
)

type MessageProducer interface {
	ProduceOrder(ctx context.Context, order *model.Order) error
}
```

**Плюсы:**
- ✅ Полное соблюдение Clean Architecture
- ✅ Usecase зависит от абстракции (domain порта)
- ✅ Легко тестировать (mock интерфейса)
- ✅ Легко заменить реализацию (Kafka → RabbitMQ → etc)

**Минусы:**
- ❌ Нужно создавать дополнительный файл/пакет

## Вариант 2: Интерфейс в usecase слое

```go
// internal/application/usecases/message_producer.go
package usecases

import (
	"context"
	"shop-microservice/internal/domain/model"
)

type MessageProducer interface {
	ProduceOrder(ctx context.Context, order *model.Order) error
}
```

**Плюсы:**
- ✅ Проще (меньше файлов)
- ✅ Все еще соблюдается Dependency Inversion

**Минусы:**
- ❌ Интерфейс привязан к usecase, а не к domain
- ❌ Сложнее переиспользовать в других usecases

## Вариант 3: Функциональный подход (без интерфейса)

```go
// internal/application/usecases/order_usecase.go
type OrderUseCase struct {
	repo           repositories.OrderRepository
	produceOrderFn func(ctx context.Context, order *model.Order) error
	// ...
}

func NewOrderUseCase(
	repo repositories.OrderRepository,
	produceOrderFn func(ctx context.Context, order *model.Order) error,
	// ...
) OrderUseCase {
	// ...
}
```

**Плюсы:**
- ✅ Максимальная гибкость
- ✅ Очень легко тестировать

**Минусы:**
- ❌ Менее читаемо
- ❌ Сложнее документировать

## Вариант 4: Dependency Injection через методы

```go
// internal/application/usecases/order_usecase.go
type OrderUseCase struct {
	repo repositories.OrderRepository
	// ...
}

func (uc *OrderUseCase) SetMessageProducer(producer func(ctx context.Context, order *model.Order) error) {
	// ...
}
```

**Плюсы:**
- ✅ Гибкая настройка

**Минусы:**
- ❌ Неявные зависимости
- ❌ Возможность забыть установить зависимость

## Мой вердикт

**Вариант 1 (интерфейс в domain слое) - лучший выбор** потому что:

1. **Соответствует DIP** (Dependency Inversion Principle)
2. **Четкое разделение слоев** - domain знает "что", infrastructure знает "как"
3. **Легко тестировать** - можно мокировать интерфейс
4. **Гибкость** - завтра сможете легко сменить брокер сообщений
5. **Профессиональный подход** - используется в крупных проектах

## Пример реализации с Вариантом 1:

**Domain порт:**
```go
// internal/domain/ports/message_producer.go
package ports

import (
	"context"
	"shop-microservice/internal/domain/model"
)

type OrderProducer interface {
	ProduceOrder(ctx context.Context, order *model.Order) error
}
```

**Infrastructure реализация:**
```go
// internal/infrastructure/kafka/adapter.go
package kafka

import (
	"context"
	"shop-microservice/internal/domain/model"
	"shop-microservice/internal/domain/ports"
)

type KafkaOrderProducer struct {
	producer *Producer
}

func NewKafkaOrderProducer(producer *Producer) ports.OrderProducer {
	return &KafkaOrderProducer{producer: producer}
}

func (k *KafkaOrderProducer) ProduceOrder(ctx context.Context, order *model.Order) error {
	return k.producer.Produce(ctx, order.OrderUID, order)
}
```

**DI контейнер:**
```go
// internal/di/container.go
kafkaProducer := kafka.NewProducer(...)
orderProducer := kafka.NewKafkaOrderProducer(kafkaProducer)

orderUseCase := usecases.NewOrderUseCase(
    orderRepo,
    orderProducer, // domain порт
    workers,
    bufferSize,
)
```

**Это профессиональный подход, который окупится при масштабировании проекта!** 🚀