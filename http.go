package orz

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// Response 标准响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Ok 成功响应
func Ok(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "ok",
		Data:    data,
	})
}

// ErrorResponse 错误响应
func ErrorResponse(c echo.Context, code int, message string) error {
	return c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
	})
}

// BadRequest 400错误
func BadRequest(c echo.Context, message string) error {
	return c.JSON(http.StatusBadRequest, Response{
		Code:    400,
		Message: message,
	})
}

// Unauthorized 401错误
func Unauthorized(c echo.Context, message string) error {
	return c.JSON(http.StatusUnauthorized, Response{
		Code:    401,
		Message: message,
	})
}

// Forbidden 403错误
func Forbidden(c echo.Context, message string) error {
	return c.JSON(http.StatusForbidden, Response{
		Code:    403,
		Message: message,
	})
}

// NotFound 404错误
func NotFound(c echo.Context, message string) error {
	return c.JSON(http.StatusNotFound, Response{
		Code:    404,
		Message: message,
	})
}

// InternalServerError 500错误
func InternalServerError(c echo.Context, message string) error {
	return c.JSON(http.StatusInternalServerError, Response{
		Code:    500,
		Message: message,
	})
}

// PageRequest 分页信息
type PageRequest struct {
	PageIndex         int       `json:"pageIndex"` // 页码，默认为1
	PageSize          int       `json:"pageSize"`  // 每页数量，默认为10
	SortField         string    `json:"sortField"` // 排序字段，例如 "created_at"
	SortOrder         SortOrder `json:"sortOrder"` // 排序方向，"asc" 或 "desc"
	SortAllowedFields []string  `json:"-"`         // 允许的排序字段
}

// GetPageRequest 从查询参数获取分页信息
func GetPageRequest(c echo.Context, allowedFields ...string) *PageRequest {
	var pr = &PageRequest{
		PageIndex:         1,
		PageSize:          10,
		SortField:         "",
		SortOrder:         "",
		SortAllowedFields: allowedFields,
	}

	if pageStr := c.QueryParam("pageIndex"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			pr.PageIndex = p
		}
	}

	if sizeStr := c.QueryParam("pageSize"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 100 {
			pr.PageSize = s
		}
	}
	if sortField := c.QueryParam("sortField"); sortField != "" {
		pr.SortField = sortField
	}
	if sortOrder := c.QueryParam("sortOrder"); sortOrder != "" {
		pr.SortOrder = SortOrder(sortOrder)
	}

	return pr
}

// GetKeyword 从查询参数获取关键词搜索值
func GetKeyword(c echo.Context, paramName ...string) string {
	keyName := "keyword"
	if len(paramName) > 0 && paramName[0] != "" {
		keyName = paramName[0]
	}
	return c.QueryParam(keyName)
}

// BindAndValidate 绑定并验证请求数据
func BindAndValidate(c echo.Context, v interface{}) error {
	if err := c.Bind(v); err != nil {
		return fmt.Errorf("bind error: %w", err)
	}

	// 如果结构体实现了验证接口，则进行验证
	if validator, ok := v.(interface{ Validate() error }); ok {
		if err := validator.Validate(); err != nil {
			return fmt.Errorf("validation error: %w", err)
		}
	}

	return nil
}
