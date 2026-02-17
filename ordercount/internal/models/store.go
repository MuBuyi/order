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

    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
