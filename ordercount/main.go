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
        api.POST("/order", handlers.PostOrder(gdb))
        api.POST("/orders/import", handlers.ImportOrders(gdb))
        api.GET("/sales/today", handlers.TodaySales(gdb))
        api.GET("/report/export", handlers.ExportReport(gdb))
        api.GET("/stats/sales-trend", handlers.SalesTrend(gdb))
        api.GET("/stats/top-products", handlers.TopProducts(gdb))
    }


    r.StaticFile("/", "frontend/dist/index.html")
    r.Static("/assets", "frontend/dist/assets")
    r.StaticFile("/vite.svg", "frontend/dist/vite.svg")

    r.Run(":8080")
}
