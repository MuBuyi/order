# 多合一镜像：构建前端并编译 Go 后端，统一由 Go 程序提供接口和静态页面
FROM golang:1.22-alpine

# 安装 Node.js 和 npm 用于前端构建
RUN apk add --no-cache nodejs npm git

WORKDIR /app

# 预先拷贝 go.mod/go.sum，提升依赖缓存利用率
COPY go.mod go.sum ./
RUN go mod download

# 拷贝剩余代码
COPY . .

# 构建前端
WORKDIR /app/frontend
RUN npm install && npm run build

# 回到后端根目录并构建二进制
WORKDIR /app
RUN go build -o ordercount .

EXPOSE 8080

ENV GIN_MODE=release

CMD ["./ordercount"]
