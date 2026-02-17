package models

import "time"

// StoreUser 关联店铺与可见的用户
// 一个店铺可以授权给多个用户，一个用户也可以看到多个店铺。
type StoreUser struct {
    ID      uint `gorm:"primaryKey" json:"id"`
    StoreID uint `json:"store_id" gorm:"index:idx_store_user,unique"`
    UserID  uint `json:"user_id" gorm:"index:idx_store_user,unique"`

    CreatedAt time.Time `json:"created_at"`
}
