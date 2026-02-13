package db

import (
    "gorm.io/driver/mysql"
    "gorm.io/gorm"

    "ordercount/internal/models"
)

func InitDB(dsn string) (*gorm.DB, error) {
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, err
    }

    // 自动迁移
    if err := db.AutoMigrate(&models.Order{}); err != nil {
        return nil, err
    }

    return db, nil
}
