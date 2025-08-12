package orz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"

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
	Log      Log       // 日志
	Database Database  // 数据库配置
	Server   Server    // Web 服务器配置
	App      AppConfig // 应用程序个性化配置
}

type Server struct {
	Addr        string
	TLS         TLS
	IPExtractor string
	IPTrustList []string // 信任的IP
}

type TLS struct {
	Enabled bool
	Auto    bool
	Cert    string
	Key     string
}

type Log struct {
	Level    string // debug, info, warn, error
	Filename string // 日志文件路径
	Encode   string // console, json
	Console  bool   // 是否输出到控制台
	MaxSize  int    // 日志文件最大大小(MB)
	MaxAge   int    // 日志保留天数
	Compress bool   // 是否压缩日志
}

type Database struct {
	Enabled  bool
	Type     DatabaseType
	Mysql    MysqlCfg
	Sqlite   SqliteConfig
	Postgres PostgresCfg
	ShowSql  bool
}

type MysqlCfg struct {
	DSN             string        `yaml:"dsn"`
	Hostname        string        `yaml:"hostname"`
	Port            int           `yaml:"port"`
	Username        string        `yaml:"username"`
	Password        string        `yaml:"password"`
	Database        string        `yaml:"database"`
	MaxIdleConns    int           `yaml:"max-idle-conns"`
	MaxOpenConns    int           `yaml:"max-open-conns"`
	ConnMaxLifetime time.Duration `yaml:"conn-max-lifetime"`
}

type PostgresCfg struct {
	DSN      string `yaml:"dsn"`
	Hostname string `yaml:"hostname"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type SqliteConfig struct {
	Path string
}

// ConfigManager 配置管理器
type ConfigManager struct {
	viper *viper.Viper
}

// NewConfigManager 创建新的配置管理器
func NewConfigManager() *ConfigManager {
	v := viper.New()
	v.SetConfigType("yaml")

	// 设置默认值
	v.SetDefault("log.level", "info")
	v.SetDefault("log.filename", "")
	v.SetDefault("log.encode", "console")
	v.SetDefault("log.console", true)
	v.SetDefault("log.maxSize", 100)
	v.SetDefault("log.maxAge", 7)
	v.SetDefault("log.compress", true)
	v.SetDefault("database.enabled", true)
	v.SetDefault("database.type", "sqlite")
	v.SetDefault("database.sqlite.path", "data/app.db")
	v.SetDefault("database.showSql", false)
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
