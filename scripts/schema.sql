-- 创建短链数据库
CREATE DATABASE IF NOT EXISTS `shorturl` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE `shorturl`;

-- 创建短链映射表
CREATE TABLE IF NOT EXISTS `url_mapping` (
                                             `id` BIGINT UNSIGNED NOT NULL COMMENT '雪花算法生成的全局唯一主键ID',
                                             `short_code` VARCHAR(16) NOT NULL COMMENT 'Base62 压缩生成的短链码(如 3xK9a)',
    `original_url` VARCHAR(2048) NOT NULL COMMENT '原始长链接',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `expired_at` DATETIME DEFAULT NULL COMMENT '过期时间(为NULL表示永久有效)',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_short_code` (`short_code`),
    KEY `idx_created_at` (`created_at`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='短链接映射表';