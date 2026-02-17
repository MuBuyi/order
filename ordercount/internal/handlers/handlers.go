package handlers

import (
    "bytes"
    "encoding/csv"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "path/filepath"
    "sort"
    "strconv"
    "strings"
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
        // 仅超级管理员和管理员可以新增或修改商品
        if roleStr != "superadmin" && roleStr != "admin" {
            c.JSON(http.StatusForbidden, gin.H{"error": "仅管理员或超级管理员可以维护商品信息"})
            return
        }

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
        // 仅超级管理员可以删除商品
        roleVal, _ := c.Get("role")
        if roleStr, ok := roleVal.(string); !ok || roleStr != "superadmin" {
            c.JSON(http.StatusForbidden, gin.H{"error": "仅超级管理员可以删除商品"})
            return
        }
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

// 店铺管理：增删改查

// ListStores 列出所有店铺
func ListStores(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var list []models.Store

        // 超级管理员可以看到所有店铺；其他角色只能看到被授权的店铺
        roleVal, _ := c.Get("role")
        userIDVal, _ := c.Get("userID")

        if roleStr, ok := roleVal.(string); ok && roleStr == "superadmin" {
            if err := db.Order("created_at asc").Find(&list).Error; err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
            }
        } else {
            uid, ok := userIDVal.(uint)
            if !ok || uid == 0 {
                c.JSON(http.StatusUnauthorized, gin.H{"error": "未获取到用户信息"})
                return
            }
            if err := db.Joins("JOIN store_users su ON su.store_id = stores.id").
                Where("su.user_id = ?", uid).
                Order("stores.created_at asc").
                Find(&list).Error; err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
            }
        }

        c.JSON(http.StatusOK, gin.H{"items": list})
    }
}

// SaveStore 新增或更新店铺
func SaveStore(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 仅超级管理员可以新增或修改店铺
        if roleVal, ok := c.Get("role"); !ok || roleVal != "superadmin" {
            c.JSON(http.StatusForbidden, gin.H{"error": "仅超级管理员可以维护店铺信息"})
            return
        }

        var body models.Store
        if err := c.ShouldBindJSON(&body); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
            return
        }
        // 国家默认印尼
        if strings.TrimSpace(body.Country) == "" {
            body.Country = "印尼"
        }
        if body.Platform == "" || body.Name == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "平台和店铺名不能为空"})
            return
        }

        // 根据是否有 ID 判断是新增还是更新
        if body.ID == 0 {
            if err := db.Create(&body).Error; err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
            }
        } else {
            var existing models.Store
            if err := db.First(&existing, body.ID).Error; err != nil {
                if err == gorm.ErrRecordNotFound {
                    c.JSON(http.StatusNotFound, gin.H{"error": "店铺不存在"})
                } else {
                    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                }
                return
            }

            existing.Platform = body.Platform
            existing.Name = body.Name
            existing.Country = body.Country
            existing.LoginAccount = body.LoginAccount
            existing.LoginPassword = body.LoginPassword
            existing.Phone = body.Phone
            existing.Email = body.Email

            if err := db.Save(&existing).Error; err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
            }
            body = existing
        }

        c.JSON(http.StatusOK, body)
    }
}

// DeleteStore 删除店铺
func DeleteStore(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 仅超级管理员可以删除店铺
        if roleVal, ok := c.Get("role"); !ok || roleVal != "superadmin" {
            c.JSON(http.StatusForbidden, gin.H{"error": "仅超级管理员可以删除店铺"})
            return
        }

        id := c.Param("id")
        if id == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "id required"})
            return
        }
        if err := db.Delete(&models.Store{}, id).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        // 同时清理授权关系
        if err := db.Where("store_id = ?", id).Delete(&models.StoreUser{}).Error; err != nil {
            // 授权记录删除失败仅记录日志，不影响主流程
            fmt.Printf("failed to delete store_users for store %s: %v\n", id, err)
        }
        c.JSON(http.StatusOK, gin.H{"ok": true})
    }
}

// GetStoreUsers 获取某个店铺已授权的用户ID列表（仅超级管理员）
func GetStoreUsers(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        if roleVal, ok := c.Get("role"); !ok || roleVal != "superadmin" {
            c.JSON(http.StatusForbidden, gin.H{"error": "仅超级管理员可以查看店铺授权"})
            return
        }

        id := c.Param("id")
        if id == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "id required"})
            return
        }

        var rels []models.StoreUser
        if err := db.Where("store_id = ?", id).Find(&rels).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        userIDs := make([]uint, 0, len(rels))
        for _, r := range rels {
            userIDs = append(userIDs, r.UserID)
        }
        c.JSON(http.StatusOK, gin.H{"user_ids": userIDs})
    }
}

