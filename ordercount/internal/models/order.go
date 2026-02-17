package models

import "time"

// 仅允许三个国家
// 增加币种字段
type Order struct {
    ID          uint      `gorm:"primaryKey" json:"id"`
    // 记录这条订单由哪个登录用户录入，便于按用户区分可见范围
    UserID      uint      `json:"user_id" gorm:"index"`
    Country     string    `json:"country" gorm:"size:20;check:country IN ('菲律宾','印尼','马来西亚')"`
    Platform    string    `json:"platform" gorm:"size:100"`
    OrderNo     string    `json:"order_no" gorm:"size:200;index"`
    ProductName string    `json:"product_name" gorm:"size:255"`
    SKU         string    `json:"sku" gorm:"size:200"`
    Quantity    int       `json:"quantity"`
    TotalAmount float64   `json:"total_amount"`
    Currency    string    `json:"currency" gorm:"size:10"`
    CreatedAt   time.Time `json:"created_at"`
}
