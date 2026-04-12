package main

import (
	"log"

	"github.com/go-orz/orz"
	_ "github.com/go-orz/orz/drivers/sqlite"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// User 用户模型
type User struct {
	ID    uint   `json:"id" gorm:"primaryKey"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (User) TableName() string {
	return "users"
}

// OptionDemoApp Option模式示例应用
type OptionDemoApp struct{}

func (a *OptionDemoApp) Configure(app *orz.App) error {
	// 获取数据库并自动迁移
	db := app.GetDatabase()

	if err := db.AutoMigrate(&User{}); err != nil {
		return err
	}

	// 获取Echo并设置路由
	e := app.GetEcho()

	e.GET("/", func(c echo.Context) error {
		return orz.Message(c, 200, "Option Demo - Hello from ORZ!")
	})

	e.GET("/users", func(c echo.Context) error {
		var users []User
		if err := db.Find(&users).Error; err != nil {
			return orz.InternalServerError(c, err.Error())
		}
		return orz.Ok(c, users)
	})

	return nil
}

func main() {
	// 配置映射
	configMap := map[string]interface{}{
		"log": map[string]interface{}{
			"level":   "info",
			"encode":  "console",
			"console": true,
		},
		"database": map[string]interface{}{
			"type": "sqlite",
			"sqlite": map[string]interface{}{
				"path": ":memory:",
			},
		},
		"server": map[string]interface{}{
			"addr": ":8082",
		},
	}

	// 创建自定义logger
	customLogger, _ := zap.NewDevelopment()

	// 示例1: 使用所有选项
	log.Println("=== 示例1: 完整配置 ===")
	app := &OptionDemoApp{}
	framework1, err := orz.NewFramework(
		orz.WithConfigMap(configMap), // 配置
		orz.WithLogger(customLogger), // 自定义logger
		orz.WithDatabase(),           // 数据库
		orz.WithHTTP(),               // HTTP服务
		orz.WithApplication(app),     // 应用
	)
	if err != nil {
		log.Fatal("框架1初始化失败:", err)
	}

	// 示例2: 仅数据库，无HTTP
	log.Println("=== 示例2: 仅数据库模式 ===")
	framework2, err := orz.NewFramework(
		orz.WithConfigMap(configMap),
		orz.WithLoggerFromConfig(), // 从配置读取logger
		orz.WithDatabase(),
		// 不使用 WithHTTP() - 将以守护进程模式运行
	)
	if err != nil {
		log.Fatal("框架2初始化失败:", err)
	}

	// 这里可以选择运行哪个框架
	_ = framework2 // 占位，避免unused变量警告

	// 运行完整配置的框架
	if err := framework1.Run(); err != nil {
		log.Fatal("运行失败:", err)
	}
}
