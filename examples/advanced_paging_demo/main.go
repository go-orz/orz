package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-orz/orz"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// License 许可证模型
type License struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	CreatedBy uint   `json:"created_by"`
}

func (License) TableName() string {
	return "licenses"
}

// User 用户模型
type User struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Name string `json:"name"`
}

func (User) TableName() string {
	return "users"
}

// LicenseListView 许可证列表视图（带创建者信息）
type LicenseListView struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Creator string `json:"creator"` // 创建者名称
}

// LicenseRepo 许可证仓库
type LicenseRepo struct {
	orz.Repository[License, uint]
}

// NewLicenseRepo 创建许可证仓库
func NewLicenseRepo(db *gorm.DB) *LicenseRepo {
	return &LicenseRepo{
		Repository: orz.NewRepository[License, uint](db),
	}
}

// PageWithCreator 使用PageBuilder查询带创建者信息的许可证列表
func (r *LicenseRepo) PageWithCreator(ctx context.Context, pageIndex, pageSize int, nameKeyword string) (*orz.PageResult[LicenseListView], error) {
	builder := orz.Query(r.Repository).
		PageIndex(pageIndex).
		PageSize(pageSize).
		Select("licenses.*, users.name as creator").
		LeftJoin("users", "licenses.created_by = users.id").
		SortByDesc("id", "id", "name", "type")

	// 添加搜索条件
	if nameKeyword != "" {
		builder = builder.Contains("name", nameKeyword)
	}

	// 执行泛型查询
	return orz.ExecuteAsTyped[LicenseListView](ctx, builder)
}

// PageBasic 基础分页查询（原类型）
func (r *LicenseRepo) PageBasic(ctx context.Context, pageIndex, pageSize int, typeFilter string) (*orz.PageResult[License], error) {
	builder := orz.Query(r.Repository).
		PageIndex(pageIndex).
		PageSize(pageSize).
		SortByDesc("id", "id", "name", "type")

	// 添加类型筛选
	if typeFilter != "" {
		builder = builder.Equal("type", typeFilter)
	}

	// 执行查询（返回原类型）
	return builder.Execute(ctx)
}

// AdvancedPagingDemoApp 高级分页示例应用
type AdvancedPagingDemoApp struct{}

func (a *AdvancedPagingDemoApp) Configure(app *orz.App) error {
	// 获取数据库并自动迁移
	db := app.GetDatabase()
	if err := db.AutoMigrate(&License{}, &User{}); err != nil {
		return err
	}

	// 插入测试数据
	users := []User{
		{ID: 1, Name: "张三"},
		{ID: 2, Name: "李四"},
		{ID: 3, Name: "王五"},
	}
	db.Create(&users)

	licenses := []License{
		{ID: 1, Name: "MIT License", Type: "开源", CreatedBy: 1},
		{ID: 2, Name: "Apache License", Type: "开源", CreatedBy: 2},
		{ID: 3, Name: "Commercial License", Type: "商业", CreatedBy: 1},
		{ID: 4, Name: "GPL License", Type: "开源", CreatedBy: 3},
		{ID: 5, Name: "BSD License", Type: "开源", CreatedBy: 2},
	}
	db.Create(&licenses)

	// 获取Echo并设置路由
	e := app.GetEcho()
	// 创建仓库
	licenseRepo := NewLicenseRepo(db)

	// 连表查询：带创建者信息的许可证列表
	e.GET("/licenses", func(c echo.Context) error {
		pageIndex := 1
		pageSize := 10
		keyword := c.QueryParam("keyword")

		result, err := licenseRepo.PageWithCreator(c.Request().Context(), pageIndex, pageSize, keyword)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "许可证列表（带创建者信息）",
			"data":    result,
		})
	})

	// 基础查询：原类型的许可证列表
	e.GET("/licenses/basic", func(c echo.Context) error {
		pageIndex := 1
		pageSize := 10
		typeFilter := c.QueryParam("type")

		result, err := licenseRepo.PageBasic(c.Request().Context(), pageIndex, pageSize, typeFilter)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "基础许可证列表",
			"data":    result,
		})
	})

	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message":     "分页查询示例 - PageBuilder",
			"description": "使用优雅的PageBuilder进行分页查询",
			"endpoints": []string{
				"GET /licenses?keyword=MIT - 连表查询（带创建者信息）",
				"GET /licenses/basic?type=开源 - 基础查询（原类型）",
			},
		})
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
			"addr": ":8083",
		},
	}

	app := &AdvancedPagingDemoApp{}

	framework, err := orz.NewFramework(
		orz.WithConfigMap(configMap),
		orz.WithLoggerFromConfig(),
		orz.WithDatabase(),
		orz.WithHTTP(),
		orz.WithApplication(app),
	)
	if err != nil {
		log.Fatal("框架初始化失败:", err)
	}

	log.Println("🚀 PageBuilder分页查询示例启动成功！")
	log.Println("📖 访问 http://localhost:8083 查看可用接口")
	log.Println("🔍 测试连表查询：GET /licenses?keyword=MIT")
	log.Println("🔍 测试基础查询：GET /licenses/basic?type=开源")

	if err := framework.Run(); err != nil {
		log.Fatal("运行失败:", err)
	}
}
