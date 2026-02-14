package models

import "time"

// User 系统登录用户，带角色信息
// 角色示例：admin（管理员，可查看/编辑成本）、staff（普通员工）
type User struct {
    ID           uint      `gorm:"primaryKey" json:"id"`
    Username     string    `gorm:"size:100;uniqueIndex" json:"username"`
    PasswordHash string    `gorm:"size:255" json:"-"`
    Role         string    `gorm:"size:50" json:"role"`
    // 逗号分隔的页面权限，例如: "settlement,product"
    Permissions  string    `gorm:"size:255" json:"permissions"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
