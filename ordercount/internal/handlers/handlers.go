package handlers

import (
    "bytes"
    "encoding/csv"
    "fmt"
    "net/http"
    "os"
    "path/filepath"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    "ordercount/internal/models"
)

// 商品列表（简单分页/全部）
func ListProducts(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var products []models.Product
        if err := db.Order("created_at desc").Find(&products).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        // 根据当前登录用户角色，只返回与其角色对应的成本价
        roleVal, _ := c.Get("role")
        roleStr, _ := roleVal.(string)

        type item struct {
            ID        uint      `json:"id"`
            SKU       string    `json:"sku"`
            Name      string    `json:"name"`
            ImageURL  string    `json:"image_url"`
            // cost 为“当前角色可见的成本”
            Cost      float64   `json:"cost"`
            // 仅超级管理员需要看到各角色的成本详情，用于配置
            CostAdmin float64   `json:"cost_admin,omitempty"`
            CostStaff float64   `json:"cost_staff,omitempty"`
            CreatedAt time.Time `json:"created_at"`
            UpdatedAt time.Time `json:"updated_at"`
        }

        res := make([]item, 0, len(products))
        for _, p := range products {
            // 默认使用基础成本
            visibleCost := p.Cost
            switch roleStr {
            case "admin":
                if p.CostAdmin != 0 {
                    visibleCost = p.CostAdmin
                }
            case "staff":
                if p.CostStaff != 0 {
                    visibleCost = p.CostStaff
                }
            }

            it := item{
                ID:        p.ID,
                SKU:       p.SKU,
                Name:      p.Name,
                ImageURL:  p.ImageURL,
                Cost:      visibleCost,
                CreatedAt: p.CreatedAt,
                UpdatedAt: p.UpdatedAt,
            }
            // 超级管理员需要看到三个角色的成本明细，便于配置
            if roleStr == "superadmin" {
                it.CostAdmin = p.CostAdmin
                it.CostStaff = p.CostStaff
            }
            res = append(res, it)
        }

        c.JSON(http.StatusOK, res)
    }
}

// 创建或更新商品
func SaveProduct(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var body struct {
            ID        uint     `json:"id"`
            SKU       string   `json:"sku"`
            Name      string   `json:"name"`
            ImageURL  string   `json:"image_url"`
            // 对于非超级管理员，这里的 cost 表示“当前角色看到/维护的成本”
            Cost      float64  `json:"cost"`
            // 超级管理员可以一次性配置不同角色的成本
            CostAdmin *float64 `json:"cost_admin"`
            CostStaff *float64 `json:"cost_staff"`
        }
        if err := c.ShouldBindJSON(&body); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        if body.SKU == "" && body.Name == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "SKU或名称至少填写一项"})
            return
        }

        roleVal, _ := c.Get("role")
        roleStr, _ := roleVal.(string)

        var p models.Product
        if body.ID != 0 {
            if err := db.First(&p, body.ID).Error; err != nil {
                c.JSON(http.StatusNotFound, gin.H{"error": "商品不存在"})
                return
            }
        }
        p.SKU = body.SKU
        p.Name = body.Name
        p.ImageURL = body.ImageURL
        // 按角色写入对应的成本字段
        switch roleStr {
        case "superadmin":
            // 超级管理员可以配置所有角色成本
            p.Cost = body.Cost
            if body.CostAdmin != nil {
                p.CostAdmin = *body.CostAdmin
            }
            if body.CostStaff != nil {
                p.CostStaff = *body.CostStaff
            }
        case "admin":
            // 管理员只能维护自己的成本
            p.CostAdmin = body.Cost
        default:
            // 员工不允许直接修改成本；新建时确保成本字段为 0
            if body.ID == 0 {
                p.Cost = 0
                p.CostAdmin = 0
                p.CostStaff = 0
            }
        }

        if err := db.Save(&p).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, p)
    }
}

// 删除商品
func DeleteProduct(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := c.Param("id")
        if id == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "id required"})
            return
        }
        if err := db.Delete(&models.Product{}, id).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{"ok": true})
    }
}

// 上传商品图片，返回可访问的 URL
func UploadProductImage() gin.HandlerFunc {
    return func(c *gin.Context) {
        file, err := c.FormFile("file")
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
            return
        }

        // 确保目录存在
        uploadDir := "uploads/products"
        if err := os.MkdirAll(uploadDir, 0755); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        // 使用时间戳+原始扩展名生成文件名
        ext := filepath.Ext(file.Filename)
        if ext == "" {
            ext = ".jpg"
        }
        filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
        fullPath := filepath.Join(uploadDir, filename)

        if err := c.SaveUploadedFile(file, fullPath); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        url := "/uploads/products/" + filename
        c.JSON(http.StatusOK, gin.H{"url": url})
    }
}

