package handlers

import (
    "bytes"
    "encoding/csv"
    "fmt"
    "net/http"
    "strconv"
    "time"
    "math/rand"

    "github.com/gin-gonic/gin"
    "github.com/xuri/excelize/v2"
    "gorm.io/gorm"

    "ordercount/internal/models"
    exch "ordercount/internal/utils"
)

func PostOrder(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var o models.Order
        if err := c.ShouldBindJSON(&o); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        if o.CreatedAt.IsZero() {
            o.CreatedAt = time.Now()
        }
        if err := db.Create(&o).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, o)
    }
}

func ImportOrders(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        file, err := c.FormFile("file")
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
            return
        }

        f, err := file.Open()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        defer f.Close()

        // 使用 excelize 读取
        xf, err := excelize.OpenReader(f)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid excel file"})
            return
        }

        sheet := xf.GetSheetName(0)
        rows, err := xf.GetRows(sheet)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        var created int
        rand.Seed(time.Now().UnixNano())
        for i, row := range rows {
            if i == 0 {
                continue // 跳过表头
            }
            // 期望列: country,platform,order_no,product_name,sku,quantity,total_amount[,created_at]
            if len(row) < 7 {
                continue
            }
            qty, _ := strconv.Atoi(row[5])
            amt, _ := strconv.ParseFloat(row[6], 64)
            // 随机金额逻辑：如果所有金额都一样或为0，则自动生成随机金额
            if amt == 0 || amt == 100.0 {
                switch row[0] {
                case "菲律宾":
                    amt = float64(rand.Intn(900)+100) // 100~999 PHP
                case "印尼":
                    amt = float64(rand.Intn(90000)+10000) // 10000~99999 IDR
                case "马来西亚":
                    amt = float64(rand.Intn(90)+10) // 10~99 MYR
                default:
                    amt = float64(rand.Intn(900)+100)
                }
            }
            currency := "PHP"
            switch row[0] {
            case "菲律宾":
                currency = "PHP"
            case "印尼":
                currency = "IDR"
            case "马来西亚":
                currency = "MYR"
            }
            o := models.Order{
                Country:     row[0],
                Platform:    row[1],
                OrderNo:     row[2],
                ProductName: row[3],
                SKU:         row[4],
                Quantity:    qty,
                TotalAmount: amt,
                Currency:    currency,
                CreatedAt:   time.Now(),
            }
            if len(row) >= 8 && row[7] != "" {
                if t, err := time.Parse("2006-01-02", row[7]); err == nil {
                    o.CreatedAt = t
                }
            }
            // 自动设置币种
            switch o.Country {
            case "菲律宾":
                o.Currency = "PHP"
            case "印尼":
                o.Currency = "IDR"
            case "马来西亚":
                o.Currency = "MYR"
            }
            if err := db.Create(&o).Error; err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
            }
            created++
        }
        c.JSON(http.StatusOK, gin.H{"created": created})
    }
}

func TodaySales(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        today := time.Now().Format("2006-01-02")
        showCNY := c.Query("cny") == "1"
        // 查询各币种金额
        type Item struct {
            Currency  string  `json:"Currency"`
            Sum       float64 `json:"Sum"`
            CNYAmount float64 `json:"cny_amount,omitempty"`
        }
        var items []Item
        if err := db.Raw("SELECT currency, IFNULL(SUM(total_amount),0) AS sum FROM orders WHERE DATE(created_at)=? GROUP BY currency", today).Scan(&items).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        rates, err := exch.GetRates()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "汇率获取失败"})
            return
        }
        var sumCNY float64
        for i := range items {
            rate := rates[items[i].Currency]
            switch items[i].Currency {
            case "IDR", "PHP", "MYR":
                if rate != 0 {
                    items[i].CNYAmount = items[i].Sum / rate
                }
            default:
                items[i].CNYAmount = items[i].Sum
            }
            sumCNY += items[i].CNYAmount
        }
        if !showCNY {
            // 返回各币种明细（含cny_amount），不返回 total_amount 字段
            c.JSON(http.StatusOK, gin.H{"today": today, "currencies": items})
            return
        }
        // 返回各币种明细和人民币总和
        c.JSON(http.StatusOK, gin.H{"today": today, "total_amount": sumCNY, "currencies": items})
    }
}

func ExportReport(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var orders []models.Order
        if err := db.Order("created_at desc").Find(&orders).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        buf := &bytes.Buffer{}
        w := csv.NewWriter(buf)
        w.Write([]string{"country", "platform", "order_no", "product_name", "sku", "quantity", "total_amount", "created_at"})
        for _, o := range orders {
            w.Write([]string{
                o.Country,
                o.Platform,
                o.OrderNo,
                o.ProductName,
                o.SKU,
                strconv.Itoa(o.Quantity),
                fmt.Sprintf("%.2f", o.TotalAmount),
                o.CreatedAt.Format("2006-01-02 15:04:05"),
            })
        }
        w.Flush()

        c.Header("Content-Disposition", "attachment; filename=orders_report.csv")
        c.Data(http.StatusOK, "text/csv", buf.Bytes())
    }
}

func SalesTrend(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 返回最近 30 天每日销售总额
        rows, err := db.Raw(`
            SELECT DATE(created_at) as day, IFNULL(SUM(total_amount),0) as total
            FROM orders
            WHERE created_at >= DATE_SUB(CURDATE(), INTERVAL 29 DAY)
            GROUP BY DATE(created_at)
            ORDER BY DATE(created_at)
        `).Rows()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        defer rows.Close()
        type Item struct{ Day string; Total float64 }
        var res []Item
        for rows.Next() {
            var it Item
            rows.Scan(&it.Day, &it.Total)
            res = append(res, it)
        }
        c.JSON(http.StatusOK, res)
    }
}

func TopProducts(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        rows, err := db.Raw(`
            SELECT product_name, SUM(total_amount) as total
            FROM orders
            GROUP BY product_name
            ORDER BY total DESC
            LIMIT 10
        `).Rows()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        defer rows.Close()
        type Item struct{ ProductName string; Total float64 }
        var res []Item
        for rows.Next() {
            var it Item
            rows.Scan(&it.ProductName, &it.Total)
            res = append(res, it)
        }
        c.JSON(http.StatusOK, res)
    }
}
