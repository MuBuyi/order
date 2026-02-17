package db

import (
    "log"

    "golang.org/x/crypto/bcrypt"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"

    "ordercount/internal/models"
)

func InitDB(dsn string) (*gorm.DB, error) {
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, err
    }

    // 自动迁移（User 表已手工迁移，避免索引兼容性问题，这里不再自动迁移 User）
    if err := db.AutoMigrate(&models.Order{}, &models.DailySettlement{}, &models.Product{}, &models.Store{}, &models.StoreUser{}); err != nil {
        return nil, err
    }

    // 如果还没有用户，创建默认超级管理员和管理员账号，便于首次登录
    var count int64
    if err := db.Model(&models.User{}).Count(&count).Error; err != nil {
        return nil, err
    }
    if count == 0 {
        // 超级管理员，默认拥有所有页面权限
        rootPwd := "root123"
        rootHash, err := bcrypt.GenerateFromPassword([]byte(rootPwd), bcrypt.DefaultCost)
        if err != nil {
            return nil, err
        }
        root := models.User{
            Username:     "root",
            PasswordHash: string(rootHash),
            Role:         "superadmin",
            Permissions:  "settlement,product,shop",
        }
        // 管理员，默认拥有结账工具和商品管理权限
        adminPwd := "admin123"
        adminHash, err := bcrypt.GenerateFromPassword([]byte(adminPwd), bcrypt.DefaultCost)
        if err != nil {
            return nil, err
        }
        admin := models.User{
            Username:     "admin",
            PasswordHash: string(adminHash),
            Role:         "admin",
            Permissions:  "settlement,product,shop",
        }
        if err := db.Create(&root).Error; err != nil {
            return nil, err
        }
        if err := db.Create(&admin).Error; err != nil {
            return nil, err
        }
        log.Println("created default superadmin user: username=root, password=root123")
        log.Println("created default admin user: username=admin, password=admin123")
    }

    return db, nil
}