// 保存每日结算记录
func SaveSettlement(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var body struct {
            Date        string  `json:"date"`
            Country     string  `json:"country"`
            Currency    string  `json:"currency"`
            SaleTotal   float64 `json:"sale_total"`
            AdCost      float64 `json:"ad_cost"`
            Exchange    float64 `json:"exchange"`
            GoodsCost   float64 `json:"goods_cost"`
            ShuaDanFee  float64 `json:"shua_dan_fee"`
            FixedCost   float64 `json:"fixed_cost"`
            Remark      string  `json:"remark"`
        }
        if err := c.ShouldBindJSON(&body); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        // 默认日期为今天
        if body.Date == "" {
            body.Date = time.Now().Format("2006-01-02")
        }

        // 计算各项费用（在后端保存一份计算凭证，和前端保持同样公式）
        adDeduction := (body.AdCost + body.AdCost*0.11) * body.Exchange
        platformFee := body.SaleTotal * 0.07
        profit := body.SaleTotal - adDeduction - body.GoodsCost - platformFee - body.ShuaDanFee - body.FixedCost

        rec := models.DailySettlement{
            Date:        body.Date,
            Country:     body.Country,
            Currency:    body.Currency,
            SaleTotal:   body.SaleTotal,
            AdCost:      body.AdCost,
            Exchange:    body.Exchange,
            GoodsCost:   body.GoodsCost,
            ShuaDanFee:  body.ShuaDanFee,
            FixedCost:   body.FixedCost,
            AdDeduction: adDeduction,
            PlatformFee: platformFee,
            Profit:      profit,
            Remark:      body.Remark,
            CreatedAt:   time.Now(),
        }

        if err := db.Create(&rec).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, rec)
    }
}

// 新增订单接口
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

// 按小时统计某天的销售金额
func HourlyStats(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        date := c.Query("date")
        if date == "" {
            date = time.Now().Format("2006-01-02")
        }
        rows, err := db.Raw(`
            SELECT HOUR(created_at) as hour, IFNULL(SUM(total_amount),0) as total
            FROM orders
            WHERE DATE(created_at) = ? AND product_name NOT IN (?, ?)
            GROUP BY hour
            ORDER BY hour
        `, date, "今日总额汇总", "今日总汇").Rows()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        defer rows.Close()
        res := make([]float64, 24)
        for rows.Next() {
            var hour int
            var total float64
            rows.Scan(&hour, &total)
            if hour >= 0 && hour < 24 {
                res[hour] = total
            }
        }
        c.JSON(http.StatusOK, gin.H{"date": date, "hourly": res})
    }
}

// 近N天每日销售金额（默认7天）
func DailyStats(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        daysStr := c.DefaultQuery("days", "7")
        days, err := strconv.Atoi(daysStr)
        if err != nil || days < 1 || days > 60 {
            days = 7
        }
        rows, err := db.Raw(`
            SELECT DATE(created_at) as day, IFNULL(SUM(total_amount),0) as total
            FROM orders
            WHERE created_at >= DATE_SUB(CURDATE(), INTERVAL ? DAY)
              AND product_name IN (?, ?)
            GROUP BY DATE(created_at)
            ORDER BY DATE(created_at)
        `, days-1, "今日总额汇总", "今日总汇").Rows()
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

// 按月统计某年每月销售金额
func MonthlyStats(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        year := c.Query("year")
        if year == "" {
            year = time.Now().Format("2006")
        }
        rows, err := db.Raw(`
            SELECT MONTH(created_at) as month, IFNULL(SUM(total_amount),0) as total
            FROM orders
            WHERE YEAR(created_at) = ?
              AND product_name NOT IN (?, ?)
            GROUP BY month
            ORDER BY month
        `, year, "今日总额汇总", "今日总汇").Rows()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        defer rows.Close()
        res := make([]float64, 12)
        for rows.Next() {
            var month int
            var total float64
            rows.Scan(&month, &total)
            if month >= 1 && month <= 12 {
                res[month-1] = total
            }
        }
        c.JSON(http.StatusOK, gin.H{"year": year, "monthly": res})
    }
}

// ImportOrders Excel 批量导入功能已下线，如需恢复可从历史版本找回实现

func TodaySales(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        today := time.Now().Format("2006-01-02")
        // 直接使用“今日总额汇总”记录中最近一次提交的总额，作为今日销售额（人民币）
        var o models.Order
        err := db.Where("product_name = ? AND DATE(created_at)=?", "今日总额汇总", today).
            Order("created_at desc").
            First(&o).Error
        if err != nil {
            if err == gorm.ErrRecordNotFound {
                // 今天还没有提交“今日总额汇总”，返回 0
                c.JSON(http.StatusOK, gin.H{"today": today, "total_amount": 0})
                return
            }
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{"today": today, "total_amount": o.TotalAmount})
    }
}

