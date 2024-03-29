package config

import (
	_ "embed"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"time"
)

var (
	instance *Config
)

type CustomConfig map[string]any

func (r CustomConfig) MustUnmarshal(v any) {
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
	Log      Log      // 日志
	Database Database // 数据库配置
	Server   Server   // Web 服务器配置
	Custom   CustomConfig
}

type Server struct {
	Port        int
	Host        string
	TLS         TLS
	IPExtractor string
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

func IPExtractor() echo.IPExtractor {
	switch Conf().Server.IPExtractor {
	case "direct":
		return echo.ExtractIPDirect()
	case "x-real-ip":
		return echo.ExtractIPFromRealIPHeader()
	case "x-forwarded-for":
		return echo.ExtractIPFromXFFHeader()
	default:
		return echo.ExtractIPDirect()
	}
}
