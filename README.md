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

## 🏗️ 架构概览

```
orz/
├── app/              # 应用容器和生命周期管理
├── config/           # 灵活的配置管理
├── database/         # 数据库模块
├── http/             # HTTP 服务模块
├── middleware/       # 中间件系统
├── repository/       # 数据仓库层
├── utils/            # 工具函数
└── examples/         # 使用示例
```

## 📋 配置管理

### 多种配置源支持

```go
// 从文件加载
orz.NewFramework("my-app", "1.0.0").
    WithConfig("config.yaml")

// 从 Map 加载（测试友好）
orz.NewFramework("my-app", "1.0.0").
    WithConfigMap(map[string]interface{}{
        "database.enabled": true,
        "server.addr": ":8080",
    })

// 从字节数据加载
orz.NewFramework("my-app", "1.0.0").
    WithConfigBytes(jsonData, "json")
```

### 应用层配置获取

```go
func (app *MyApp) Configure(a *app.App) error {
    // 方式1: 获取完整配置
    config := a.GetConfig()
    fmt.Printf("数据库类型: %s", config.Database.Type)
    
    // 方式2: 获取具体配置值
    configMgr := a.ConfigManager()
    appName := configMgr.GetString("app.name")
    debugMode := configMgr.GetBool("app.debug")
    
    // 方式3: 解析到自定义结构
    var myConfig MyConfig
    err := configMgr.UnmarshalKey("app", &myConfig)
    
    // 方式4: 动态设置配置
    configMgr.Set("runtime.started_at", time.Now())
    
    return nil
}
```

## 🛡️ 类型安全的服务容器

告别强制类型转换：

```go
func (app *MyApp) Configure(a *app.App) error {
    // 类型安全的服务访问
    db := a.MustGetDatabase()    // *gorm.DB，编译时类型检查
    echo := a.MustGetEcho()      // *echo.Echo，自动类型推断
    
    // 带错误处理的版本
    if logger, err := a.GetLogger(); err == nil {
        logger.Info("应用启动")
    }
    
    return nil
}
```

## 🔧 核心概念

### 1. 应用容器

```go
app := orz.NewFramework("my-app", "1.0.0").
    WithConfig("config.yaml").
    WithDatabase().
    WithHTTP()
```

### 2. 模块系统

```go
type MyModule struct{}

func (m *MyModule) Name() string { return "my-module" }
func (m *MyModule) Init(app *app.App) error { /* 初始化逻辑 */ }
func (m *MyModule) Start(ctx context.Context) error { /* 启动逻辑 */ }
func (m *MyModule) Stop(ctx context.Context) error { /* 停止逻辑 */ }

// 注册模块
app.RegisterModule(&MyModule{})
```

### 3. 类型安全的仓库模式

```go
type User struct {
    ID   uint   `json:"id" gorm:"primaryKey"`
    Name string `json:"name"`
}

// 创建仓库
userRepo := repository.NewRepository[User, uint](func(ctx context.Context) *gorm.DB {
    return db
})

// 使用仓库
user, err := userRepo.FindById(ctx, 1)
users, err := userRepo.FindAll(ctx)
```

### 4. HTTP 响应助手

```go
helper := http.NewResponseHelper()

// 标准响应
helper.Success(c, data)
helper.PageSuccess(c, items, total)
helper.BadRequest(c, "Invalid input")
```

## 🎯 使用方式

### 快速启动

```go
orz.Quick("config.yaml", func(e *echo.Echo, db *gorm.DB) error {
    // 设置路由和业务逻辑
    return nil
})
```

### 完全控制

```go
type MyApp struct{}

func (a *MyApp) Configure(app *app.App) error {
    // 自定义配置逻辑
    return nil
}

func main() {
    orz.NewFramework("my-app", "1.0.0").
        WithConfigMap(testConfig).
        WithDatabase().
        WithHTTP().
        WithApplication(&MyApp{}).
        Run()
}
```

## 📚 示例项目

- `examples/simple/` - 基础 CRUD 应用
- `examples/advanced/` - 企业级应用结构  
- `examples/config_demo/` - 配置管理演示
- `examples/type_safe/` - 类型安全特性演示

## 📖 文档

- [配置使用指南](CONFIG_GUIDE.md) - 详细的配置管理文档
- [改进说明](IMPROVEMENTS.md) - 框架改进详情

## 🛠️ 开发

```bash
# 克隆项目
git clone https://github.com/go-orz/orz.git
cd orz

# 运行测试
./test_build.sh

# 运行示例
cd examples/simple && go run .
```

## 🎉 特点

- ✨ **现代化**: 充分利用 Go 1.18+ 的泛型特性
- 🛡️ **类型安全**: 减少运行时类型错误
- 🔧 **高度灵活**: 支持多种配置和部署方式  
- 📚 **易于使用**: 直观的 API 设计
- 🏗️ **架构清晰**: 模块化设计，职责分明
- 🧪 **测试友好**: 支持内存配置，便于单元测试

## 📄 许可证

MIT License