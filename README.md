# ORZ - 现代化 Go 微框架

ORZ 是一个简洁、模块化的 Go Web 框架，专注于提供开发者友好的 API 和清晰的架构。

## ✨ 特性

- 🚀 **模块化架构** - 可插拔的模块系统
- 🔧 **类型安全的依赖注入** - 避免强制类型转换
- 🛡️ **类型安全** - 充分利用 Go 泛型
- 📊 **数据库集成** - 支持 MySQL、PostgreSQL、SQLite
- 🌐 **HTTP 服务** - 基于 Echo 的高性能 HTTP 服务器
- 📝 **结构化日志** - 基于 Zap 的高性能日志系统
- ⚙️ **灵活配置** - 支持文件、字节、Map 多种配置源
- ⚡ **高性能** - 轻量级设计，性能优异

## 🚀 快速开始

### 安装

```bash
go get github.com/go-orz/orz
```

### 简单示例

```go
package main

import (
    "log"
    
    "github.com/go-orz/orz"
    "github.com/labstack/echo/v4"
    "gorm.io/gorm"
)

func main() {
    err := orz.Quick("config.yaml", func(e *echo.Echo, db *gorm.DB) error {
        // 设置路由
        e.GET("/", func(c echo.Context) error {
            return c.JSON(200, map[string]string{
                "message": "Hello from ORZ!",
            })
        })
        
        return nil
    })
    
    if err != nil {
        log.Fatal("Failed to start:", err)
    }
}
```

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