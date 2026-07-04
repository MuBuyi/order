# OrderCount 项目

简单的订单统计服务（Go 后端 + Vue 前端 + MySQL）。

环境变量:
## 本地开发
- `MYSQL_DSN`：MySQL 连接字符串，如果不设置则从 `config.yaml` 中读取。

启动后端：
```
cd /home/condingyang/coding/order/ordercount
go run main.go
```

前端：位于 `frontend/`，用 `npm install` 后运行。
```
cd frontend
npm install
npm run dev
```

## Docker 一键部署（推荐用于公网环境）

仓库根目录已经提供 `docker-compose.yml` 和后端的 `Dockerfile`，可以在装有 Docker / Docker Compose 的服务器上一键启动：

1. 克隆仓库并进入目录：
```
git clone <your-repo-url> order
cd order
```

2. 直接启动：
```
docker compose up -d
```

这会启动：
- 一个 MySQL 容器（服务名 `db`，数据库名 `ordercount`，用户 `appuser/app123456`）。
- 一个 Go 后端容器（服务名 `backend`），对外映射 `8081` 端口（`http://<服务器IP>:8081`）。
- 一个前端 Nginx 容器（服务名 `frontend`），对外映射 `8080` 端口（`http://<服务器IP>:8080/`）。

后端容器通过环境变量 `MYSQL_DSN` 连接到 `db` 容器，形如：
```
appuser:app123456@tcp(db:3306)/ordercount?charset=utf8mb4&parseTime=True&loc=Local
```

前端容器通过内置 Nginx 将 `/api` 与 `/uploads` 反向代理到 `backend:8080`，因此浏览器统一访问 `http://<服务器IP>:8080` 即可。

如果在服务器上需要配置企业微信 / 豆包等，可在 `docker-compose.yml` 的 `backend.environment` 中补充：
- `WECHAT_ROBOT_WEBHOOK`
- `DOUBAO_API_KEY`
- `DOUBAO_ENDPOINT`
- `DOUBAO_MODEL`
