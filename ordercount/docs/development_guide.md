# 订单统计管理后台 开发说明（新手友好版）

> 本文是给第一次接触本项目的开发者看的，从整体架构开始，一步一步讲清楚：后端从入口到数据库的完整流程、前端页面和接口的对应关系。

---

## 一、整体架构概览

- **后端技术栈**：Go + Gin + GORM + MySQL
  - 主要代码位置：`main.go`、`internal/db`、`internal/handlers`、`internal/models`、`internal/utils`。
- **前端技术栈**：Vue 3 + Vite + Element Plus + ECharts
  - 主要代码位置：`frontend/src/main.js`、`frontend/src/App.vue`、`frontend/src/components/*`。
- **功能模块**：
  - 订单录入与统计（订单列表、今日销售额、今日货款成本、走势图、商品排行）
  - 结账工具（按天计算利润、保存每日结算记录、广告费折算趋势）
  - 商品管理（SKU、名称、成本、图片）
  - 用户与角色管理（超级管理员 / 管理员 / 员工，页面权限控制）
  - 实时汇率展示与结算中自动带入汇率

下面从后端入口开始，一步一步说明。

---

## 二、后端：从入口到数据库

### 1. 入口 main.go

文件：`main.go`

主要做三件事：

1. 读取配置，初始化数据库：
   - 从 `config.yaml` 读取 MySQL DSN，如果没有则读环境变量 `MYSQL_DSN`：
   - 调用 `db.InitDB(dsn)` 连接 MySQL 并自动迁移部分表。
2. 创建 Gin 引擎，挂载所有 `/api/...` 接口。
3. 配置前端静态资源路径，监听 `:8080` 端口。

核心结构（简化）：

```go
func main() {
    dsn := 从 config.yaml 或 MYSQL_DSN 读取
    gdb, _ := db.InitDB(dsn)

    r := gin.Default()
    api := r.Group("/api")

    // 登录、用户信息
    api.POST("/login", handlers.Login(gdb))
    authGroup := api.Group("").Use(handlers.AuthMiddleware())
    authGroup.GET("/me", handlers.Me())

    // 用户管理（only superadmin）
    users := api.Group("/users").Use(handlers.AuthMiddleware(), handlers.RequireRole("superadmin"))
    users.GET("", handlers.ListUsers(gdb))
    users.POST("", handlers.CreateUser(gdb))
    users.PUT("/:id/role", handlers.UpdateUserRole(gdb))
    users.PUT("/:id/permissions", handlers.UpdateUserPermissions(gdb))
    users.PUT("/:id/password", handlers.UpdateUserPassword(gdb))

    // 订单录入和统计
    api.POST("/order", handlers.PostOrder(gdb))
    api.GET("/orders", handlers.ListOrders(gdb))
    api.PUT("/orders/:id", handlers.UpdateOrder(gdb))
    api.PUT("/orders/:id/date", handlers.UpdateOrderDate(gdb))
    api.DELETE("/orders/:id", handlers.DeleteOrder(gdb))

    // 每日销售额 / 成本 / 导出
    api.GET("/sales/today", handlers.TodaySales(gdb))
    api.GET("/costs/today", handlers.TodayGoodsCost(gdb))
    api.GET("/report/export", handlers.ExportReport(gdb))

    // 订单统计图表
    api.GET("/stats/sales-trend", handlers.SalesTrend(gdb))
    api.GET("/stats/top-products", handlers.TopProducts(gdb))
    api.GET("/stats/hourly", handlers.HourlyStats(gdb))
    api.GET("/stats/daily", handlers.DailyStats(gdb))
    api.GET("/stats/monthly", handlers.MonthlyStats(gdb))

    // 结算记录 & 广告费折算统计
    api.POST("/settlement", handlers.SaveSettlement(gdb))
    api.GET("/settlements", handlers.ListSettlements(gdb))
    api.GET("/stats/ad-deduction/daily", handlers.AdDeductionDailyStats(gdb))
    api.GET("/stats/ad-deduction/monthly", handlers.AdDeductionMonthlyStats(gdb))

    // 商品管理
    products := api.Group("/products").Use(handlers.AuthMiddleware())
    products.GET("", handlers.ListProducts(gdb))
    products.POST("", handlers.SaveProduct(gdb))
    products.DELETE(":id", handlers.DeleteProduct(gdb))
    products.POST("/upload", handlers.UploadProductImage())

    // 汇率
    api.GET("/exchange/rates", handlers.ExchangeRates)

    // 前端静态资源
    r.StaticFile("/", "frontend/dist/index.html")
    r.Static("/assets", "frontend/dist/assets")
    r.StaticFile("/vite.svg", "frontend/dist/vite.svg")
    r.Static("/uploads", "uploads")

    r.Run(":8080")
}
```

