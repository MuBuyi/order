package main

import (
    "io/ioutil"
    "log"
    "os"

    "github.com/gin-gonic/gin"

    "ordercount/internal/db"
    "ordercount/internal/handlers"
    "gopkg.in/yaml.v3"
)

type Config struct {
    MySQL struct {
        DSN string `yaml:"dsn"`
    } `yaml:"mysql"`
}

func main() {
    var dsn string
    // 优先从 config.yaml 读取
    if data, err := ioutil.ReadFile("config.yaml"); err == nil {
        var cfg Config
        if err := yaml.Unmarshal(data, &cfg); err == nil && cfg.MySQL.DSN != "" {
            dsn = cfg.MySQL.DSN
        }
    }
    if dsn == "" {
        dsn = os.Getenv("MYSQL_DSN")
    }
    if dsn == "" {
        log.Fatal("please set MYSQL_DSN environment variable or config.yaml")
    }

    gdb, err := db.InitDB(dsn)
    if err != nil {
        log.Fatalf("db init: %v", err)
    }

    r := gin.Default()

    api := r.Group("/api")
    {
        // 登录与用户信息
        api.POST("/login", handlers.Login(gdb))
        authGroup := api.Group("")
        authGroup.Use(handlers.AuthMiddleware())
        authGroup.GET("/me", handlers.Me())

        // 用户与角色管理（仅超级管理员）
        users := api.Group("/users")
        users.Use(handlers.AuthMiddleware(), handlers.RequireRole("superadmin"))
        users.GET("", handlers.ListUsers(gdb))
        users.POST("", handlers.CreateUser(gdb))
        // 注意这里的路径需要以斜杠开头，才能匹配 /api/users/:id/role 这样的请求
        users.PUT("/:id/role", handlers.UpdateUserRole(gdb))
        users.PUT("/:id/permissions", handlers.UpdateUserPermissions(gdb))

        api.POST("/order", handlers.PostOrder(gdb))
        api.GET("/sales/today", handlers.TodaySales(gdb))
        api.GET("/costs/today", handlers.TodayGoodsCost(gdb))
        api.GET("/report/export", handlers.ExportReport(gdb))
        api.GET("/stats/sales-trend", handlers.SalesTrend(gdb))
        api.GET("/stats/top-products", handlers.TopProducts(gdb))
        api.GET("/stats/ad-deduction/daily", handlers.AdDeductionDailyStats(gdb))
        api.GET("/stats/ad-deduction/monthly", handlers.AdDeductionMonthlyStats(gdb))
		api.POST("/settlement", handlers.SaveSettlement(gdb))
		api.GET("/settlements", handlers.ListSettlements(gdb))
        api.GET("/orders", handlers.ListOrders(gdb))
        api.PUT("/orders/:id", handlers.UpdateOrder(gdb))
        api.PUT("/orders/:id/date", handlers.UpdateOrderDate(gdb))
        api.DELETE("/orders/:id", handlers.DeleteOrder(gdb))

        // 商品管理（需登录）
        products := api.Group("/products")
        products.Use(handlers.AuthMiddleware())
        products.GET("", handlers.ListProducts(gdb))
        products.POST("", handlers.SaveProduct(gdb))
        products.DELETE(":id", handlers.DeleteProduct(gdb))
        products.POST("/upload", handlers.UploadProductImage())

        // 新增统计接口
        api.GET("/stats/hourly", handlers.HourlyStats(gdb))
        api.GET("/stats/daily", handlers.DailyStats(gdb))
        api.GET("/stats/monthly", handlers.MonthlyStats(gdb))

		// 汇率接口
		api.GET("/exchange/rates", handlers.ExchangeRates)
    }

    // 静态资源与上传文件
    r.StaticFile("/", "frontend/dist/index.html")
    r.Static("/assets", "frontend/dist/assets")
    r.StaticFile("/vite.svg", "frontend/dist/vite.svg")
    // 本地上传的文件（商品图片等）
    r.Static("/uploads", "uploads")

    r.Run(":8080")
}
