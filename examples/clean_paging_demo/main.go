package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-orz/orz"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Product 产品模型
type Product struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	Name       string `json:"name"`
	CategoryID uint   `json:"category_id"`
	Price      int    `json:"price"` // 分为单位
	Status     string `json:"status"`
}

func (Product) TableName() string {
	return "products"
}

// Category 分类模型
type Category struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Name string `json:"name"`
}

func (Category) TableName() string {
	return "categories"
}

// ProductListView 产品列表视图（带分类信息）
type ProductListView struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	CategoryName string `json:"category_name"`
	Price        int    `json:"price"`
	Status       string `json:"status"`
}

// ProductRepo 产品仓库
type ProductRepo struct {
	orz.Repository[Product, uint]
}

// NewProductRepo 创建产品仓库
func NewProductRepo(db *gorm.DB) *ProductRepo {
	return &ProductRepo{
		Repository: orz.NewRepository[Product, uint](db),
	}
}

// PageWithCategory 查询带分类信息的产品列表
func (r *ProductRepo) PageWithCategory(ctx context.Context, pageIndex, pageSize int, nameKeyword, status string) (*orz.PageResult[ProductListView], error) {
	builder := orz.Query(r.Repository).
		Index(pageIndex).
		Size(pageSize).
		Select("products.*, categories.name as category_name").
		LeftJoin("categories", "products.category_id = categories.id").
		SortByDesc("id", "id", "name", "price")

	// 添加名称关键词搜索
	if nameKeyword != "" {
		builder = builder.Contains("name", nameKeyword)
	}

	// 添加状态筛选
	if status != "" {
		builder = builder.Equal("status", status)
	}

	return orz.ExecuteAsTyped[ProductListView](ctx, builder)
}

// PageActiveProducts 查询活跃产品（多条件示例）
func (r *ProductRepo) PageActiveProducts(ctx context.Context, pageIndex, pageSize int, categories []uint) (*orz.PageResult[ProductListView], error) {
	builder := orz.Query(r.Repository).
		Index(pageIndex).
		Size(pageSize).
		Select("products.*, categories.name as category_name").
		LeftJoin("categories", "products.category_id = categories.id").
		Equal("products.status", "active").     // 只查询活跃产品
		NotEqual("products.price", 0).          // 排除免费产品
		SortBy("products.price", "price", "id") // 按价格升序

	// 可选：按分类筛选
	if len(categories) > 0 {
		builder = builder.In("products.category_id", categories)
	}

	return orz.ExecuteAsTyped[ProductListView](ctx, builder)
}

// PageBasicProducts 基础产品查询（原类型）
func (r *ProductRepo) PageBasicProducts(ctx context.Context, pageIndex, pageSize int, priceRange []int) (*orz.PageResult[Product], error) {
	builder := orz.Query(r.Repository).
		Index(pageIndex).
		Size(pageSize).
		SortByDesc("id", "id", "name", "price")

	// 价格范围筛选
	if len(priceRange) == 2 {
		// 这里演示如何处理复杂条件，虽然当前框架还不支持范围查询
		// 可以考虑扩展 PageBuilder 支持 GreaterThan, LessThan 等
		// builder = builder.GreaterThanOrEqual("price", priceRange[0]).LessThanOrEqual("price", priceRange[1])
	}

	return builder.Execute(ctx)
}

// CleanPagingDemoApp 清洁分页示例应用
type CleanPagingDemoApp struct{}

func (a *CleanPagingDemoApp) Configure(app *orz.App) error {
	// 获取数据库并自动迁移
	db, err := app.GetDatabase()
	if err != nil {
		return err
	}

	if err := db.AutoMigrate(&Product{}, &Category{}); err != nil {
		return err
	}

	// 插入测试数据
	categories := []Category{
		{ID: 1, Name: "电子产品"},
		{ID: 2, Name: "服装"},
		{ID: 3, Name: "图书"},
	}
	db.Create(&categories)

	products := []Product{
		{ID: 1, Name: "iPhone 15", CategoryID: 1, Price: 599900, Status: "active"},
		{ID: 2, Name: "MacBook Pro", CategoryID: 1, Price: 199900, Status: "active"},
		{ID: 3, Name: "休闲T恤", CategoryID: 2, Price: 9900, Status: "active"},
		{ID: 4, Name: "牛仔裤", CategoryID: 2, Price: 29900, Status: "inactive"},
		{ID: 5, Name: "Go语言编程", CategoryID: 3, Price: 6900, Status: "active"},
		{ID: 6, Name: "设计模式", CategoryID: 3, Price: 8900, Status: "active"},
	}
	db.Create(&products)

	// 获取Echo并设置路由
	e, err := app.GetEcho()
	if err != nil {
		return err
	}

	// 创建仓库
	productRepo := NewProductRepo(db)

	// 连表查询：带分类信息的产品列表
	e.GET("/products", func(c echo.Context) error {
		pageIndex := 1
		pageSize := 10
		keyword := c.QueryParam("keyword")
		status := c.QueryParam("status")

		result, err := productRepo.PageWithCategory(c.Request().Context(), pageIndex, pageSize, keyword, status)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "产品列表（带分类信息）",
			"data":    result,
		})
	})

	// 多条件查询：活跃产品
	e.GET("/products/active", func(c echo.Context) error {
		pageIndex := 1
		pageSize := 10
		// 这里简化处理，实际可以从查询参数解析分类ID数组
		categories := []uint{1, 3} // 只查询电子产品和图书

		result, err := productRepo.PageActiveProducts(c.Request().Context(), pageIndex, pageSize, categories)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "活跃产品列表（多条件筛选）",
			"data":    result,
		})
	})

	// 基础查询：原类型产品列表
	e.GET("/products/basic", func(c echo.Context) error {
		pageIndex := 1
		pageSize := 10

		result, err := productRepo.PageBasicProducts(c.Request().Context(), pageIndex, pageSize, nil)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "基础产品列表（原类型）",
			"data":    result,
		})
	})

	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message":     "清洁分页查询示例 - 无modifier的PageBuilder",
			"description": "展示移除modifier后更清晰的API设计",
			"endpoints": []string{
				"GET /products?keyword=iPhone&status=active - 连表查询（带搜索和状态筛选）",
				"GET /products/active - 多条件查询（只查询活跃产品）",
				"GET /products/basic - 基础查询（原类型）",
			},
			"improvements": []string{
				"✅ 移除了复杂的modifier函数",
				"✅ 提供了专门的Select()、LeftJoin()等方法",
				"✅ 增加了NotEqual()、NotContains()、NotIn()等查询条件",
				"✅ 支持InnerJoin()、RightJoin()等多种连接类型",
				"✅ API更加清晰和类型安全",
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
			"addr": ":8084",
		},
	}

	app := &CleanPagingDemoApp{}

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

	log.Println("🚀 清洁PageBuilder示例启动成功！")
	log.Println("📖 访问 http://localhost:8084 查看API改进")
	log.Println("🔍 测试连表查询：GET /products?keyword=iPhone&status=active")
	log.Println("🔍 测试多条件查询：GET /products/active")

	if err := framework.Run(); err != nil {
		log.Fatal("运行失败:", err)
	}
}
