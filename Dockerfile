# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/server ./server
# 生产请挂载 config/config.yaml 或 compose 卷覆盖，勿使用示例中的开发密钥
COPY config/config.example.yaml ./config/config.yaml

EXPOSE 8080

CMD ["./server"]
