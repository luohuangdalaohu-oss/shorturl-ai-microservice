package dao

import (
	"context"
	"errors"
	"time"

	"shorturl/internal/transform/config"
	"shorturl/internal/transform/model"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// DAO 数据访问对象（专门操作 MySQL 和 Redis）
type DAO struct {
	db  *gorm.DB
	rdb *redis.Client
	ttl time.Duration
}

var GlobalDAO *DAO

// InitDAO 初始化 MySQL 和 Redis 连接
func InitDAO(cfg *config.Config) (*DAO, error) {
	// 1. 连接 MySQL
	db, err := gorm.Open(mysql.Open(cfg.MySQL.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// 2. 连接 Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// 测试 Redis 连通性
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	GlobalDAO = &DAO{
		db:  db,
		rdb: rdb,
		ttl: time.Duration(cfg.Redis.ExpirationHours) * time.Hour,
	}

	return GlobalDAO, nil
}

// SaveURL 保存短链（双写：MySQL 插入 + Redis 缓存）
func (d *DAO) SaveURL(ctx context.Context, mapping *model.URLMapping) error {
	// ① 写入 MySQL 持久化
	if err := d.db.WithContext(ctx).Create(mapping).Error; err != nil {
		return err
	}

	// ② 写入 Redis 缓存（Key: "shorturl:" + short_code）
	cacheKey := "shorturl:" + mapping.ShortCode
	_ = d.rdb.Set(ctx, cacheKey, mapping.OriginalURL, d.ttl).Err()

	return nil
}

// GetOriginalURL 获取长链接（缓存优先 Cache-Aside 模式）
func (d *DAO) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	cacheKey := "shorturl:" + shortCode

	// ① 先查 Redis 缓存（极速响应）
	val, err := d.rdb.Get(ctx, cacheKey).Result()
	if err == nil && val != "" {
		return val, nil // 命中缓存，直接返回！
	}

	// ② 缓存没命中，查 MySQL 数据库
	var mapping model.URLMapping
	err = d.db.WithContext(ctx).Where("short_code = ?", shortCode).First(&mapping).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("短链接不存在")
		}
		return "", err
	}

	// ③ 查到后回填 Redis 缓存
	_ = d.rdb.Set(ctx, cacheKey, mapping.OriginalURL, d.ttl).Err()

	return mapping.OriginalURL, nil
}