// UpdateStoreUsers 更新店铺授权的用户列表（仅超级管理员）
func UpdateStoreUsers(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        if roleVal, ok := c.Get("role"); !ok || roleVal != "superadmin" {
            c.JSON(http.StatusForbidden, gin.H{"error": "仅超级管理员可以维护店铺授权"})
            return
        }

        id := c.Param("id")
        if id == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "id required"})
            return
        }

        var body struct {
            UserIDs []uint `json:"user_ids"`
        }
        if err := c.ShouldBindJSON(&body); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
            return
        }

        // 先删除旧的授权关系
        if err := db.Where("store_id = ?", id).Delete(&models.StoreUser{}).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        // 再批量插入新的授权关系
        if len(body.UserIDs) > 0 {
            rels := make([]models.StoreUser, 0, len(body.UserIDs))
            for _, uid := range body.UserIDs {
                rels = append(rels, models.StoreUser{StoreID: 0, UserID: uid})
            }
            // 这里需要把 path 参数转换为 uint
            var store models.Store
            if err := db.Where("id = ?", id).First(&store).Error; err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
            }
            for i := range rels {
                rels[i].StoreID = store.ID
            }
            if err := db.Create(&rels).Error; err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
            }
        }

        c.JSON(http.StatusOK, gin.H{"ok": true})
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

        // 需求：同一天可以多次计算利润，但最终只保留最后一次的结果。
        // 实现方式：按 Date + Country 查找已存在记录，若有则覆盖更新，否则创建新记录。

        var existing models.DailySettlement
        // 按当前登录用户 + 日期 + 国家 唯一一条记录
        userIDVal, _ := c.Get("userID")
        uid, _ := userIDVal.(uint)

        // 如果没有取到 userID，视为未登录或上下文异常
        if uid == 0 {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "未获取到用户信息"})
            return
        }

        err := db.Where("date = ? AND country = ? AND user_id = ?", body.Date, body.Country, uid).First(&existing).Error

        if err != nil && err != gorm.ErrRecordNotFound {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        // 用本次计算结果覆盖字段
        existing.Date = body.Date
        existing.UserID = uid
        existing.Country = body.Country
        existing.Currency = body.Currency
        existing.SaleTotal = body.SaleTotal
        existing.AdCost = body.AdCost
        existing.Exchange = body.Exchange
        existing.GoodsCost = body.GoodsCost
        existing.ShuaDanFee = body.ShuaDanFee
        existing.FixedCost = body.FixedCost
        existing.AdDeduction = adDeduction
        existing.PlatformFee = platformFee
        existing.Profit = profit
        existing.Remark = body.Remark

        // 如果之前没有记录，CreatedAt 会由 GORM 自动填充当前时间；
        // 如果有记录，则保留原来的 ID，仅更新时间和字段。
        if err == gorm.ErrRecordNotFound {
            if err := db.Create(&existing).Error; err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
            }
        } else {
            if err := db.Save(&existing).Error; err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
            }
        }

        // 每次保存结算，即视为一次手动推送当日利润汇总到企微群
        if err := NotifyWecomSettlementForDate(db, existing.Date); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "保存结算成功，但推送企业微信失败: " + err.Error()})
            return
        }

        c.JSON(http.StatusOK, existing)
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
        // 将当前登录用户写入订单，用于后续按用户过滤
        userIDVal, _ := c.Get("userID")
        if uid, ok := userIDVal.(uint); ok && uid != 0 {
            o.UserID = uid
        } else {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "未获取到用户信息"})
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

