# 订单统计系统 API 接口文档（前后端联调用）

> 说明：本项目后端基于 Go + Gin，所有接口路径均以 `/api` 为前缀。除登录接口外，其余大部分接口都需要在请求头中携带 `Authorization: Bearer <token>`。

---

## 1. 认证与用户

### 1.1 登录

- 方法：`POST`
- 路径：`/api/login`
- 是否需要登录：否
- 请求体（JSON）：
  - `username` (string) 必填：用户名
  - `password` (string) 必填：密码
- 返回示例（JSON）：
  - `token`：JWT 字符串，后续请求需放入 Header `Authorization: Bearer <token>`
  - `user`：用户信息对象（包含 `id`、`username`、`role`、`permissions` 等）

### 1.2 当前用户信息

- 方法：`GET`
- 路径：`/api/me`
- 是否需要登录：是
- Query / Body：无
- 返回：当前登录用户信息对象

### 1.3 用户列表（仅超级管理员）

- 方法：`GET`
- 路径：`/api/users`
- 是否需要登录：是（角色必须为 `superadmin`）
- Query：无
- 返回：用户数组

### 1.4 创建用户（仅超级管理员）

- 方法：`POST`
- 路径：`/api/users`
- 是否需要登录：是（`superadmin`）
- 请求体（JSON）：
  - `username` (string) 必填
  - `password` (string) 必填
  - `role` (string) 可选：`superadmin` / `admin` / `staff`
  - `permissions` (string) 可选：逗号分隔页面权限标识，如 `"settlement,product,shop"`
- 返回：创建的用户信息或 `{ "success": true }`

### 1.5 修改用户角色（仅超级管理员）

- 方法：`PUT`
- 路径：`/api/users/:id/role`
- 是否需要登录：是（`superadmin`）
- 路径参数：
  - `id`：用户 ID
- 请求体（JSON）：
  - `role` (string) 必填
- 返回：更新后的用户信息或 `{ "success": true }`

### 1.6 修改用户权限（仅超级管理员）

- 方法：`PUT`
- 路径：`/api/users/:id/permissions`
- 是否需要登录：是（`superadmin`）
- 路径参数：
  - `id`：用户 ID
- 请求体（JSON）：
  - `permissions` (array<string> 或 string) 必填：前端通常传数组，例如 `["settlement","product"]`
- 返回：更新后的用户信息或 `{ "success": true }`

### 1.7 重置用户密码（仅超级管理员）

- 方法：`PUT`
- 路径：`/api/users/:id/password`
- 是否需要登录：是（`superadmin`）
- 路径参数：
  - `id`：用户 ID
- 请求体（JSON）：
  - `password` (string) 必填：新密码
- 返回：`{ "success": true }` 或错误信息

---

## 2. 订单录入与查询

### 2.1 录入订单 / 今日总额汇总

- 方法：`POST`
- 路径：`/api/order`
- 是否需要登录：是
- 请求体（JSON）：分两种模式

#### 2.1.1 普通订单明细模式

- 字段：
  - `date` (string) 必填：业务日期，格式 `YYYY-MM-DD`
  - `country` (string) 必填：国家，值限定为 `"菲律宾"` / `"印尼"` / `"马来西亚"`
  - `sku` (string) 必填：商品 SKU
  - `quantity` (number) 必填：数量
  - `total_amount` (number) 必填：此条明细对应金额
  - `product_name` (string) 可选：商品名称（通常后端会基于 SKU 补充）

#### 2.1.2 今日总额汇总模式

- 字段：
  - `date` (string) 必填：业务日期 `YYYY-MM-DD`
  - `country` (string) 必填
  - `product_name` (string) 必填：固定为 `"今日总额汇总"`
  - `total_amount` (number) 必填：当天总销售额（人民币）

- 返回：保存后的订单记录信息或 `{ "error": "..." }`

### 2.2 订单列表

- 方法：`GET`
- 路径：`/api/orders`
- 是否需要登录：是
- Query 参数：
  - `date` (string) 可选：业务日期 `YYYY-MM-DD`
  - 可能还包含 `page`、`page_size` 等（视后端实现而定）