// TodayGoodsCost 计算今日货款成本：
// 按当天订单中各商品的数量 * 商品基础成本（Product.Cost）求和，排除“今日总额汇总”等数量为 0 的汇总记录。
func TodayGoodsCost(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        today := time.Now().Format("2006-01-02")
        var total float64

        // JOIN 订单与商品表，按 SKU 匹配，使用基础成本字段 Cost 计算货款成本
        err := db.Table("orders AS o").
            Joins("JOIN products AS p ON o.sku = p.sku").
            Where("DATE(o.created_at) = ? AND o.quantity > 0 AND o.product_name <> ?", today, "今日总额汇总").
            Select("IFNULL(SUM(o.quantity * p.cost), 0)").
            Scan(&total).Error
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "today":      today,
            "total_cost": total,
        })
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

// 按日期查询结算明细（可选传 date，不传则默认今天），返回该日所有结算记录
func ListSettlements(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        date := c.Query("date")
        if date == "" {
            date = time.Now().Format("2006-01-02")
        }

        country := c.Query("country")

        var list []models.DailySettlement
        q := db.Where("date = ?", date)
        if country != "" {
            q = q.Where("country = ?", country)
        }
        if err := q.Order("created_at asc").Find(&list).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{"date": date, "items": list})
    }
}

// AdDeductionDailyStats 近N天每日广告费折算（默认7天），基于每日结算表 DailySettlement
func AdDeductionDailyStats(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        daysStr := c.DefaultQuery("days", "7")
        days, err := strconv.Atoi(daysStr)
        if err != nil || days < 1 || days > 60 {
            days = 7
        }

        rows, err := db.Raw(`
            SELECT date as day, IFNULL(SUM(ad_deduction),0) as total
            FROM daily_settlements
            WHERE date >= DATE_SUB(CURDATE(), INTERVAL ? DAY)
            GROUP BY date
            ORDER BY date
        `, days-1).Rows()
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

// AdDeductionMonthlyStats 按月统计某年每月广告费折算
func AdDeductionMonthlyStats(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        year := c.Query("year")
        if year == "" {
            year = time.Now().Format("2006")
        }
        rows, err := db.Raw(`
            SELECT MONTH(date) as month, IFNULL(SUM(ad_deduction),0) as total
            FROM daily_settlements
            WHERE YEAR(date) = ?
            GROUP BY month
            ORDER BY month
        `, year).Rows()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        defer rows.Close()
        res := make([]float64, 12)
        for rows.Next() {
            var month int
            var total float64
            rows.Scan(&month, &total)
            if month >= 1 && month <= 12 {
                res[month-1] = total
            }
        }
        c.JSON(http.StatusOK, gin.H{"year": year, "monthly": res})
    }
}

func TopProducts(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        rows, err := db.Raw(`
            SELECT product_name, SUM(quantity) as total
            FROM orders
            WHERE product_name NOT IN (?, ?)
            GROUP BY product_name
            ORDER BY total DESC
            LIMIT 10
        `, "今日总额汇总", "今日总汇").Rows()
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

// 按日期查询订单记录（默认今天）
func ListOrders(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        date := c.Query("date")
        if date == "" {
            date = time.Now().Format("2006-01-02")
        }

        var orders []models.Order
        if err := db.Where("DATE(created_at) = ?", date).Order("created_at desc").Find(&orders).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{"date": date, "items": orders})
    }
}

// 更新单条订单记录
func UpdateOrder(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        idStr := c.Param("id")
        id, err := strconv.Atoi(idStr)
        if err != nil || id <= 0 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
            return
        }

        var payload struct {
            Country     *string  `json:"country"`
            ProductName *string  `json:"product_name"`
            SKU         *string  `json:"sku"`
            Quantity    *int     `json:"quantity"`
            TotalAmount *float64 `json:"total_amount"`
        }
        if err := c.ShouldBindJSON(&payload); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        var o models.Order
        if err := db.First(&o, id).Error; err != nil {
            if err == gorm.ErrRecordNotFound {
                c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
                return
            }
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        if err := db.Model(&o).Updates(payload).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        // 重新查询最新数据返回
        if err := db.First(&o, id).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, o)
    }
}

// 单独修改订单日期（用于补录订单时调整统计日期）
func UpdateOrderDate(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        idStr := c.Param("id")
        id, err := strconv.Atoi(idStr)
        if err != nil || id <= 0 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
            return
        }

        var body struct {
            Date string `json:"date"`
        }
        if err := c.ShouldBindJSON(&body); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        if body.Date == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "date required"})
            return
        }

        // 仅按日期调整，时间统一设为 00:00:00，当天统计按 DATE(created_at) 即可
        t, err := time.Parse("2006-01-02", body.Date)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, want YYYY-MM-DD"})
            return
        }

        var o models.Order
        if err := db.First(&o, id).Error; err != nil {
            if err == gorm.ErrRecordNotFound {
                c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
                return
            }
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        o.CreatedAt = t
        if err := db.Save(&o).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, o)
    }
}

// 删除单条订单记录
func DeleteOrder(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        idStr := c.Param("id")
        id, err := strconv.Atoi(idStr)
        if err != nil || id <= 0 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
            return
        }
        if err := db.Delete(&models.Order{}, id).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{"deleted": true})
    }
}
