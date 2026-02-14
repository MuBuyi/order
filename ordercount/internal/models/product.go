package models

import "time"

// Product 商品信息
// 目前只包含图片和 SKU/名称，可后续扩展
type Product struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    SKU       string    `json:"sku" gorm:"size:100;index"`
    Name      string    `json:"name" gorm:"size:255"`
    ImageURL  string    `json:"image_url" gorm:"size:500"`
    // 基础成本（可理解为超级管理员视角的默认成本）
    Cost      float64   `json:"cost" gorm:"type:decimal(10,2);default:0"`
    // 不同角色专属成本：管理员成本、员工成本
    CostAdmin float64   `json:"cost_admin" gorm:"type:decimal(10,2);default:0"`
    CostStaff float64   `json:"cost_staff" gorm:"type:decimal(10,2);default:0"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
