# OrderCount 项目

简单的订单统计服务（Go 后端 + Vue 前端 + MySQL）。

环境变量:
- `MYSQL_DSN`：MySQL 连接字符串，示例: `user:pass@tcp(127.0.0.1:3306)/orders_db?charset=utf8mb4&parseTime=True&loc=Local`

启动后端：
```
cd /home/condingyang/coding/order/ordercount
go run main.go
```

前端：位于 `frontend/`，用 `npm install` 后运行。
