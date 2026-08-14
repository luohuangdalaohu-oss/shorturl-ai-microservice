package model

import "time"

// URLMapping 短链在 MySQL 中对应的表结构体
type URLMapping struct {
	ID          uint64     `gorm:"primaryKey;column:id"`
	ShortCode   string     `gorm:"uniqueIndex:uk_short_code;column:short_code"`
	OriginalURL string     `gorm:"column:original_url"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	ExpiredAt   *time.Time `gorm:"column:expired_at"`
}

// TableName 指定 MySQL 表名
func (URLMapping) TableName() string {
	return "url_mapping"
}
