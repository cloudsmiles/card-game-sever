package config

import (
	"encoding/json"
	"log"
	"os"
)

// Config 应用配置
type Config struct {
	Server ServerConfig `json:"server"`
	MySQL  MySQLConfig  `json:"mysql"`
	JWT    JWTConfig    `json:"jwt"`
	WeChat WeChatConfig `json:"wechat"`
}

type ServerConfig struct {
	Port string `json:"port"`
}

type MySQLConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
}

// DSN 返回 GORM 连接字符串
func (c MySQLConfig) DSN() string {
	return c.User + ":" + c.Password + "@tcp(" + c.Host + ":" + c.Port + ")/" + c.DBName + "?charset=utf8mb4&parseTime=True&loc=Local"
}

type JWTConfig struct {
	Secret     string `json:"secret"`
	ExpireHour int    `json:"expire_hour"` // token 过期时间（小时）
}

type WeChatConfig struct {
	// 微信公众号（H5 网页授权）
	H5AppID     string `json:"h5_app_id"`
	H5AppSecret string `json:"h5_app_secret"`
	// 微信开放平台（PC 扫码登录）
	OpenAppID     string `json:"open_app_id"`
	OpenAppSecret string `json:"open_app_secret"`
}

// Global 全局配置实例
var Global *Config

// Load 从配置文件加载配置
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	// 设置默认值
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}
	if cfg.MySQL.Host == "" {
		cfg.MySQL.Host = "127.0.0.1"
	}
	if cfg.MySQL.Port == "" {
		cfg.MySQL.Port = "3306"
	}
	if cfg.JWT.ExpireHour == 0 {
		cfg.JWT.ExpireHour = 168 // 默认7天
	}
	Global = &cfg
	log.Printf("配置加载完成 [MySQL: %s:%s/%s]", cfg.MySQL.Host, cfg.MySQL.Port, cfg.MySQL.DBName)
	return &cfg, nil
}
