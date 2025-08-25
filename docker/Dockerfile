FROM golang:1.24-alpine

WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN go build -o main ./cmd/main.go

# Копируем миграции
COPY migrations ./migrations

# Открываем порт
EXPOSE 8081

# Запускаем приложение
CMD ["./main"]
#ускоряем через следующее выражение : 
# ENV GOCACHE=/root/.cache/go-build
# RUN --mount=type=cache,target="/root/.cache/go-build" CGO_ENABLE=0 GOOS=linux
# GOARCH=amd64 go build -o shop ./cmd/main.go
