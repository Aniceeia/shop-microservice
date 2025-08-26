FROM golang:1.24-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Ускоряем сборку с кешированием
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go
EXPOSE 8081
CMD ["./main"]

#ускоряем через следующее выражение : 
# ENV GOCACHE=/root/.cache/go-build
# RUN --mount=type=cache,target="/root/.cache/go-build" CGO_ENABLE=0 GOOS=linux
# GOARCH=amd64 go build -o shop ./cmd/main.go