### 2. 数据库初始化 db.InitDB

文件：`internal/db/db.go`

- 使用 GORM 连接 MySQL：
  - `gorm.Open(mysql.Open(dsn), &gorm.Config{})`
- 自动迁移表：`Order`、`DailySettlement`、`Product`。
- 首次启动时，如果用户表没有数据，会自动创建：
  - 超级管理员：`root / root123`
  - 管理员：`admin / admin123`

这就是为什么你第一次跑项目时，可以用 root / admin 登录。

### 3. 模型层 models

位于 `internal/models`：

- `order.go`：订单表结构 `Order`
  - 包含：国家、平台、订单号、商品名、SKU、数量、总额、币种、创建时间。
- `product.go`：商品表 `Product`
  - SKU、名称、图片地址、三种角色的成本（基础 / 管理员 / 员工）。
- `settlement.go`：每日结算表 `DailySettlement`
  - 日期、国家、币种、销售额、广告费、汇率、货款成本、刷单费用、固定成本、广告折算、平台手续费、利润、备注。
- `user.go`：用户表 `User`
  - 用户名、密码哈希、角色（superadmin/admin/staff）、页面权限字符串（例如 `"settlement,product"`）。

你可以把这些理解成数据库中的几张核心表。

### 4. 认证与用户管理 handlers

#### 4.1 登录与身份校验（auth.go）

- `Login(db)`：处理 `/api/login`：
  1. 接收 JSON：`{ username, password }`。
  2. 在 `users` 表中查找用户；用 bcrypt 比对密码。
  3. 生成 JWT，写入：用户 ID、用户名、角色。
  4. 返回给前端：`{ token, user: { id, username, role, permissions } }`。
- `AuthMiddleware()`：校验每个需要登录的请求：
  1. 从 `Authorization: Bearer <token>` 取出 token。
  2. 解析 JWT，把 `userID/username/role` 放进 Gin 的 `Context`。
- `RequireRole("superadmin")`：
  - 检查当前 `role` 是否在允许的列表中，否则返回 403。
- `Me()`：返回当前登录用户信息 `/api/me`。

#### 4.2 用户与角色管理（user_admin.go）

这些接口都挂在 `/api/users/...` 下，并且 **只有超级管理员** 可以访问（在 main.go 中通过 `RequireRole("superadmin")` 控制）：

- `ListUsers(db)`：`GET /api/users`
  - 查询所有用户，返回去掉密码哈希后的列表，包含：用户名、角色、页面权限、创建时间。
- `CreateUser(db)`：`POST /api/users`
  - body：`{ username, password, role, permissions[] }`。
  - 检查角色是否合法，用户名是否已存在。
  - bcrypt 加密密码，保存到 `users` 表。
- `UpdateUserPermissions(db)`：`PUT /api/users/:id/permissions`
  - body：`{ permissions: [] }`。
  - 仅允许 `settlement` / `product` 两种页面权限。
  - 防止用户修改自己的权限把自己锁死。
- `UpdateUserRole(db)`：`PUT /api/users/:id/role`
  - body：`{ role }`。
  - 校验角色，防止超级管理员把自己的角色改掉。
- `UpdateUserPassword(db)`：`PUT /api/users/:id/password`
  - body：`{ password }`。
  - 重新生成密码哈希并保存，实现“root 随时重置任何用户密码”。

---

## 三、后端：业务功能接口

### 1. 商品管理相关

文件：`internal/handlers/handlers.go` 中的商品部分。

- `ListProducts(db)`：`GET /api/products`
  - 查询所有 `Product`。
  - 根据当前登录用户角色，决定返回哪种成本字段（员工看到的是 employee 成本，管理员看到 admin 成本，超级管理员看到所有）。
- `SaveProduct(db)`：`POST /api/products`
  - 用于新增或修改商品：
    - 如果 body 中带 `id`，先查出原记录再更新。
    - 超级管理员可以一次性配置三种成本；管理员只能改自己的成本；员工不能直接改成本。
