package models

import "time"

// NavLink 用户自定义导航链接（导航助手）
// 仅存储站点名称、URL 和账户名等非敏感信息，不建议存储密码。
type NavLink struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"`
	Category  string    `gorm:"size:100;index" json:"category"` // 分类，便于按类别筛选和快速导航
	Title     string    `gorm:"size:200" json:"title"`   // 站点名称
	URL       string    `gorm:"size:500" json:"url"`     // 目标网址
	Account   string    `gorm:"size:200" json:"account"` // 登录账户名或备注
	Remark    string    `gorm:"size:500" json:"remark"`  // 备注
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
