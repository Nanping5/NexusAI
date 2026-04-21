package mmysql

import (
	"NexusAi/config"
	"context"
	"fmt"
	"time"

	"NexusAi/model"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

// ErrDbNotInitialized 数据库未初始化错误
var ErrDbNotInitialized = fmt.Errorf("database not initialized")

func InitMySQL() error {

	host := config.GetConfig().MysqlConfig.Host
	port := config.GetConfig().MysqlConfig.Port
	user := config.GetConfig().MysqlConfig.User
	password := config.GetConfig().MysqlConfig.Password
	dbName := config.GetConfig().MysqlConfig.DbName

	// 先连接到 mysql（不指定数据库），用于检查/创建数据库
	dsnWithoutDB := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=%s&parseTime=true&loc=Local", user, password, host, port, "utf8mb4")

	var log logger.Interface

	if gin.Mode() == gin.DebugMode {
		log = logger.Default.LogMode(logger.Info)
	} else {
		log = logger.Default
	}

	// 连接 MySQL 服务器（不指定数据库）
	gormDb, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsnWithoutDB,
		DefaultStringSize:         256,
		DisableDatetimePrecision:  true,
		DontSupportRenameIndex:    true,
		DontSupportRenameColumn:   true,
		SkipInitializeWithVersion: false,
	}), &gorm.Config{
		Logger: log,
	})
	if err != nil {
		return err
	}

	// 检查数据库是否存在，不存在则创建
	var exists string
	err = gormDb.Raw("SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = ?", dbName).Scan(&exists).Error
	if err != nil {
		return err
	}

	if exists == "" {
		// 数据库不存在，创建数据库
		if err := gormDb.Exec(fmt.Sprintf("CREATE DATABASE `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", dbName)).Error; err != nil {
			return fmt.Errorf("创建数据库失败: %w", err)
		}
	}

	// 关闭当前连接
	sqlDB, err := gormDb.DB()
	if err != nil {
		return err
	}
	sqlDB.Close()

	// 重新连接到指定数据库
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=Local", user, password, host, port, dbName, "utf8mb4")

	gormDb, err = gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,
		DefaultStringSize:         256,
		DisableDatetimePrecision:  true,
		DontSupportRenameIndex:    true,
		DontSupportRenameColumn:   true,
		SkipInitializeWithVersion: false,
	}), &gorm.Config{
		Logger: log,
	})
	if err != nil {
		return err
	}
	sqlDb, err := gormDb.DB()
	if err != nil {
		return err
	}
	sqlDb.SetMaxIdleConns(10)
	sqlDb.SetMaxOpenConns(100)
	sqlDb.SetConnMaxLifetime(time.Hour)
	db = gormDb
	return migration()
}

func migration() error {
	return db.AutoMigrate(
		new(model.User),
		new(model.Session),
		new(model.Message),
		new(model.Admin),
		new(model.AIModelConfig),
	)
}

func NewDbClient(ctx context.Context) (*gorm.DB, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if db == nil {
		return nil, ErrDbNotInitialized
	}
	return db.WithContext(ctx), nil
}

// CloseMySQL 关闭 MySQL 连接
func CloseMySQL() error {
	if db == nil {
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GetDB 获取原始数据库连接（用于事务）
func GetDB() *gorm.DB {
	return db
}