// 按日期查询结算明细：
// - 可选传 date、country 进行过滤；
// - 支持分页参数 page、page_size；
// - 不传 date 时表示查询所有日期的结算记录，并按日期倒序分页返回。
func ListSettlements(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        date := c.Query("date")
        country := c.Query("country")

        // 分页参数，默认第 1 页，每页 10 条
        pageStr := c.DefaultQuery("page", "1")
        sizeStr := c.DefaultQuery("page_size", "10")
        page, err := strconv.Atoi(pageStr)
        if err != nil || page < 1 {
            page = 1
        }
        pageSize, err := strconv.Atoi(sizeStr)
        if err != nil || pageSize < 1 || pageSize > 100 {
            pageSize = 10
        }

        var list []models.DailySettlement
        q := db.Model(&models.DailySettlement{})
        if date != "" {
            q = q.Where("date = ?", date)
        }
        if country != "" {
            q = q.Where("country = ?", country)
        }

        // 权限控制：
        // - 超级管理员可以看到所有用户的结算记录；
        // - 其他角色只能看到自己创建的结算记录。
        roleVal, _ := c.Get("role")
        roleStr, _ := roleVal.(string)
        userIDVal, _ := c.Get("userID")
        if roleStr != "superadmin" {
            uid, ok := userIDVal.(uint)
            if !ok || uid == 0 {
                c.JSON(http.StatusUnauthorized, gin.H{"error": "未获取到用户信息"})
                return
            }
            q = q.Where("user_id = ?", uid)
        }

        // 先统计总数
        var total int64
        if err := q.Count(&total).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        // 按日期倒序、同一天内按创建时间正序，便于查看最新记录
        offset := (page - 1) * pageSize
        if err := q.Order("date desc").Order("created_at asc").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        // 为避免前端因浏览器本地时区导致日期偏移，这里将时间字段格式化为不带时区的本地时间字符串
        type item struct {
            ID           uint    `json:"id"`
            Date         string  `json:"date"`
            Country      string  `json:"country"`
            Currency     string  `json:"currency"`
            SaleTotal    float64 `json:"sale_total"`
            AdCost       float64 `json:"ad_cost"`
            Exchange     float64 `json:"exchange"`
            GoodsCost    float64 `json:"goods_cost"`
            ShuaDanFee   float64 `json:"shua_dan_fee"`
            FixedCost    float64 `json:"fixed_cost"`
            AdDeduction  float64 `json:"ad_deduction"`
            PlatformFee  float64 `json:"platform_fee"`
            Profit       float64 `json:"profit"`
            Remark       string  `json:"remark"`
            CreatedAtStr string  `json:"created_at"`
        }

        items := make([]item, 0, len(list))
        for _, s := range list {
            it := item{
                ID:          s.ID,
                Date:        s.Date,
                Country:     s.Country,
                Currency:    s.Currency,
                SaleTotal:   s.SaleTotal,
                AdCost:      s.AdCost,
                Exchange:    s.Exchange,
                GoodsCost:   s.GoodsCost,
                ShuaDanFee:  s.ShuaDanFee,
                FixedCost:   s.FixedCost,
                AdDeduction: s.AdDeduction,
                PlatformFee: s.PlatformFee,
                Profit:      s.Profit,
                Remark:      s.Remark,
            }
            if !s.CreatedAt.IsZero() {
                it.CreatedAtStr = s.CreatedAt.Format("2006-01-02 15:04:05")
            }
            items = append(items, it)
        }

        c.JSON(http.StatusOK, gin.H{
            "date":      date,
            "items":     items,
            "total":     total,
            "page":      page,
            "page_size": pageSize,
        })
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

        // 权限控制：
        // - 超级管理员可以看到所有用户在该日期的订单；
        // - 其他角色只能看到自己录入的订单。
        roleVal, _ := c.Get("role")
        roleStr, _ := roleVal.(string)
        userIDVal, _ := c.Get("userID")

        q := db.Model(&models.Order{}).Where("DATE(created_at) = ?", date)
        if roleStr != "superadmin" {
            uid, ok := userIDVal.(uint)
            if !ok || uid == 0 {
                c.JSON(http.StatusUnauthorized, gin.H{"error": "未获取到用户信息"})
                return
            }
            q = q.Where("user_id = ?", uid)
        }

        if err := q.Order("created_at desc").Find(&orders).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        // 为避免浏览器按本地时区偏移导致日期从 16 号变成 17 号，
        // 这里将时间字段格式化为不带时区的本地时间字符串再返回给前端。
        type item struct {
            ID          uint    `json:"id"`
            Country     string  `json:"country"`
            Platform    string  `json:"platform"`
            OrderNo     string  `json:"order_no"`
            ProductName string  `json:"product_name"`
            SKU         string  `json:"sku"`
            Quantity    int     `json:"quantity"`
            TotalAmount float64 `json:"total_amount"`
            Currency    string  `json:"currency"`
            CreatedAt   string  `json:"created_at"`
        }

        items := make([]item, 0, len(orders))
        for _, o := range orders {
            it := item{
                ID:          o.ID,
                Country:     o.Country,
                Platform:    o.Platform,
                OrderNo:     o.OrderNo,
                ProductName: o.ProductName,
                SKU:         o.SKU,
                Quantity:    o.Quantity,
                TotalAmount: o.TotalAmount,
                Currency:    o.Currency,
            }
            if !o.CreatedAt.IsZero() {
                it.CreatedAt = o.CreatedAt.Format("2006-01-02 15:04:05")
            }
            items = append(items, it)
        }

        c.JSON(http.StatusOK, gin.H{"date": date, "items": items})
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

// NotifyWecomTodayOrders HTTP 接口：将指定日期（默认今天）的订单汇总通过企业微信机器人推送到企微群
// 推送内容包含：日期、各国家下商品的数量汇总，以及当日总额汇总。
func NotifyWecomTodayOrders(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var body struct {
            Date string `json:"date"`
        }
        if err := c.ShouldBindJSON(&body); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
            return
        }

        date := body.Date
        if date == "" {
            date = time.Now().Format("2006-01-02")
        }

        if err := NotifyWecomOrdersForDate(db, date); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "ok":      true,
            "message": "已推送到企业微信群",
        })
    }
}

