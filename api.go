package orz

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"strconv"
)

type Api struct {
}

func (api Api) GetPagination(c echo.Context) (int, int) {
	pageIndex, err := strconv.Atoi(c.QueryParam("pageIndex"))
	if err != nil {
		pageIndex = 1
	}
	pageSize, err := strconv.Atoi(c.QueryParam("pageSize"))
	if err != nil {
		pageSize = 10
	}
	return pageIndex, pageSize
}

func (api Api) GetPageRequest(c echo.Context, allowedSortFields ...string) *PageRequest {
	sortJson := c.QueryParam("sort")
	var sort = make(map[string]string)
	_ = json.Unmarshal([]byte(sortJson), &sort)
	var (
		order string
		field string
	)
	for k, v := range sort {
		field = k
		order = v
	}
	if field == "" && len(allowedSortFields) > 0 {
		field = allowedSortFields[0]
	}
	pageIndex, pageSize := api.GetPagination(c)
	pr := PageRequest{
		PageIndex: pageIndex,
		PageSize:  pageSize,
		Sort:      NewSort(SortDirection(order), field, allowedSortFields...),
	}

	return &pr
}

func (api Api) Ok(c echo.Context, data interface{}) error {
	return c.JSON(200, data)
}

func (api Api) Okay(c echo.Context, items interface{}, total int64) error {
	return c.JSON(200, map[string]interface{}{
		"items": items,
		"total": total,
	})
}

func (api Api) ServerError(c echo.Context, data interface{}) error {
	return c.JSON(500, data)
}

func (api Api) BadRequest(c echo.Context, data interface{}) error {
	return c.JSON(400, data)
}