- 返回：订单数组，每项包含：
  - `id`、`country`、`platform`、`order_no`、`product_name`、`sku`、`quantity`、`total_amount`、`currency`、`created_at`（字符串）等

### 2.3 修改订单

- 方法：`PUT`
- 路径：`/api/orders/:id`
- 是否需要登录：是
- 路径参数：
  - `id`：订单 ID
- 请求体（JSON）：
  - 可更新字段（与前端编辑表单一致）：
    - `country` (string)
    - `sku` (string)
    - `quantity` (number)
    - `total_amount` (number)
    - `platform` (string) 等
- 返回：更新后的订单信息或 `{ "success": true }`

### 2.4 修改订单业务日期

- 方法：`PUT`
- 路径：`/api/orders/:id/date`
- 是否需要登录：是
- 路径参数：
  - `id`：订单 ID
- 请求体（JSON）：
  - `date` (string) 必填：新业务日期 `YYYY-MM-DD`
- 返回：`{ "success": true }` 或更新后的订单信息

### 2.5 删除订单

- 方法：`DELETE`
- 路径：`/api/orders/:id`
- 是否需要登录：是
- 路径参数：
  - `id`：订单 ID
- 返回：`{ "success": true }` 或 `{ "error": "..." }`

---

## 3. 今日销售额与成本

### 3.1 今日销售额汇总

- 方法：`GET`
- 路径：`/api/sales/today`
- 是否需要登录：否
- Query 参数：
  - `date` (string) 可选：业务日期 `YYYY-MM-DD`，不传则使用“今天”（雅加达时区）的业务日期
- 返回（JSON）：
  - `total_amount` (number)：当天销售总额

### 3.2 今日货款成本

- 方法：`GET`
- 路径：`/api/costs/today`
- 是否需要登录：否
- Query 参数：
  - `date` (string) 可选：业务日期
- 返回（JSON）：
  - 一般为 `{ "total": number }` 或类似字段，表示当天货款成本之和

---

## 4. 每日结算（利润工具）

### 4.1 保存每日结算

- 方法：`POST`
- 路径：`/api/settlement`
- 是否需要登录：是
- 请求体（JSON）：
  - `date` (string) 必填：业务日期 `YYYY-MM-DD`，为空时后端使用当天
  - `country` (string) 必填：国家
  - `currency` (string) 必填：币种，如 `PHP`、`IDR`、`USD`
  - `sale_total` (number) 必填：当天销售额
  - `ad_cost` (number) 必填：广告费（原币）
  - `exchange` (number) 必填：汇率（1 外币 ≈ ? CNY）
  - `goods_cost` (number) 必填：货款成本（人民币）
  - `shua_dan_fee` (number) 可选：刷单费用（人民币）
  - `fixed_cost` (number) 可选：固定成本（人民币）
  - `ad_deduction` (number) 可选：广告折算成本，通常由后端计算
  - `platform_fee` (number) 可选：平台手续费，通常由后端计算
  - `profit` (number) 可选：利润，通常以前端展示为主，最终以后端计算为准
  - `remark` (string) 可选：备注
- 行为：后端根据 `date + country + user_id` 查找已有记录，有则更新，无则创建
- 返回：保存后的结算记录或 `{ "success": true }`

### 4.2 查询每日结算列表

- 方法：`GET`
- 路径：`/api/settlements`
- 是否需要登录：是
- Query 参数：
  - `page` (number) 可选：页码，默认 1
  - `page_size` (number) 可选：每页条数，默认 10
  - `start_date` (string) 可选：起始业务日期
  - `end_date` (string) 可选：结束业务日期
  - `date` (string) 可选：单一业务日期（与 start/end 互斥）
  - `country` (string) 可选：国家
- 返回（JSON）：
  - `date`：本次查询的日期（可能为空）
  - `items`：结算记录数组；每项包含
    - `id`、`date`、`country`、`currency`、`sale_total`、`ad_cost`、`exchange`、`goods_cost`、`shua_dan_fee`、`fixed_cost`、`ad_deduction`、`platform_fee`、`profit`、`remark`、`created_at`
  - `total`：总记录数
  - `page`：当前页码
  - `page_size`：每页条数

