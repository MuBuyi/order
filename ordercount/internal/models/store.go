package models

import "time"

// Store 记录店铺基础信息
// 注意：其中的登录密码等字段仅用于内部管理，请勿对外暴露。
type Store struct {
    ID uint `gorm:"primaryKey" json:"id"`

    Country       string `json:"country" gorm:"size:50"`   // 店铺所在国家
    Platform      string `json:"platform" gorm:"size:50"`  // 店铺所属平台
    Name          string `json:"name" gorm:"size:100"`     // 店铺名
    LoginAccount  string `json:"login_account" gorm:"size:100"` // 店铺登录账号
    LoginPassword string `json:"login_password" gorm:"size:255"` // 店铺登录密码（明文存储，仅内部使用）
    Phone         string `json:"phone" gorm:"size:30"`     // 绑定手机号
    Email         string `json:"email" gorm:"size:100"`    // 绑定邮箱

    IsBlocked     bool   `json:"is_blocked" gorm:"default:false"` // 是否封禁，默认未封禁

    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// StoreDailyStat 记录店铺按天的运营数据（目前仅使用广告费用，销售额字段预留暂未启用）
type StoreDailyStat struct {
    ID uint `gorm:"primaryKey" json:"id"`

    StoreID uint   `json:"store_id" gorm:"index"`  // 关联店铺 ID
    Date    string `json:"date" gorm:"size:10;index"` // 统计日期：YYYY-MM-DD

    AdCost   float64 `json:"ad_cost"`    // 当天广告费用（本国货币，例如印尼盾、菲律宾比索等）
    SaleTotal float64 `json:"sale_total"` // 当天销售额（预留字段，前端暂不展示）

    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
