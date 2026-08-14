package service

import (
	"context"
	"errors"
	"time"

	"shorturl/internal/pkg/base62"
	"shorturl/internal/pkg/snowflake"
	"shorturl/internal/transform/dao"
	"shorturl/internal/transform/model"
)

type ShortenerService struct {
	dao *dao.DAO
}

func NewShortenerService(d *dao.DAO) *ShortenerService {
	return &ShortenerService{
		dao: d,
	}
}

// Shorten 长链转短链核心业务
func (s *ShortenerService) Shorten(ctx context.Context, originalURL string) (string, error) {
	if originalURL == "" {
		return "", errors.New("url 不能为空")
	}

	// 1. 调雪花算法生成全局唯一 64 位整数 ID
	id := snowflake.GenID()

	// 2. 调 Base62 将数字压缩为 6 位短码（如 3xK9a）
	shortCode := base62.Encode(uint64(id))

	// 3. 构建数据模型
	mapping := &model.URLMapping{
		ID:          uint64(id),
		ShortCode:   shortCode,
		OriginalURL: originalURL,
		CreatedAt:   time.Now(),
	}

	// 4. 调用 DAO 双写写入 MySQL 和 Redis
	if err := s.dao.SaveURL(ctx, mapping); err != nil {
		return "", err
	}

	return shortCode, nil
}

// Expand 短码还原长链接核心业务
func (s *ShortenerService) Expand(ctx context.Context, shortCode string) (string, error) {
	if shortCode == "" {
		return "", errors.New("短码不能为空")
	}

	// 调 DAO 层按 Cache-Aside 策略查询（先查 Redis，未命中查 MySQL）
	return s.dao.GetOriginalURL(ctx, shortCode)
}