### 4.3 主动推送结算日报

- 方法：`POST`
- 路径：`/api/settlements/push`
- 是否需要登录：是
- 请求体（JSON）：
  - `date` (string) 可选：业务日期 `YYYY-MM-DD`
    - SettlementList 中，当所选区间为同一天时会传此字段
- 行为：后端根据 date（缺省则取最近日期）汇总结算数据，通过企业微信 webhook 推送
- 返回：`{ "success": true }` 或错误信息

---

## 5. 统计接口

### 5.1 仪表盘汇总

- 方法：`GET`
- 路径：`/api/stats/dashboard/summary`
- 是否需要登录：否
- Query 参数：
  - 可能包含 `start_date`、`end_date`、`country` 等（视后端实现）
- 返回：用于首页 / 统计页的汇总数据对象

### 5.2 畅销商品

- 方法：`GET`
- 路径：`/api/stats/top-products`
- 是否需要登录：否
- Query 参数（部分可选）：
  - `limit` (number) 可选：返回前几名
  - `start_date`、`end_date` (string) 可选：日期区间
- 返回：商品统计数组（每项含 SKU/名称及销量/销售额）

### 5.3 近 N 天每日统计

- 方法：`GET`
- 路径：`/api/stats/daily`
- 是否需要登录：否
- Query 参数：
  - `days` (number) 可选：默认 7，表示最近 N 天
- 返回：日期 + 对应统计值数组，用于折线图

### 5.4 按月统计

- 方法：`GET`
- 路径：`/api/stats/monthly`
- 是否需要登录：否
- Query 参数：
  - `year` (number) 必填：年份，例如 `2026`
- 返回：12 个月或有数据月份的统计数组

### 5.5 广告费用折算（近 N 天）

- 方法：`GET`
- 路径：`/api/stats/ad-deduction/daily`
- 是否需要登录：否
- Query 参数：
  - `days` (number) 可选：1–60，默认 7
- 返回：`[{ day: "YYYY-MM-DD", total: number }, ...]`

### 5.6 广告费用折算（按月）

- 方法：`GET`
- 路径：`/api/stats/ad-deduction/monthly`
- 是否需要登录：否
- Query 参数：
  - `year` (number) 必填
- 返回：每个月的广告折算总额数组

---

## 6. 店铺与店铺日数据

### 6.1 店铺列表

- 方法：`GET`
- 路径：`/api/shops`
- 是否需要登录：是
- Query 参数：
  - `country` (string) 可选：按国家筛选；支持中文值（如 `菲律宾`），`全部` 或不传表示不过滤
  - `status` (string) 可选：店铺状态筛选；可选值：`all` / `enabled` / `disabled`
  - `is_blocked` (bool) 可选：按封禁状态筛选；`true` 表示停用，`false` 表示启用
  - 说明：当 `status` 与 `is_blocked` 同时传入时，两者条件会同时生效
- 返回：店铺数组

### 6.2 新增 / 更新店铺

- 方法：`POST`
- 路径：`/api/shops`
- 是否需要登录：是
- 请求体（JSON）：
  - `id` (number) 可选：存在则更新，不存在则创建
  - 其他字段：`name`、`country`、`platform`、`is_blocked` 等（以后端模型为准）
- 返回：保存后的店铺信息

### 6.3 删除店铺

- 方法：`DELETE`
- 路径：`/api/shops/:id`
- 是否需要登录：是
- 路径参数：
  - `id`：店铺 ID
- 返回：`{ "success": true }`

### 6.4 获取店铺绑定用户

- 方法：`GET`
- 路径：`/api/shops/:id/users`
- 是否需要登录：是
- 路径参数：
  - `id`：店铺 ID
- 返回：有权访问该店铺的用户数组

### 6.5 更新店铺绑定用户

