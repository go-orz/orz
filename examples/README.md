# ORZ Framework 示例

本目录包含 ORZ 框架的核心使用示例，展示框架的主要功能。

所有示例都按需引入数据库驱动子模块，不会因为使用框架核心包而默认携带全部数据库驱动。

## 📁 示例目录

### 1. `simple/` - 基础使用示例
展示 ORZ 框架最简单的使用方式：
- 快速启动方式 (`orz.Quick`)
- 按需引入数据库驱动（示例里使用 SQLite）
- 基本的 HTTP 路由设置
- 数据库模型和 CRUD 操作
- 适合初学者快速上手

**运行方式：**
```bash
cd simple
go run main.go
```

### 2. `service_demo/` - Service 层演示
展示 ORZ 框架的 Service 模式和事务管理：
- Service 基类继承和使用
- Repository 模式
- 事务管理 (`Transaction`)
- 业务逻辑层设计
- 数据验证和错误处理

**运行方式：**
```bash
cd service_demo
go run main.go
```

### 3. `advanced_paging_demo/` - 高级分页演示
展示 ORZ 框架的高级分页功能：
- 安全的排序字段白名单验证
- SQL 注入防护演示
- 链式 API 调用
- 条件查询和模糊搜索
- 分页查询最佳实践

**运行方式：**
```bash
cd advanced_paging_demo
go run main.go
```

### 4. `clean_paging_demo/` - 精简分页演示
展示更直接的分页查询 API：
- 连表分页查询
- 多条件筛选
- 原类型和视图类型查询

## 🚀 快速开始

1. **克隆项目**
   ```bash
   git clone <repository-url>
   cd orz/examples
   ```

2. **选择示例运行**
   ```bash
   # 基础示例
   cd simple && go run main.go
   
   # Service 示例
   cd service_demo && go run main.go
   
   # 高级分页示例
   cd advanced_paging_demo && go run main.go

   # 精简分页示例
   cd clean_paging_demo && go run main.go
   ```

## 📚 学习路径

建议按以下顺序学习示例：

1. **`simple`** - 了解框架基础用法
2. **`service_demo`** - 学习业务层设计和事务管理
3. **`advanced_paging_demo`** - 掌握高级查询和视图查询
4. **`clean_paging_demo`** - 对比更精简的分页 API

每个示例都包含详细的注释和说明，帮助理解 ORZ 框架的设计理念和最佳实践。

## 🛠️ 技术特性

这些示例展示了 ORZ 框架的核心特性：

- **类型安全** - 基于 Go 泛型的类型安全设计
- **依赖注入** - 灵活的服务容器和模块化架构
- **事务管理** - 简洁的事务处理机制
- **Repository 模式** - 标准化的数据访问层
- **安全防护** - SQL 注入防护和字段白名单验证
- **链式 API** - 流畅的 API 设计
- **配置管理** - 灵活的配置加载方式

## 📖 更多文档

请参考项目根目录的 README.md 获取更详细的文档和 API 说明。
