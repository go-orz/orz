package config

import (
	_ "embed"
	"encoding/json"
	"github.com/spf13/viper"
	"strings"
	"time"
)

var (
	instance *Config
)

type AppConfig map[string]any

func (r AppConfig) MustUnmarshal(v any) {
	data, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, v)
	if err != nil {
		panic(err)
	}
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
	Level    string
	Filename string
}

type Database struct {
	Enabled    bool
	Type       string
	Mysql      MysqlCfg
	ClickHouse ClickHouseConfig
	Sqlite     SqliteConfig
	Postgresql PostgresqlCfg
	ShowSql    bool
}

type MysqlCfg struct {
	Hostname        string        `yaml:"hostname"`
	Port            int           `yaml:"port"`
	Username        string        `yaml:"username"`
	Password        string        `yaml:"password"`
	Database        string        `yaml:"database"`
	MaxIdleConns    int           `yaml:"max-idle-conns"`
	MaxOpenConns    int           `yaml:"max-open-conns"`
	ConnMaxLifetime time.Duration `yaml:"conn-max-lifetime"`
}

type PostgresqlCfg struct {
	Hostname string `yaml:"hostname"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type ClickHouseConfig struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

type SqliteConfig struct {
	Path string
}

type ProxyCfg struct {
	Enabled bool
	Proxy   string
}

func MustInit(config string) {
	if config == "" {
		config = "config.yaml"
	}
	viper.SetConfigFile(config)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if err := viper.ReadInConfig(); err != nil {
		panic(`read config err: ` + err.Error())
	} else {
		instance = &Config{}
		if err := viper.Unmarshal(instance); err != nil {
			panic(`unmarshal config err: ` + err.Error())
		}
	}
}

func Conf() *Config {
	if instance == nil {
		panic(`you must call config.MustInit(config string) first`)
	}
	return instance
}
