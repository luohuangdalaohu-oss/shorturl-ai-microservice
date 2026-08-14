package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// 全局配置结构体
type Config struct {
	Server    ServerConfig    `mapstructure:"service"`
	Snowflake SnowflakeConfig `mapstructure:"snowflake"`
	MySQL     MySQLConfig     `mapstructure:"mysql"`
	Redis     RedisConfig     `mapstructure:"redis"`
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type SnowflakeConfig struct {
	StartTime string `mapstructure:"start_time"`
	MachineID int64  `mapstructure:"machine_id"`
}

type MySQLConfig struct {
	DSN string `mapstructure:"dsn"`
}

type RedisConfig struct {
	Addr            string `mapstructure:"addr"`
	Password        string `mapstructure:"password"`
	DB              int    `mapstructure:"db"`
	ExpirationHours int    `mapstructure:"expiration_hours"`
}

var Conf *Config

// LoadConfig 从指定 yaml 路径加载配置
func LoadConfig(filePath string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(filePath)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	Conf = &c
	return Conf, nil
}