- `DeleteProduct(db)`：`DELETE /api/products/:id`
- `UploadProductImage()`：`POST /api/products/upload`
  - 接收上传图片，保存到 `uploads/products` 目录。
  - 生成 URL `/uploads/products/xxxx.jpg` 返回给前端。

### 2. 订单录入与统计

仍在 `handlers.go`：

- `PostOrder(db)`：`POST /api/order`
  - 前端录入订单或总额时调用。
  - 接收 `Order` 的 JSON（国家、商品、SKU、数量、总额等），插入 `orders` 表。
  - 如果没传 `created_at`，会用当前时间。
- `ListOrders(db)`：`GET /api/orders`（代码在后半部分）
  - 支持按日期（和国家）查询当天所有订单，用于“订单列表”和“今日已提交出单记录”。
- `UpdateOrder(db)` / `UpdateOrderDate(db)` / `DeleteOrder(db)`：
  - 对单条订单进行修改、改日期、删除。

#### 2.1 今日销售额 TodaySales

- 接口：`GET /api/sales/today`
- 逻辑：
  - 查找今天日期内，`product_name = '今日总额汇总'` 的最新一条 `Order` 记录。
  - 返回其中的 `total_amount` 作为“今日销售额（人民币）”。

#### 2.2 今日货款成本 TodayGoodsCost

- 接口：`GET /api/costs/today`
- 逻辑（在 `TodayGoodsCost` 函数中）：
  - 对当天所有订单（排除“今日总额汇总”等汇总记录），按商品 SKU 找到对应的商品成本（`Product.Cost`）。
  - 计算：`sum(数量 × 成本)`，得到当天货款成本（人民币）。

#### 2.3 图表统计：Hourly/Daily/Monthly/SalesTrend/TopProducts

- `HourlyStats(db)`：`GET /api/stats/hourly`（前端目前不再使用，只保留接口）
  - 按小时统计某天销售金额，排除“今日总额汇总/今日总汇”记录。
- `DailyStats(db)`：`GET /api/stats/daily?days=7`
  - 近 N 天，每天的销售金额。
  - 这里是从 `orders` 中筛选 `product_name` 为 “今日总额汇总/今日总汇”，按天求和。
- `MonthlyStats(db)`：`GET /api/stats/monthly?year=YYYY`
  - 按月统计某年每月销售金额（使用订单汇总记录）。
- `SalesTrend(db)`：早期版本的趋势接口（现在主要用 `DailyStats/MonthlyStats`）。
- `TopProducts(db)`：`GET /api/stats/top-products`
  - 按商品名汇总当天 **数量**，返回销量排行前几名，用于“商品销售排行”柱状图。

### 3. 结算工具与每日结算记录

#### 3.1 保存每日结算 SaveSettlement

- 接口：`POST /api/settlement`
- 前端（结账工具）在用户输入好当日数据后调用，body 包含：
  - 日期、国家、币种、当天销售总额、广告费、汇率、货款成本、刷单费用、固定成本、备注。
- 后端会：
  1. 如果 `date` 为空，用今天日期。
  2. 计算：
     - 广告费折算 `ad_deduction = (广告费 + 广告费 × 11%) × 汇率`
     - 平台手续费 `platform_fee = 当天销售总额 × 7%`
     - 利润 `profit = 销售额 - 广告折算 - 货款成本 - 手续费 - 刷单 - 固定成本`
  3. 写入 `daily_settlements` 表。

#### 3.2 查询每日结算 ListSettlements

- 接口：`GET /api/settlements?date=YYYY-MM-DD&country=...`
- 用于前端“每日结算记录”列表展示和合计当日利润。

#### 3.3 广告费折算趋势

- `AdDeductionDailyStats(db)`：`GET /api/stats/ad-deduction/daily?days=7`
  - 从 `daily_settlements` 表按日期聚合 `ad_deduction`，用于“近 7 天广告费折算”柱状图。
- `AdDeductionMonthlyStats(db)`：`GET /api/stats/ad-deduction/monthly?year=YYYY`
  - 按年+月聚合广告折算金额，返回 12 个月的数组，用于“按月广告费折算”图表。

### 4. 汇率接口

#### 4.1 工具层：`internal/utils/exchange.go`

- `GetRates()`：
  - 调第三方接口 exchangerate.host 的 `/live`。
  - 以 CNY 为基准，获取 PHP/IDR/MYR 相对人民币的汇率。
  - 做 10 分钟缓存，避免频繁请求外部服务。