- 方法：`POST`
- 路径：`/api/shops/:id/users`
- 是否需要登录：是
- 路径参数：
  - `id`：店铺 ID
- 请求体（JSON）：
  - `user_ids` (array<number>) 必填：用户 ID 列表
- 返回：`{ "success": true }`

### 6.6 店铺日数据列表

- 方法：`GET`
- 路径：`/api/store_stats`
- 是否需要登录：是
- Query 参数：
  - `date` (string) 可选：业务日期 `YYYY-MM-DD`
  - 可能还有店铺 / 国家等过滤条件
- 返回：店铺每日统计数组（含广告费等）

### 6.7 保存店铺日数据

- 方法：`POST`
- 路径：`/api/store_stats`
- 是否需要登录：是
- 请求体（JSON）：
  - `date` (string) 必填：业务日期
  - `store_id` / `shop_id` (number) 必填：店铺 ID
  - `ad_cost` 等字段：广告费用等
- 返回：`{ "success": true }` 或保存后的记录

### 6.8 店铺区间数据

- 方法：`GET`
- 路径：`/api/store_stats/range`
- 是否需要登录：是
- Query 参数：
  - `start_date` (string) 必填：起始业务日期
  - `end_date` (string) 必填：结束业务日期
  - 其他筛选条件视后端实现
- 返回：区间内店铺日数据列表

---

## 7. 商品管理

### 7.1 商品列表

- 方法：`GET`
- 路径：`/api/products`
- 是否需要登录：是
- Query：无
- 返回：商品数组；字段包括
  - `id`、`sku`、`name`、`image_url`、`cost`、`cost_admin`（仅 superadmin）、`cost_staff`（仅 superadmin）等

### 7.2 新增 / 更新商品

- 方法：`POST`
- 路径：`/api/products`
- 是否需要登录：是
- 请求体（JSON）：
  - `id` (number) 可选：存在则更新
  - `sku` (string) 必填
  - `name` (string) 必填
  - `image_url` (string) 可选：图片地址
  - `cost` (number) 必填：当前角色看到/维护的成本
  - `cost_admin` (number) 可选：超级管理员专用成本
  - `cost_staff` (number) 可选：员工专用成本
- 返回：保存后的商品信息

### 7.3 删除商品

- 方法：`DELETE`
- 路径：`/api/products/:id`
- 是否需要登录：是
- 路径参数：
  - `id`：商品 ID
- 返回：`{ "success": true }`

### 7.4 上传商品图片

- 方法：`POST`
- 路径：`/api/products/upload`
- 是否需要登录：是
- 请求体：`multipart/form-data`
  - `file`：文件字段，上传的图片
- 返回（JSON）：
  - `url` (string)：图片访问地址

---

## 8. 其它公共接口

### 8.1 国家列表

- 方法：`GET`
- 路径：`/api/countries`
- 是否需要登录：否
- Query：无
- 返回：国家字符串数组，例如 `["菲律宾","印尼","马来西亚"]`

### 8.2 汇率

- 方法：`GET`
- 路径：`/api/exchange/rates`
- 是否需要登录：否
- Query：无
- 返回：主要币种对人民币的汇率信息和最近更新时间

### 8.3 首页概览

- 方法：`GET`
- 路径：`/api/dashboard/home`
- 是否需要登录：是
- Query：
  - `year`：可选，年份，如 `2026`
  - `month`：可选，月份，如 `06`
  - 未传时默认返回当前月数据
- 返回：用于首页 HomeDashboard 的统计数据

### 8.4 手动推送订单日报

- 方法：`POST`
- 路径：`/api/notify/wecom/today-orders`
- 是否需要登录：是
- 请求体：空 JSON 或省略
- 返回：`{ "success": true }`

### 8.5 导出报表 CSV

- 方法：`GET`
- 路径：`/api/report/export`
- 是否需要登录：视后端实现（推荐登录后使用）
- Query：
  - 一般包含 `start_date`、`end_date`、`country` 等筛选条件
- 返回：`text/csv` 文件下载

---

> 本文档为前后端联调用接口说明，如有新增接口或参数变更，建议同步更新本文件。