// NotifyWecomOrdersForDate 是实际执行推送的逻辑，供 HTTP 接口和定时任务复用
func NotifyWecomOrdersForDate(db *gorm.DB, date string) error {
    if date == "" {
        date = time.Now().Format("2006-01-02")
    }

    // 聚合指定日期下、按国家+商品名统计的数量（排除今日总额汇总等虚拟商品，排除数量<=0 的记录）
    rows, err := db.Raw(`
        SELECT country, product_name, SUM(quantity) as qty
        FROM orders
        WHERE DATE(created_at) = ?
          AND quantity > 0
          AND product_name NOT IN (?, ?)
        GROUP BY country, product_name
        ORDER BY country, qty DESC
    `, date, "今日总额汇总", "今日总汇").Rows()
    if err != nil {
        return err
    }
    defer rows.Close()

    type item struct {
        ProductName string
        Qty         int
    }
    countryMap := make(map[string][]item)
    for rows.Next() {
        var country, name string
        var qty int
        if err := rows.Scan(&country, &name, &qty); err != nil {
            return err
        }
        countryMap[country] = append(countryMap[country], item{ProductName: name, Qty: qty})
    }

    // 获取当日“今日总额汇总”记录的总额
    var totalAmount float64
    var o models.Order
    if err := db.Where("product_name = ? AND DATE(created_at)=?", "今日总额汇总", date).
        Order("created_at desc").
        First(&o).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            // 没有汇总记录，保持 totalAmount=0 即可
        } else {
            return err
        }
    } else {
        totalAmount = o.TotalAmount
    }

    // 按国家名排序，保证输出稳定
    countries := make([]string, 0, len(countryMap))
    for k := range countryMap {
        countries = append(countries, k)
    }
    sort.Strings(countries)

    // 组装企业微信 markdown 消息
    var buf bytes.Buffer
    fmt.Fprintf(&buf, "【订单日报】%s\n", date)
    if totalAmount > 0 {
        fmt.Fprintf(&buf, "\n今日总额（人民币）：**￥%.2f**\n", totalAmount)
    } else {
        buf.WriteString("\n今日总额（人民币）：**暂无\"今日总额汇总\"记录**\n")
    }

    if len(countries) == 0 {
        buf.WriteString("\n今日暂无明细订单记录。\n")
    } else {
        for _, country := range countries {
            buf.WriteString("\n> ")
            buf.WriteString(country)
            buf.WriteString("：\n")
            for _, it := range countryMap[country] {
                fmt.Fprintf(&buf, "> - %s：%d 单\n", it.ProductName, it.Qty)
            }
        }
    }

    content := buf.String()

    // 读取企微机器人地址：优先使用 config.yaml 注入的 WecomWebhook，
    // 如未配置则回退到环境变量 WECHAT_ROBOT_WEBHOOK / WECHAT_ROBOT_KEY。
    webhook := WecomWebhook
    if webhook == "" {
        webhook = os.Getenv("WECHAT_ROBOT_WEBHOOK")
        if webhook == "" {
            if key := os.Getenv("WECHAT_ROBOT_KEY"); key != "" {
                webhook = fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=%s", key)
            }
        }
    }
    if webhook == "" {
        return fmt.Errorf("未配置企业微信机器人地址，请在 config.yaml 的 wecom.webhook 中配置，或设置 WECHAT_ROBOT_WEBHOOK / WECHAT_ROBOT_KEY 环境变量")
    }

    payload := map[string]any{
        "msgtype": "markdown",
        "markdown": map[string]any{
            "content": content,
        },
    }
    b, err := json.Marshal(payload)
    if err != nil {
        return err
    }

    resp, err := http.Post(webhook, "application/json", bytes.NewReader(b))
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("企业微信返回状态码 %d", resp.StatusCode)
    }

    return nil
}