- `CurrencyName(cur)`：返回币种对应的中文名。

#### 4.2 对外 handler：`internal/handlers/exchange.go`

- `ExchangeRates(c)`：`GET /api/exchange/rates`
  - 调用 `utils.GetRates()` 取汇率。
  - 返回 JSON：
    - `rates`: `{ PHP: ..., IDR: ..., MYR: ..., CNY: 1 }`
    - `labels`: 币种对应中文名。

---

## 四、前端：从页面到接口

### 1. 启动与全局布局

- `frontend/src/main.js`：
  - 创建 Vue 应用，挂载 Element Plus。
  - 如果本地有 `ordercount-token`，自动把它加到 axios 默认请求头里（保证刷新后仍保持登录态）。
- `frontend/src/App.vue`：
  - 顶层布局：
    - 未登录：显示 `<Login />` 组件。
    - 已登录：显示左侧菜单 + 顶部用户信息 + 右侧各个业务页面。
  - 根据 `currentUser.role` 和 `currentUser.permissions` 控制各菜单是否可见：
    - 订单统计（所有人可见）
    - 结账工具（需要 settlement 权限或超级管理员）
    - 商品管理（需要 product 权限或超级管理员）
    - 用户管理（仅超级管理员）

### 2. 登录与权限

- 组件：`frontend/src/components/Login.vue`
  - 表单：用户名 + 密码（带小眼睛可见/隐藏）。
  - 点击“登录”按钮：
    - 调用 `POST /api/login`。
    - 成功后：
      - 把 `token` 存到 `localStorage`，并设置到 axios 默认头。
      - 把 `user` 信息存到 `localStorage`，并通过事件把 `user` 传给 App.vue。
- 组件：`frontend/src/components/UserManager.vue`
  - 入口：在 App.vue 中，当 `activeMenu === 'users' && isSuperAdmin` 时显示。
  - 用到的接口：
    - `GET /api/users`：加载所有用户。
    - `POST /api/users`：创建新用户。
    - `PUT /api/users/:id/role`：修改角色。
    - `PUT /api/users/:id/permissions`：修改页面权限。
    - `PUT /api/users/:id/password`：重置任意用户密码（在“修改密码”对话框中调用）。

### 3. 订单录入与统计页面（菜单：订单统计）

#### 3.1 录入订单：OrderForm.vue

- 文件：`frontend/src/components/OrderForm.vue`
- 功能：
  1. 录入当天的各个 SKU 出单明细：国家 + SKU + 数量。
  2. 最后录入“今日总额”，以一条特殊订单（商品名是 `"今日总额汇总"`）的形式保存。
- 接口调用：
  - `GET /api/products`：
    - 用于 SKU 下拉框的数据来源（SKU + 商品名），录入时只需要选 SKU。
  - `POST /api/order`：
    - 提交每条出单明细（总额为 0）。
    - 保存“今日总额汇总”记录（数量为 0，总额为填写的金额）。
  - `GET /api/orders`：
    - 加载当天所有订单，用于下方“今日已提交出单记录”表格。
  - `PUT /api/orders/:id`、`DELETE /api/orders/:id`：
    - 修改或删除某条出单明细（前端在表格操作中调用）。

#### 3.2 今日销售额：TodaySales.vue

- 接口：`GET /api/sales/today`
- 展示：一个卡片，显示“人民币金额：￥xxx.xx”。

#### 3.3 今日货款成本：TodayGoodsCost.vue

- 接口：`GET /api/costs/today`
- 展示：一个卡片，显示“今日货款成本（人民币）：￥xxx.xx”。

#### 3.4 图表统计：OrderCharts.vue

- 接口：
  - `GET /api/stats/daily?days=7`：近 7 天销售金额折线图。
  - `GET /api/stats/monthly?year=YYYY`：按月销售金额折线图。
  - `GET /api/stats/top-products`：商品销量排行柱状图（按数量）。
- 图表库：ECharts
  - 近 7 天和按月两个模式，支持切换，并在图线上直接显示数值标签（便于肉眼查看）。

#### 3.5 订单列表：OrderList.vue

- 接口：`GET /api/orders?date=YYYY-MM-DD&country=...`
- 功能：
  - 展示某天的所有订单记录（国家筛选）。
  - 支持修改单条订单的日期（调用 `PUT /api/orders/:id/date`）。

### 4. 结账工具与每日结算（菜单：结账工具）

