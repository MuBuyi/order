# 后端镜像（仅 Go API 服务）
FROM golang:1.22-alpine AS builder

WORKDIR /src

# 预先拷贝依赖定义，提升构建缓存命中
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/ordercount .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/ordercount /app/ordercount
COPY config.yaml /app/config.yaml

# 上传目录可通过 compose volume 挂载覆盖
RUN mkdir -p /app/uploads/products

EXPOSE 8080

ENV GIN_MODE=release

CMD ["/app/ordercount"]