// NotifyWecomSettlementForDate 将指定日期的每日结算记录，通过企业微信机器人以简易表格形式推送到企微群。
// 表头字段：国家、时间、销售额、广告成本、货款成本、平台手续费、刷单费用、固定成本、利润。
// 时间只展示到天（YYYY-MM-DD），不包含小时和分钟。
func NotifyWecomSettlementForDate(db *gorm.DB, date string) error {
    if date == "" {
        date = time.Now().Format("2006-01-02")
    }

    var list []models.DailySettlement
    if err := db.Where("date = ?", date).
        Order("country asc, created_at asc").
        Find(&list).Error; err != nil {
        return err
    }

    var buf bytes.Buffer
    fmt.Fprintf(&buf, "【每日结算利润汇总】%s\n", date)

    if len(list) == 0 {
        buf.WriteString("\n今日暂无结算记录。\n")
    } else {
        // 先做整体汇总，便于在群里快速看到当天总情况
        // 这里的广告成本使用折算后的 AdDeduction 字段，而不是原始广告费 AdCost
        var totalSale, totalAdDeduction, totalGoodsCost, totalPlatformFee, totalShuaDanFee, totalFixedCost, totalProfit float64
        for _, s := range list {
            totalSale += s.SaleTotal
            totalAdDeduction += s.AdDeduction
            totalGoodsCost += s.GoodsCost
            totalPlatformFee += s.PlatformFee
            totalShuaDanFee += s.ShuaDanFee
            totalFixedCost += s.FixedCost
            totalProfit += s.Profit
        }

        buf.WriteString("\n**整体汇总（人民币）**\n")
        fmt.Fprintf(&buf, "> 销售额：￥%.2f\n", totalSale)
        fmt.Fprintf(&buf, "> 广告成本：￥%.2f，货款成本：￥%.2f\n", totalAdDeduction, totalGoodsCost)
        fmt.Fprintf(&buf, "> 平台手续费：￥%.2f，刷单费用：￥%.2f，固定成本：￥%.2f\n", totalPlatformFee, totalShuaDanFee, totalFixedCost)
        fmt.Fprintf(&buf, "> **利润合计：￥%.2f**\n", totalProfit)

        buf.WriteString("\n**按国家明细**\n")
        for _, s := range list {
            // 时间列只展示日期，不包含时分
            day := s.Date
            if day == "" && !s.CreatedAt.IsZero() {
                day = s.CreatedAt.Format("2006-01-02")
            }

            profitStr := fmt.Sprintf("￥%.2f", s.Profit)
            // 亏损时显式带负号，便于在群里一眼识别
            if s.Profit < 0 {
                profitStr = fmt.Sprintf("-￥%.2f", -s.Profit)
            }

            fmt.Fprintf(&buf, "\n> %s %s\n", s.Country, day)
            fmt.Fprintf(&buf, "> 销售：￥%.2f | 广告成本：￥%.2f | 货款：￥%.2f\n", s.SaleTotal, s.AdDeduction, s.GoodsCost)
            fmt.Fprintf(&buf, "> 平台：￥%.2f | 刷单：￥%.2f | 固定：￥%.2f | 利润：**%s**\n",
                s.PlatformFee,
                s.ShuaDanFee,
                s.FixedCost,
                profitStr,
            )
        }
    }

    content := buf.String()

    // 读取企微机器人地址：优先使用 config.yaml 注入的 WecomWebhook，
    // 如未配置则回退到环境变量 WECHAT_ROBOT_WEBHOOK / WECHAT_ROBOT_KEY。
    webhook := WecomWebhook
    if webhook == "" {
        webhook = os.Getenv("WECHAT_ROBOT_WEBHOOK")
        if webhook == "" {
            if key := os.Getenv("WECHAT_ROBOT_KEY"); key != "" {
                webhook = fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=%s", key)
            }
        }
    }
    if webhook == "" {
        return fmt.Errorf("未配置企业微信机器人地址，请在 config.yaml 的 wecom.webhook 中配置，或设置 WECHAT_ROBOT_WEBHOOK / WECHAT_ROBOT_KEY 环境变量")
    }

    payload := map[string]any{
        "msgtype": "markdown",
        "markdown": map[string]any{
            "content": content,
        },
    }
    b, err := json.Marshal(payload)
    if err != nil {
        return err
    }

    resp, err := http.Post(webhook, "application/json", bytes.NewReader(b))
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("企业微信返回状态码 %d", resp.StatusCode)
    }

    return nil
}

// NotifyWecomSettlementReminder 在企微群里发送结算提醒，而不是直接推送结算汇总。
// 一般用于每天固定时间提醒相关同事尽快完成某日的结算录入与检查。
func NotifyWecomSettlementReminder(date string) error {
    if date == "" {
        date = time.Now().Format("2006-01-02")
    }

    // 这里的日期通常是“昨天”的业务日期，由调用方决定具体是哪一天。
    var buf bytes.Buffer
    fmt.Fprintf(&buf, "【每日结算提醒】%s\n", date)
    buf.WriteString("\n请尽快完成结算核对\n")

    content := buf.String()

    // 读取企微机器人地址：优先使用 config.yaml 注入的 WecomWebhook，
    // 如未配置则回退到环境变量 WECHAT_ROBOT_WEBHOOK / WECHAT_ROBOT_KEY。
    webhook := WecomWebhook
    if webhook == "" {
        webhook = os.Getenv("WECHAT_ROBOT_WEBHOOK")
        if webhook == "" {
            if key := os.Getenv("WECHAT_ROBOT_KEY"); key != "" {
                webhook = fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=%s", key)
            }
        }
    }
    if webhook == "" {
        return fmt.Errorf("未配置企业微信机器人地址，请在 config.yaml 的 wecom.webhook 中配置，或设置 WECHAT_ROBOT_WEBHOOK / WECHAT_ROBOT_KEY 环境变量")
    }

    payload := map[string]any{
        "msgtype": "markdown",
        "markdown": map[string]any{
            "content": content,
        },
    }
    b, err := json.Marshal(payload)
    if err != nil {
        return err
    }

    resp, err := http.Post(webhook, "application/json", bytes.NewReader(b))
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("企业微信返回状态码 %d", resp.StatusCode)
    }

    return nil
}

