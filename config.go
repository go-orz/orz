package orz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type AppConfig map[string]any
type DatabaseType string

const (
	DatabaseSqlite     DatabaseType = "sqlite"
	DatabaseMysql      DatabaseType = "mysql"
	DatabasePostgres   DatabaseType = "postgres"
	DatabasePostgresql DatabaseType = "postgresql"
)

// Unmarshal 解析配置到指定结构
func (r AppConfig) Unmarshal(v any) error {
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

type Config struct {
	Log      LogConfig      `yaml:"log"`      // 日志配置
	Database DatabaseConfig `yaml:"database"` // 数据库配置
	Server   ServerConfig   `yaml:"server"`   // Web 服务器配置
	App      AppConfig      `yaml:"app"`      // 应用程序个性化配置
}

type ServerConfig struct {
	Addr        string    `yaml:"addr"`
	TLS         TLSConfig `yaml:"tls"`
	IPExtractor string    `yaml:"ip_extractor"`
	IPTrustList []string  `yaml:"ip_trust_list"` // 信任的IP
}

type TLSConfig struct {
	Enabled bool   `yaml:"enabled"`
	Auto    bool   `yaml:"auto"`
	Cert    string `yaml:"cert"`
	Key     string `yaml:"key"`
}

type LogConfig struct {
	Level    string `yaml:"level"`    // debug, info, warn, error
	Filename string `yaml:"filename"` // 日志文件路径
	Encode   string `yaml:"encode"`   // console, json
	Console  bool   `yaml:"console"`  // 是否输出到控制台
	MaxSize  int    `yaml:"max_size"` // 日志文件最大大小(MB)
	MaxAge   int    `yaml:"max_age"`  // 日志保留天数
	Compress bool   `yaml:"compress"` // 是否压缩日志
}

type DatabaseConfig struct {
	Enabled  bool         `yaml:"enabled"`
	Type     DatabaseType `yaml:"type"`
	URL      string       `yaml:"url"`
	Mysql    MysqlCfg     `yaml:"mysql"`
	Sqlite   SqliteConfig `yaml:"sqlite"`
	Postgres PostgresCfg  `yaml:"postgres"`
	ShowSql  bool         `yaml:"show_sql"`
}

type MysqlCfg struct {
	Hostname string `yaml:"hostname"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type PostgresCfg struct {
	Hostname string `yaml:"hostname"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type SqliteConfig struct {
	Path string `yaml:"path"`
}

// ConfigManager 配置管理器
type ConfigManager struct {
	viper *viper.Viper
}

// NewConfigManager 创建新的配置管理器
func NewConfigManager() *ConfigManager {
	v := viper.New()
	v.SetConfigType("yaml")

	// 启用环境变量支持
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 设置默认值
	v.SetDefault("log.level", "info")
	v.SetDefault("log.filename", "")
	v.SetDefault("log.encode", "console")
	v.SetDefault("log.console", true)
	v.SetDefault("log.max_size", 100)
	v.SetDefault("log.max_age", 7)
	v.SetDefault("log.compress", true)
	v.SetDefault("database.enabled", true)
	v.SetDefault("database.type", "sqlite")
	v.SetDefault("database.show_sql", false)
	v.SetDefault("server.addr", ":8080")

	return &ConfigManager{viper: v}
}

// LoadFromFile 从文件加载配置
func (cm *ConfigManager) LoadFromFile(configPath string) error {
	if configPath == "" {
		return fmt.Errorf("config path is empty")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file does not exist: %s", configPath)
	}

	cm.viper.SetConfigFile(configPath)
	return cm.viper.ReadInConfig()
}

// LoadFromBytes 从字节数组加载配置
func (cm *ConfigManager) LoadFromBytes(data []byte) error {
	return cm.viper.ReadConfig(bytes.NewReader(data))
}

// LoadFromMap 从Map加载配置
func (cm *ConfigManager) LoadFromMap(data map[string]interface{}) error {
	return cm.viper.MergeConfigMap(data)
}

// GetConfig 获取配置
func (cm *ConfigManager) GetConfig() *Config {
	config := &Config{}
	if err := cm.viper.Unmarshal(config); err != nil {
		return nil
	}
	return config
}

// Get 获取配置值
func (cm *ConfigManager) Get(key string) interface{} {
	return cm.viper.Get(key)
}

// GetString 获取字符串配置
func (cm *ConfigManager) GetString(key string) string {
	return cm.viper.GetString(key)
}

// GetInt 获取整数配置
func (cm *ConfigManager) GetInt(key string) int {
	return cm.viper.GetInt(key)
}

// GetBool 获取布尔配置
func (cm *ConfigManager) GetBool(key string) bool {
	return cm.viper.GetBool(key)
}
