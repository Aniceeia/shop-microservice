FROM golang:1.24-alpine
WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata curl
RUN go install github.com/air-verse/air@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN mkdir -p tmp
COPY internal/static /app/internal/static

EXPOSE 8081 9000
CMD ["./tmp/main"]