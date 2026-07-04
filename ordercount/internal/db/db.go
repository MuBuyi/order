package db

import (
	"log"
	"strings"

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

	// 自动迁移: 包括 User 表，便于首次部署时自动创建缺失表结构
	modelsToMigrate := []interface{}{
		&models.User{},
		&models.Order{},
		&models.DailySettlement{},
		&models.Product{},
		&models.Store{},
		&models.StoreUser{},
		&models.StoreDailyStat{},
		&models.NavLink{},
		&models.ReturnRecord{},
	}

	for _, m := range modelsToMigrate {
		if err := db.AutoMigrate(m); err != nil {
			// MySQL 有时会因为索引/约束名不一致尝试执行 DROP，返回 1091 错误（Can't DROP ...）
			// 为了兼容已有数据库，遇到此类错误时记录警告并继续，而非直接失败。
			if strings.Contains(err.Error(), "Can't DROP") || strings.Contains(err.Error(), "Error 1091") {
				log.Println("warning: AutoMigrate non-fatal drop error ignored:", err)
				continue
			}
			return nil, err
		}
	}

	// 兜底：某些旧库在 AutoMigrate 触发 1091 时会导致后续字段迁移未落地，这里显式补齐 users.is_active。
	if !db.Migrator().HasColumn(&models.User{}, "is_active") {
		if err := db.Migrator().AddColumn(&models.User{}, "IsActive"); err != nil {
			return nil, err
		}
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
			IsActive:     true,
			Permissions:  "stats,settlement,returns,product,shop,shop-manage,exchange,charts,nav-helper,users",
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
			IsActive:     true,
			Permissions:  "stats,settlement,returns,product,shop,shop-manage,exchange,charts,nav-helper",
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
