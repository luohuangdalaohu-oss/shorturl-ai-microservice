package dao

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"time"

	aiV1 "shorturl/api/ai/v1"
	"shorturl/internal/ai/config"

	"github.com/redis/go-redis/v9"
)

type DAO struct {
	rdb *redis.Client
	ttl time.Duration
}

func InitDAO(cfg *config.Config) (*DAO, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &DAO{
		rdb: rdb,
		ttl: time.Duration(cfg.Redis.ExpirationHours) * time.Hour,
	}, nil
}

// 计算 URL 的 MD5 作为 Redis Key
func (d *DAO) getCacheKey(url string) string {
	h := md5.Sum([]byte(url))
	return "ai:safety:" + hex.EncodeToString(h[:])
}

// GetSafetyCache 获取缓存的安全检测结果
func (d *DAO) GetSafetyCache(ctx context.Context, url string) (*aiV1.CheckURLSafetyResponse, error) {
	val, err := d.rdb.Get(ctx, d.getCacheKey(url)).Result()
	if err != nil {
		return nil, err
	}

	var resp aiV1.CheckURLSafetyResponse
	if err := json.Unmarshal([]byte(val), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SetSafetyCache 写入安全检测缓存
func (d *DAO) SetSafetyCache(ctx context.Context, url string, resp *aiV1.CheckURLSafetyResponse) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return d.rdb.Set(ctx, d.getCacheKey(url), data, d.ttl).Err()
}
