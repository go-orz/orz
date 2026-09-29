# ORZ - 现代化 Go 微框架

ORZ 是一个简洁、模块化的 Go Web 框架，专注于提供开发者友好的 API 和清晰的架构。

## ✨ 特性

- 🚀 **模块化架构** - 可插拔的模块系统
- 🔧 **类型安全的依赖注入** - 避免强制类型转换
- 🛡️ **类型安全** - 充分利用 Go 泛型
- 📊 **数据库集成** - 支持 MySQL、PostgreSQL、SQLite，驱动按需引入
- 🌐 **HTTP 服务** - 基于 Echo 的高性能 HTTP 服务器
- 📝 **结构化日志** - 基于 Zap 的高性能日志系统
- ⚙️ **灵活配置** - 支持文件、字节、Map 多种配置源
- ⚡ **高性能** - 轻量级设计，性能优异

## 🚀 快速开始

### 安装

```bash
go get github.com/go-orz/orz
```

数据库驱动已拆成独立子模块，按需 blank import 对应驱动即可，不会随着核心框架默认引入全部数据库依赖。

### 简单示例

```go
package main

import (
    "log"

    "github.com/go-orz/orz"
    _ "github.com/go-orz/orz/drivers/sqlite"
    "github.com/labstack/echo/v5"
)

type User struct {
    ID   uint   `gorm:"primaryKey"`
    Name string `json:"name"`
}

func main() {
    err := orz.Quick("config.yaml", func(app *orz.App) error {
        db := app.GetDatabase()
        e := app.GetEcho()

        if err := db.AutoMigrate(&User{}); err != nil {
            return err
        }

        // 设置路由
        e.GET("/", func(c *echo.Context) error {
            return orz.Message(c, 200, "Hello from ORZ!")
        })
        
        return nil
    })
    
    if err != nil {
        log.Fatal("Failed to start:", err)
    }
}
```

按需引入数据库驱动，例如 SQLite:

```go
import _ "github.com/go-orz/orz/drivers/sqlite"
```

如果你使用 `go mod tidy`，Go 会自动补齐对应驱动子模块依赖。

HTTP helper 默认行为：

- `orz.Ok(c, data)` 直接返回原始 JSON 数据
- `orz.Created(c, data)` 直接返回原始 JSON 数据
- `orz.Message(c, status, message)` 返回最小的 `{ "message": "..." }`
- `orz.ErrorResponse(c, code, message)` 返回最小的 `{ "message": "..." }`

### 配置文件 (config.yaml)

数据库连接池由框架统一配置，适用于所有注册驱动和数据库连接入口。省略 `database.pool` 时默认最多打开 30 个连接、保留 10 个空闲连接，连接最长存活 3600 秒，最大空闲时间 300 秒。直接调用 `ConnectDatabase` 等函数也使用相同默认值。

各字段使用非指针整数类型，可单独用正值覆盖；未配置或设为 `0` 时使用对应默认值，旧部署无需修改配置文件。负值或超出可表示范围的时间会返回错误；最大空闲连接数超过最大打开连接数时自动收紧。限制按连接池生效，多实例部署需预留 PostgreSQL 管理连接和其他服务连接的容量。

```yaml
log:
  level: "info"
  filename: "logs/app.log"

database:
  enabled: true
  type: "sqlite"
  pool:
    max_open_conns: 30
    max_idle_conns: 10
    conn_max_lifetime_seconds: 3600
    conn_max_idle_time_seconds: 300
  sqlite:
    path: "data/app.db"

server:
  addr: ":8080"
  ip_extractor: "direct"           # direct, x-forwarded-for, x-real-ip，或自定义 Header 名称
  ip_trust_list: []                # 额外可信代理 IP/CIDR 列表；回环、链路本地、私网代理默认可信
```
