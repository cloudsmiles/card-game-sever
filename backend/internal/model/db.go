package model

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 全局数据库实例
var DB *gorm.DB

// InitDB 初始化数据库连接并自动迁移
func InitDB(dsn string) error {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return err
	}

	// 自动迁移
	if err := db.AutoMigrate(&User{}, &UserScore{}); err != nil {
		return err
	}

	DB = db
	log.Println("数据库连接成功，表迁移完成")
	return nil
}