#### 4.1 结账工具：ProfitTool.vue

- 接口使用：
  - `GET /api/exchange/rates`：
    - 获取当前主要货币汇率，用于自动计算“1 外币 ≈ ? 人民币”的汇率，和顶部汇率栏展示。
  - `GET /api/sales/today`：
    - 自动带出“当天销售总额”，输入框只读。
  - `GET /api/costs/today`：
    - 自动带出“货款成本”，输入框只读。
  - `POST /api/settlement`：
    - 将用户填好的广告费、刷单费、固定成本等，连同销售额/货款/汇率一起保存为一条每日结算记录。
  - `GET /api/stats/ad-deduction/daily`、`GET /api/stats/ad-deduction/monthly`：
    - 用于“广告费折算趋势”柱状图（近 7 天 / 按月）。
- 前端计算展示：
  - 根据当前表单中的值，实时在说明区域显示：广告费折算、平台手续费、利润等公式和结果。

#### 4.2 每日结算记录：SettlementList.vue

- 接口：`GET /api/settlements?date=YYYY-MM-DD&country=...`
- 展示：
  - 同一天可以有多条结算记录（比如一天多次结算）。
  - 列出每条的：销售额、广告费、汇率、广告折算金额、平台手续费、货款成本、刷单、固定成本、备注、利润。
  - 底部合计当日所有记录的利润总和。

### 5. 商品管理（菜单：商品管理）

- 组件：`ProductManager.vue`
- 接口：
  - `GET /api/products`：加载商品列表。
  - `POST /api/products`：新增或更新商品（区别在于是否传 id）。
  - `DELETE /api/products/:id`：删除商品。
  - `POST /api/products/upload`：上传商品图片，返回图片 URL。
- 权限控制（前端 + 后端配合）：
  - 员工：只能看到成本，不能修改删除。
  - 管理员：可以维护自己角色的成本，编辑/删除商品。
  - 超级管理员：可配置三种角色成本，查看全部成本字段。

### 6. 汇率栏：ExchangeRatesBar.vue

- 接口：`GET /api/exchange/rates`
- 展示：
  - “当前主要汇率（1 人民币 ≈ ? 外币）”，显示 PHP/IDR/MYR 三种货币的当前（缓存）汇率。

---

## 五、开发常用操作与建议

### 1. 本地启动

```bash
# 后端（在项目根目录）
cd ordercount
MYSQL_DSN="user:pass@tcp(127.0.0.1:3306)/ordercount?charset=utf8mb4&parseTime=True&loc=Local" go run .

# 前端（在 frontend 目录）
cd frontend
npm install
npm run dev
```

- 浏览器访问：`http://localhost:5173/`。

### 2. 角色与权限测试

- 首次启动后数据库会自动有两个账号：
  - 超级管理员：`root / root123`
  - 管理员：`admin / admin123`
- 建议：
  - 用 root 登录后，先在“用户管理”中修改默认密码，并按需要创建普通员工账号。

### 3. 如何新增一个接口（思路）

1. 在 `internal/models` 中为新表建模型（如有需要）。
2. 在 `internal/handlers` 里写 Handler 函数：
   - 接收参数 `func Xxx(db *gorm.DB) gin.HandlerFunc`。
   - 在内部使用 `db` 操作模型，返回 JSON。
3. 在 `main.go` 中注册路由：
   - 如：`api.GET("/your/path", handlers.YourHandler(gdb))`。
4. 在前端：
   - 在对应组件中用 axios 调这个接口，处理返回数据并展示。

---

## 六、总结

- 可以把整个项目理解为：
  1. **后端** 提供一组清晰的 REST 接口：订单、每日结算、商品、用户、汇率、统计。
  2. **前端** 通过 axios 调用这些接口，组合成“订单统计”、“结账工具”、“商品管理”、“用户管理”四大页面。
  3. 各种图表、卡片、表格，其实都是对这些接口返回数据的不同展示方式。

当你要阅读或修改功能时，可以按照下面的路径来查代码：

- 先在前端组件里找到调用的 URL（比如 `/api/xxx`）。
- 在 `main.go` 里搜索这个路径，看看对应哪个 handler。
- 打开对应的 handler 函数，看它是如何从数据库读写数据的。

这样就能从“前端页面按钮”一路 tracing 到“数据库表字段”，对整个系统建立完整的理解。希望这份文档能让你作为新手也能比较轻松地上手这个项目。
