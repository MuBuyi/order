package models

import "time"

// ReturnRecord 表示一次退货记录，由登录用户创建并按用户隔离
type ReturnRecord struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"user_id" gorm:"index"`
	OrderID      string    `json:"order_id" gorm:"size:100;index"`
	Country      string    `json:"country" gorm:"size:100"`
	StoreName    string    `json:"store_name" gorm:"size:255"`
	ProductName  string    `json:"product_name" gorm:"size:255"`
	SKU          string    `json:"sku" gorm:"size:200"`
	Quantity     int       `json:"quantity"`
	RefundAmount float64   `json:"refund_amount"`
	LossAmount   float64   `json:"loss_amount"`
	ReturnDate   string    `json:"return_date" gorm:"size:20"` // 业务日期，YYYY-MM-DD
	Handler      string    `json:"handler" gorm:"size:100"`
	Remark       string    `json:"remark" gorm:"size:500"`
	CreatedAt    time.Time `json:"created_at"`
}
