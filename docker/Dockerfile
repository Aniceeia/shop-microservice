FROM golang:1.24-alpine as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o shop ./cmd/main.go

FROM scratch
COPY --from=builder /app/shop /
USER 1001:1001
EXPOSE 8081
CMD ["/shop"]
#ускоряем через следующее выражение : 
# ENV GOCACHE=/root/.cache/go-build
# RUN --mount=type=cache,target="/root/.cache/go-build" CGO_ENABLE=0 GOOS=linux
# GOARCH=amd64 go build -o shop ./cmd/main.go