// TriggerSettlementPush 主动触发某一天的结算利润汇总推送
// 支持通过 query/body 传入 date（YYYY-MM-DD），不传则默认昨天的日期。
func TriggerSettlementPush(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 仅超级管理员允许手动触发推送
        roleVal, _ := c.Get("role")
        roleStr, _ := roleVal.(string)
        if roleStr != "superadmin" {
            c.JSON(http.StatusForbidden, gin.H{"error": "仅超级管理员可以主动推送结算汇总"})
            return
        }

        // 优先从 query 获取日期，其次从 JSON body 获取
        date := c.Query("date")
        if date == "" {
            var body struct{ Date string `json:"date"` }
            if err := c.ShouldBindJSON(&body); err == nil && body.Date != "" {
                date = body.Date
            }
        }
        if date == "" {
            // 默认推送“昨天”的结算
            date = time.Now().Add(-24 * time.Hour).Format("2006-01-02")
        }

        if err := NotifyWecomSettlementForDate(db, date); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{"ok": true, "date": date})
    }
}

// NotifyWecomSettlementForRange 将一段日期区间内的每日结算记录汇总成报表，并推送到企业微信。
// reportType 可为 "weekly" 或 "monthly"，用于调整标题文案。
func NotifyWecomSettlementForRange(db *gorm.DB, startDate, endDate, reportType string) error {
    if startDate == "" || endDate == "" {
        return fmt.Errorf("startDate 和 endDate 不能为空")
    }

    var list []models.DailySettlement
    if err := db.Where("date >= ? AND date <= ?", startDate, endDate).
        Order("date asc, country asc, created_at asc").
        Find(&list).Error; err != nil {
        return err
    }

    var title string
    switch reportType {
    case "weekly":
        title = "【每周结算利润汇总】"
    case "monthly":
        title = "【每月结算利润汇总】"
    default:
        title = "【结算利润汇总】"
    }

    var buf bytes.Buffer
    fmt.Fprintf(&buf, "%s%s ~ %s\n", title, startDate, endDate)

    if len(list) == 0 {
        buf.WriteString("\n本期内暂无结算记录。\n")
    } else {
        // 一、明细列表
        buf.WriteString("\n**一、明细列表**\n")
        buf.WriteString("\n| 日期 | 国家 | 销售额 | 广告成本 | 货款成本 | 平台手续费 | 刷单费用 | 固定成本 | 利润 |\n")
        buf.WriteString("| --- | --- | --- | --- | --- | --- | --- | --- | --- |\n")

        type agg struct {
            SaleTotal   float64
            AdCost      float64
            GoodsCost   float64
            PlatformFee float64
            ShuaDanFee  float64
            FixedCost   float64
            Profit      float64
        }

        total := agg{}
        countryAgg := make(map[string]*agg)
        dayProfit := make(map[string]float64)

        for _, s := range list {
            day := s.Date
            if day == "" && !s.CreatedAt.IsZero() {
                day = s.CreatedAt.Format("2006-01-02")
            }

            fmt.Fprintf(&buf,
                "| %s | %s | %.2f | %.2f | %.2f | %.2f | %.2f | %.2f | %.2f |\n",
                day,
                s.Country,
                s.SaleTotal,
                s.AdCost,
                s.GoodsCost,
                s.PlatformFee,
                s.ShuaDanFee,
                s.FixedCost,
                s.Profit,
            )

            // 汇总整体
            total.SaleTotal += s.SaleTotal
            total.AdCost += s.AdCost
            total.GoodsCost += s.GoodsCost
            total.PlatformFee += s.PlatformFee
            total.ShuaDanFee += s.ShuaDanFee
            total.FixedCost += s.FixedCost
            total.Profit += s.Profit

            // 按国家汇总
            ca := countryAgg[s.Country]
            if ca == nil {
                ca = &agg{}
                countryAgg[s.Country] = ca
            }
            ca.SaleTotal += s.SaleTotal
            ca.AdCost += s.AdCost
            ca.GoodsCost += s.GoodsCost
            ca.PlatformFee += s.PlatformFee
            ca.ShuaDanFee += s.ShuaDanFee
            ca.FixedCost += s.FixedCost
            ca.Profit += s.Profit

            // 按日期统计利润
            dayProfit[day] += s.Profit
        }

        // 二、总体汇总
        buf.WriteString("\n**二、总体汇总（人民币）**\n")
        fmt.Fprintf(&buf, "\n- 期间总销售额：%.2f\n", total.SaleTotal)
        fmt.Fprintf(&buf, "- 期间总广告成本：%.2f\n", total.AdCost)
        fmt.Fprintf(&buf, "- 期间总货款成本：%.2f\n", total.GoodsCost)
        fmt.Fprintf(&buf, "- 期间总平台手续费：%.2f\n", total.PlatformFee)
        fmt.Fprintf(&buf, "- 期间总刷单费用：%.2f\n", total.ShuaDanFee)
        fmt.Fprintf(&buf, "- 期间总固定成本：%.2f\n", total.FixedCost)
        fmt.Fprintf(&buf, "- 期间总利润：%.2f\n", total.Profit)

        // 三、按国家汇总
        if len(countryAgg) > 0 {
            buf.WriteString("\n**三、按国家汇总（人民币）**\n")
            buf.WriteString("\n| 国家 | 销售额 | 广告成本 | 货款成本 | 平台手续费 | 刷单费用 | 固定成本 | 总利润 |\n")
            buf.WriteString("| --- | --- | --- | --- | --- | --- | --- | --- |\n")

            countries := make([]string, 0, len(countryAgg))
            for k := range countryAgg {
                countries = append(countries, k)
            }
            sort.Strings(countries)

            var bestCountry, worstCountry string
            var bestProfit, worstProfit float64
            first := true

            for _, ctry := range countries {
                ca := countryAgg[ctry]
                fmt.Fprintf(&buf,
                    "| %s | %.2f | %.2f | %.2f | %.2f | %.2f | %.2f | %.2f |\n",
                    ctry,
                    ca.SaleTotal,
                    ca.AdCost,
                    ca.GoodsCost,
                    ca.PlatformFee,
                    ca.ShuaDanFee,
                    ca.FixedCost,
                    ca.Profit,
                )

                if first {
                    bestCountry, worstCountry = ctry, ctry
                    bestProfit, worstProfit = ca.Profit, ca.Profit
                    first = false
                } else {
                    if ca.Profit > bestProfit {
                        bestProfit = ca.Profit
                        bestCountry = ctry
                    }
                    if ca.Profit < worstProfit {
                        worstProfit = ca.Profit
                        worstCountry = ctry
                    }
                }
            }

            // 四、简要分析
            buf.WriteString("\n**四、简要分析**\n")

            // 统计天数和日均利润
            days := make([]string, 0, len(dayProfit))
            for d := range dayProfit {
                days = append(days, d)
            }
            sort.Strings(days)

            dayCount := float64(len(days))
            avgProfit := 0.0
            if dayCount > 0 {
                avgProfit = total.Profit / dayCount
            }

            fmt.Fprintf(&buf, "\n- 本期共计结算天数：%.0f 天，期间总利润：%.2f，日均利润约：%.2f。\n", dayCount, total.Profit, avgProfit)

            if bestCountry != "" {
                fmt.Fprintf(&buf, "- 利润最高的国家：%s（总利润：%.2f）。\n", bestCountry, bestProfit)
            }
            if worstCountry != "" && worstCountry != bestCountry {
                fmt.Fprintf(&buf, "- 利润最低的国家：%s（总利润：%.2f）。\n", worstCountry, worstProfit)
            }

            // 找出利润最高/最低的日期
            if len(days) > 0 {
                bestDay, worstDay := days[0], days[0]
                bestDayProfit, worstDayProfit := dayProfit[bestDay], dayProfit[bestDay]
                for _, d := range days[1:] {
                    p := dayProfit[d]
                    if p > bestDayProfit {
                        bestDayProfit = p
                        bestDay = d
                    }
                    if p < worstDayProfit {
                        worstDayProfit = p
                        worstDay = d
                    }
                }

                fmt.Fprintf(&buf, "- 利润最高的日期：%s（利润：%.2f）。\n", bestDay, bestDayProfit)
                if worstDay != bestDay {
                    fmt.Fprintf(&buf, "- 利润最低的日期：%s（利润：%.2f）。\n", worstDay, worstDayProfit)
                }
            }

            // 五、AI 数据分析解读（通过豆包）
            if DoubaoEndpoint != "" && DoubaoAPIKey != "" {
                type aiPayload struct {
                    ReportType string             `json:"report_type"`
                    StartDate  string             `json:"start_date"`
                    EndDate    string             `json:"end_date"`
                    Total      agg                `json:"total"`
                    Countries  map[string]*agg    `json:"countries"`
                    DayProfit  map[string]float64 `json:"day_profit"`
                }

                payload := aiPayload{
                    ReportType: reportType,
                    StartDate:  startDate,
                    EndDate:    endDate,
                    Total:      total,
                    Countries:  countryAgg,
                    DayProfit:  dayProfit,
                }

                aiText, err := callDoubaoForAnalysis(payload)
                if err != nil {
                    fmt.Printf("[ai-analysis] 调用豆包分析失败：%v\n", err)
                } else if aiText != "" {
                    buf.WriteString("\n**五、AI 数据分析解读（豆包）**\n\n")
                    buf.WriteString(aiText)
                }
            }
        }
    }

    // 发送到企业微信
    content := buf.String()

    webhook := WecomWebhook
    if webhook == "" {
        webhook = os.Getenv("WECHAT_ROBOT_WEBHOOK")
        if webhook == "" {
            if key := os.Getenv("WECHAT_ROBOT_KEY"); key != "" {
                webhook = fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=%s", key)
            }
        }
    }
    if webhook == "" {
        return fmt.Errorf("未配置企业微信机器人地址，请在 config.yaml 的 wecom.webhook 中配置，或设置 WECHAT_ROBOT_WEBHOOK / WECHAT_ROBOT_KEY 环境变量")
    }

    payload := map[string]any{
        "msgtype": "markdown",
        "markdown": map[string]any{
            "content": content,
        },
    }
    b, err := json.Marshal(payload)
    if err != nil {
        return err
    }

    resp, err := http.Post(webhook, "application/json", bytes.NewReader(b))
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("企业微信返回状态码 %d", resp.StatusCode)
    }

    return nil
}

