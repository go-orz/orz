package main

import (
	"log"

	"github.com/go-orz/orz"
	_ "github.com/go-orz/orz/drivers/sqlite"
	"github.com/labstack/echo/v5"
)

// User 用户模型示例
type User struct {
	ID    uint   `json:"id" gorm:"primaryKey"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (User) TableName() string {
	return "users"
}

func main() {
	// 快速启动方式
	err := orz.Quick("config.yaml", func(app *orz.App) error {
		db := app.GetDatabase()
		e := app.GetEcho()

		// 自动迁移
		if err := db.AutoMigrate(&User{}); err != nil {
			return err
		}

		// 设置路由
		e.GET("/", func(c *echo.Context) error {
			return orz.Message(c, 200, "Hello from ORZ framework!")
		})

		e.GET("/users", func(c *echo.Context) error {
			var users []User
			if err := db.Find(&users).Error; err != nil {
				return orz.InternalServerError(c, err.Error())
			}
			return orz.Ok(c, users)
		})

		e.POST("/users", func(c *echo.Context) error {
			var user User
			if err := c.Bind(&user); err != nil {
				return orz.BadRequest(c, err.Error())
			}

			if err := db.Create(&user).Error; err != nil {
				return orz.InternalServerError(c, err.Error())
			}

			return orz.Created(c, user)
		})

		return nil
	})

	if err != nil {
		log.Fatal("Failed to start application:", err)
	}
}
