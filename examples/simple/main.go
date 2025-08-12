package main

import (
	"log"
	"net/http"

	"github.com/go-orz/orz"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
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
	err := orz.Quick("config.yaml", func(e *echo.Echo, db *gorm.DB) error {
		// 自动迁移
		if err := db.AutoMigrate(&User{}); err != nil {
			return err
		}

		// 设置路由
		e.GET("/", func(c echo.Context) error {
			return c.JSON(http.StatusOK, map[string]string{
				"message": "Hello from ORZ framework!",
			})
		})

		e.GET("/users", func(c echo.Context) error {
			var users []User
			if err := db.Find(&users).Error; err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": err.Error(),
				})
			}
			return c.JSON(http.StatusOK, users)
		})

		e.POST("/users", func(c echo.Context) error {
			var user User
			if err := c.Bind(&user); err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"error": err.Error(),
				})
			}

			if err := db.Create(&user).Error; err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": err.Error(),
				})
			}

			return c.JSON(http.StatusCreated, user)
		})

		return nil
	})

	if err != nil {
		log.Fatal("Failed to start application:", err)
	}
}
