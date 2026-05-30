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

```yaml
log:
  level: "info"
  filename: "logs/app.log"

database:
  enabled: true
  type: "sqlite"
  sqlite:
    path: "data/app.db"

server:
  addr: ":8080"
```