// 调用豆包大模型进行结算数据分析
// TestDoubaoConfig 提供一个简单接口，用于验证豆包配置是否正确可用。
// 访问 /api/test-doubao 即可触发一次测试调用。
func TestDoubaoConfig() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 构造一个很简单的测试数据
        payload := map[string]any{
            "type": "config_test",
            "time": time.Now().Format("2006-01-02 15:04:05"),
        }

        text, err := callDoubaoForAnalysis(payload)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "ok":    false,
                "error": err.Error(),
            })
            return
        }

        if text == "" {
            c.JSON(http.StatusOK, gin.H{
                "ok":      false,
                "message": "调用成功，但未返回内容。请检查模型配置和权限。",
            })
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "ok":       true,
            "analysis": text,
        })
    }
}

func callDoubaoForAnalysis(data interface{}) (string, error) {
    if DoubaoEndpoint == "" || DoubaoAPIKey == "" {
        return "", nil
    }

    if DoubaoModel == "" {
        fmt.Println("[ai-analysis] DoubaoModel 未配置，跳过 AI 分析")
        return "", nil
    }

    raw, err := json.Marshal(data)
    if err != nil {
        return "", err
    }

    prompt := fmt.Sprintf(
        "你是一个精通电商数据分析的运营专家。下面是电商结算的汇总数据（JSON）：%s。请用中文生成一段不超过 500 字的周/月度经营分析报告，重点说明：整体盈利情况、主要利润来源国家、亏损或利润较低的国家、利润波动较大的日期，以及可以给运营的 3-5 条具体建议。",
        string(raw),
    )

    // 使用 Responses API：model + input，多模态结构但这里只用文本
    body := map[string]any{
        "model": DoubaoModel,
        "input": []map[string]any{
            {
                "role": "user",
                "content": []map[string]string{
                    {
                        "type": "input_text",
                        "text": prompt,
                    },
                },
            },
        },
    }

    reqBytes, err := json.Marshal(body)
    if err != nil {
        return "", err
    }

    req, err := http.NewRequest("POST", DoubaoEndpoint, bytes.NewReader(reqBytes))
    if err != nil {
        return "", err
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+DoubaoAPIKey)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("豆包接口返回状态码 %d", resp.StatusCode)
    }

    // Responses API 返回结构：从 output[0].content[*].text 中取文本
    var respData struct {
        Output []struct {
            Content []struct {
                Type string `json:"type"`
                Text string `json:"text"`
            } `json:"content"`
        } `json:"output"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
        return "", err
    }
    if len(respData.Output) == 0 {
        return "", nil
    }

    // 遍历所有 output 段，找到第一段文本内容
    for _, out := range respData.Output {
        for _, item := range out.Content {
            if item.Type == "output_text" || item.Type == "text" || item.Type == "output_text_block" {
                return item.Text, nil
            }
        }
    }
    return "", nil
}
