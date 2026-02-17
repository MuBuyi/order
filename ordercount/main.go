package main

import (
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "time"

    "github.com/gin-gonic/gin"
    "gopkg.in/yaml.v3"

    "ordercount/internal/db"
    "ordercount/internal/handlers"
)

type Config struct {
    MySQL struct {
        DSN string `yaml:"dsn"`
    } `yaml:"mysql"`
    Wecom struct {
        Webhook             string `yaml:"webhook"`
        PushTime            string `yaml:"push_time"`
        SettlementPushTime  string `yaml:"settlement_push_time"`
    } `yaml:"wecom"`
    AI struct {
        DoubaoAPIKey string `yaml:"doubao_api_key"`
        DoubaoEndpoint string `yaml:"doubao_endpoint"`
        DoubaoModel string `yaml:"doubao_model"`
    } `yaml:"ai"`
}

func main() {
    // 将全局时区设置为雅加达时间（UTC+7），用于所有时间存储和“今天/昨天”等逻辑
    if loc, err := time.LoadLocation("Asia/Jakarta"); err != nil {
        log.Printf("failed to load Asia/Jakarta location, using default local timezone: %v", err)
    } else {
        time.Local = loc
    }

    var dsn string
    var wecomWebhook string
    var wecomPushTime string
    var wecomSettlementPushTime string
    var doubaoAPIKey string
    var doubaoEndpoint string
    var doubaoModel string
    // 优先从 config.yaml 读取
    if data, err := ioutil.ReadFile("config.yaml"); err == nil {
        var cfg Config
        if err := yaml.Unmarshal(data, &cfg); err == nil {
            if cfg.MySQL.DSN != "" {
                dsn = cfg.MySQL.DSN
            }
            if cfg.Wecom.Webhook != "" {
                wecomWebhook = cfg.Wecom.Webhook
            }
            if cfg.Wecom.PushTime != "" {
                wecomPushTime = cfg.Wecom.PushTime
            }
            if cfg.Wecom.SettlementPushTime != "" {
                wecomSettlementPushTime = cfg.Wecom.SettlementPushTime
            }
            if cfg.AI.DoubaoAPIKey != "" {
                doubaoAPIKey = cfg.AI.DoubaoAPIKey
            }
            if cfg.AI.DoubaoEndpoint != "" {
                doubaoEndpoint = cfg.AI.DoubaoEndpoint
            }
            if cfg.AI.DoubaoModel != "" {
                doubaoModel = cfg.AI.DoubaoModel
            }
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

    // 将企业微信 webhook 配置注入 handlers，优先来自 config.yaml
    if wecomWebhook != "" {
        handlers.WecomWebhook = wecomWebhook
    }

    // 将豆包 AI 配置注入 handlers，可来自 config.yaml 或环境变量
    if doubaoAPIKey == "" {
        doubaoAPIKey = os.Getenv("DOUBAO_API_KEY")
    }
    if doubaoEndpoint == "" {
        doubaoEndpoint = os.Getenv("DOUBAO_ENDPOINT")
    }
    if doubaoModel == "" {
        doubaoModel = os.Getenv("DOUBAO_MODEL")
    }
    handlers.DoubaoAPIKey = doubaoAPIKey
    handlers.DoubaoEndpoint = doubaoEndpoint
    handlers.DoubaoModel = doubaoModel

    api := r.Group("/api")
    {
        // 登录与用户信息
        api.POST("/login", handlers.Login(gdb))
        authGroup := api.Group("")
        authGroup.Use(handlers.AuthMiddleware())
        authGroup.GET("/me", handlers.Me())
        authGroup.POST("/notify/wecom/today-orders", handlers.NotifyWecomTodayOrders(gdb))

        // 用户与角色管理（仅超级管理员）
        users := api.Group("/users")
        users.Use(handlers.AuthMiddleware(), handlers.RequireRole("superadmin"))
        users.GET("", handlers.ListUsers(gdb))
        users.POST("", handlers.CreateUser(gdb))
        // 注意这里的路径需要以斜杠开头，才能匹配 /api/users/:id/role 这样的请求
        users.PUT("/:id/role", handlers.UpdateUserRole(gdb))
        users.PUT("/:id/permissions", handlers.UpdateUserPermissions(gdb))
        users.PUT("/:id/password", handlers.UpdateUserPassword(gdb))

        // 订单与结算相关接口需要登录，用于按用户过滤
        authGroup.POST("/order", handlers.PostOrder(gdb))
        authGroup.POST("/settlement", handlers.SaveSettlement(gdb))
        authGroup.GET("/settlements", handlers.ListSettlements(gdb))
        authGroup.GET("/orders", handlers.ListOrders(gdb))
        authGroup.PUT("/orders/:id", handlers.UpdateOrder(gdb))
        authGroup.PUT("/orders/:id/date", handlers.UpdateOrderDate(gdb))
        authGroup.DELETE("/orders/:id", handlers.DeleteOrder(gdb))

        // 汇总类统计保持原样，可按需要后续再加登录控制
        api.GET("/sales/today", handlers.TodaySales(gdb))
        api.GET("/costs/today", handlers.TodayGoodsCost(gdb))
        api.GET("/report/export", handlers.ExportReport(gdb))
        api.GET("/stats/sales-trend", handlers.SalesTrend(gdb))
        api.GET("/stats/top-products", handlers.TopProducts(gdb))
        api.GET("/stats/ad-deduction/daily", handlers.AdDeductionDailyStats(gdb))
        api.GET("/stats/ad-deduction/monthly", handlers.AdDeductionMonthlyStats(gdb))

        // 商品管理（需登录）
        products := api.Group("/products")
        products.Use(handlers.AuthMiddleware())
        products.GET("", handlers.ListProducts(gdb))
        products.POST("", handlers.SaveProduct(gdb))
        products.DELETE(":id", handlers.DeleteProduct(gdb))
        products.POST("/upload", handlers.UploadProductImage())

        // 店铺管理（需登录）
        shops := api.Group("/shops")
        shops.Use(handlers.AuthMiddleware())
        shops.GET("", handlers.ListStores(gdb))
        shops.POST("", handlers.SaveStore(gdb))
        shops.DELETE(":id", handlers.DeleteStore(gdb))
        shops.GET(":id/users", handlers.GetStoreUsers(gdb))
        shops.POST(":id/users", handlers.UpdateStoreUsers(gdb))

        // 新增统计接口
        api.GET("/stats/hourly", handlers.HourlyStats(gdb))
        api.GET("/stats/daily", handlers.DailyStats(gdb))
        api.GET("/stats/monthly", handlers.MonthlyStats(gdb))

		// 汇率接口
		api.GET("/exchange/rates", handlers.ExchangeRates)

        // 结算每日汇总手动推送接口（需登录）
        authGroup.POST("/settlements/push", handlers.TriggerSettlementPush(gdb))

        // 豆包配置与连通性测试接口（仅用于调试）
        api.GET("/test-doubao", handlers.TestDoubaoConfig())
    }

    // 静态资源与上传文件
    r.StaticFile("/", "frontend/dist/index.html")
    r.Static("/assets", "frontend/dist/assets")
    r.StaticFile("/vite.svg", "frontend/dist/vite.svg")
    // 本地上传的文件（商品图片等）
    r.Static("/uploads", "uploads")

    // 启动一个后台协程，每天在配置的时间自动推送前一天的订单日报到企业微信
    // push_time 配置格式为 "HH:MM"，例如 "05:00" 表示每天早上 5 点（北京时间）。
    go func() {
        // 默认时间 05:00（北京时间）
        hour, minute := 5, 0
        if wecomPushTime != "" {
            if len(wecomPushTime) == 5 {
                // 简单解析 "HH:MM"，不做过多校验
                if h := wecomPushTime[0:2]; h >= "00" && h <= "23" {
                    if m := wecomPushTime[3:5]; m >= "00" && m <= "59" {
                        // 转成整数
                        var hh, mm int
                        fmt.Sscanf(h, "%d", &hh)
                        fmt.Sscanf(m, "%d", &mm)
                        hour, minute = hh, mm
                    }
                }
            }
        }

        // 北京时间时区
        bjLoc, err := time.LoadLocation("Asia/Shanghai")
        if err != nil {
            log.Printf("[scheduler] 加载 Asia/Shanghai 时区失败，将使用本地时区：%v", err)
            bjLoc = time.Local
        }

        for {
            // 使用北京时间计算下一次推送时间
            nowBJ := time.Now().In(bjLoc)
            nextBJ := time.Date(nowBJ.Year(), nowBJ.Month(), nowBJ.Day(), hour, minute, 0, 0, bjLoc)
            if !nextBJ.After(nowBJ) {
                // 如果今天的推送时间已过，则推到明天（依然按北京时间）
                nextBJ = nextBJ.Add(24 * time.Hour)
            }

            sleepDuration := nextBJ.Sub(nowBJ)
            log.Printf("[scheduler] 下次企业微信订单日报推送时间（北京时间）：%s", nextBJ.Format(time.RFC3339))
            time.Sleep(sleepDuration)

            // 计算要推送的日期：前一天（按当前全局本地时区，即雅加达时间）
            pushDate := time.Now().Add(-24 * time.Hour).Format("2006-01-02")
            log.Printf("[scheduler] 开始自动推送企业微信订单日报，日期：%s", pushDate)

            if err := handlers.NotifyWecomOrdersForDate(gdb, pushDate); err != nil {
                log.Printf("[scheduler] 自动推送企业微信订单日报失败：%v", err)
            } else {
                log.Printf("[scheduler] 自动推送企业微信订单日报成功，日期：%s", pushDate)
            }
        }
    }()

    // 启动一个后台协程，每天在配置的时间通过企微机器人推送一次“结算提醒”
    go func() {
        // 默认时间 00:00（北京时间），可通过 wecom.settlement_push_time 覆盖
        hour, minute := 0, 0
        if wecomSettlementPushTime != "" {
            if len(wecomSettlementPushTime) == 5 {
                if h := wecomSettlementPushTime[0:2]; h >= "00" && h <= "23" {
                    if m := wecomSettlementPushTime[3:5]; m >= "00" && m <= "59" {
                        var hh, mm int
                        fmt.Sscanf(h, "%d", &hh)
                        fmt.Sscanf(m, "%d", &mm)
                        hour, minute = hh, mm
                    }
                }
            }
        }

        // 北京时间时区
        bjLoc, err := time.LoadLocation("Asia/Shanghai")
        if err != nil {
            log.Printf("[settlement-scheduler] 加载 Asia/Shanghai 时区失败，将使用本地时区：%v", err)
            bjLoc = time.Local
        }

        for {
            // 使用北京时间计算下一次推送时间
            nowBJ := time.Now().In(bjLoc)
            nextBJ := time.Date(nowBJ.Year(), nowBJ.Month(), nowBJ.Day(), hour, minute, 0, 0, bjLoc)
            if !nextBJ.After(nowBJ) {
                // 如果今天的推送时间已过，则推到明天同一时间（北京时间）
                nextBJ = nextBJ.Add(24 * time.Hour)
            }

            sleepDuration := nextBJ.Sub(nowBJ)
            log.Printf("[settlement-scheduler] 下次每日结算提醒推送时间（北京时间）：%s", nextBJ.Format(time.RFC3339))
            time.Sleep(sleepDuration)

            // 计算要推送的日期：前一天（按当前全局本地时区，即雅加达日期）
            pushDate := time.Now().Add(-24 * time.Hour).Format("2006-01-02")
            log.Printf("[settlement-scheduler] 开始推送每日结算提醒，日期：%s", pushDate)

            if err := handlers.NotifyWecomSettlementReminder(pushDate); err != nil {
                log.Printf("[settlement-scheduler] 推送每日结算提醒失败：%v", err)
            } else {
                log.Printf("[settlement-scheduler] 推送每日结算提醒成功，日期：%s", pushDate)
            }
        }
    }()

    // 启动一个后台协程，每周一北京时间 08:00 推送一次上周 7 天的结算汇总周报
    go func() {
        bjLoc, err := time.LoadLocation("Asia/Shanghai")
        if err != nil {
            log.Printf("[weekly-scheduler] 加载 Asia/Shanghai 时区失败，将使用本地时区：%v", err)
            bjLoc = time.Local
        }

        for {
            nowBJ := time.Now().In(bjLoc)
            // 计算下一个周一 08:00（北京时间）
            // Go 的 Weekday：Sunday=0, Monday=1, ...
            daysUntilMonday := (int(time.Monday) - int(nowBJ.Weekday()) + 7) % 7
            nextMonday := time.Date(nowBJ.Year(), nowBJ.Month(), nowBJ.Day(), 8, 0, 0, 0, bjLoc).AddDate(0, 0, daysUntilMonday)
            if !nextMonday.After(nowBJ) {
                nextMonday = nextMonday.AddDate(0, 0, 7)
            }

            sleepDuration := nextMonday.Sub(nowBJ)
            log.Printf("[weekly-scheduler] 下次每周结算周报推送时间（北京时间）：%s", nextMonday.Format(time.RFC3339))
            time.Sleep(sleepDuration)

            // 触发时，按当前本地时区（雅加达）计算“上周 7 天”日期范围：昨天起往前推 6 天
            end := time.Now().Add(-24 * time.Hour)
            start := end.AddDate(0, 0, -6)
            startDate := start.Format("2006-01-02")
            endDate := end.Format("2006-01-02")

            log.Printf("[weekly-scheduler] 开始推送结算周报，日期范围：%s ~ %s", startDate, endDate)

            if err := handlers.NotifyWecomSettlementForRange(gdb, startDate, endDate, "weekly"); err != nil {
                log.Printf("[weekly-scheduler] 推送结算周报失败：%v", err)
            } else {
                log.Printf("[weekly-scheduler] 推送结算周报成功，日期范围：%s ~ %s", startDate, endDate)
            }
        }
    }()

    // 启动一个后台协程，每月 1 日北京时间 08:00 推送一次上一个自然月的结算汇总月报
    go func() {
        bjLoc, err := time.LoadLocation("Asia/Shanghai")
        if err != nil {
            log.Printf("[monthly-scheduler] 加载 Asia/Shanghai 时区失败，将使用本地时区：%v", err)
            bjLoc = time.Local
        }

        for {
            nowBJ := time.Now().In(bjLoc)

            // 计算下一个“每月 1 日 08:00”（北京时间）
            thisMonthFirst := time.Date(nowBJ.Year(), nowBJ.Month(), 1, 8, 0, 0, 0, bjLoc)
            var nextRun time.Time
            if nowBJ.Before(thisMonthFirst) {
                nextRun = thisMonthFirst
            } else {
                // 下个月 1 日 08:00
                nextRun = thisMonthFirst.AddDate(0, 1, 0)
            }

            sleepDuration := nextRun.Sub(nowBJ)
            log.Printf("[monthly-scheduler] 下次结算月报推送时间（北京时间）：%s", nextRun.Format(time.RFC3339))
            time.Sleep(sleepDuration)

            // 触发时，按当前本地时区（雅加达）计算“上一个自然月”的起止日期
            nowLocal := time.Now()
            lastMonth := nowLocal.AddDate(0, -1, 0)
            firstDay := time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, nowLocal.Location())
            firstNextMonth := firstDay.AddDate(0, 1, 0)
            lastDay := firstNextMonth.AddDate(0, 0, -1)

            startDate := firstDay.Format("2006-01-02")
            endDate := lastDay.Format("2006-01-02")

            log.Printf("[monthly-scheduler] 开始推送结算月报，日期范围：%s ~ %s", startDate, endDate)

            if err := handlers.NotifyWecomSettlementForRange(gdb, startDate, endDate, "monthly"); err != nil {
                log.Printf("[monthly-scheduler] 推送结算月报失败：%v", err)
            } else {
                log.Printf("[monthly-scheduler] 推送结算月报成功，日期范围：%s ~ %s", startDate, endDate)
            }
        }
    }()

    r.Run(":8080")
}
