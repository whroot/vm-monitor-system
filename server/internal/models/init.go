package models

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"vm-monitoring-system/internal/config"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB    *gorm.DB
	Cache *redis.Client
)

// InitDB 初始化数据库连接
func InitDB(cfg config.DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取底层数据库失败: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	DB = db
	return db, nil
}

// InitCache 初始化Redis缓存
func InitCache(cfg config.RedisConfig) error {
	Cache = redis.NewClient(&redis.Options{
		Addr:         cfg.Addr(),
		Password:     cfg.Password,
		DB:           cfg.Database,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Cache.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("连接Redis失败: %w", err)
	}

	return nil
}

// AutoMigrate 自动迁移数据库表
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&UserRole{},
		&Role{},
		&Permission{},
		&RolePermission{},
		&UserSession{},
		&VM{},
		&VMGroup{},
		&VMGroupMember{},
		&AlertRule{},
		&AlertCondition{},
		&AlertRecord{},
		&AuditLog{},
	)
}

// JSONMap 用于存储JSON数据的类型
type JSONMap map[string]interface{}

// Value 实现driver.Valuer接口
func (j JSONMap) Value() (interface{}, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现sql.Scanner接口
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("无法扫描类型 %T 到 JSONMap", value)
	}

	return json.Unmarshal(bytes, j)
}

// StringArray 字符串数组类型
type StringArray []string

// Value 实现driver.Valuer接口
func (a StringArray) Value() (interface{}, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

// Scan 实现sql.Scanner接口
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("无法扫描类型 %T 到 StringArray", value)
	}

	return json.Unmarshal(bytes, a)
}

// TimePtr 返回当前时间的指针
func TimePtr(t time.Time) *time.Time {
	return &t
}
