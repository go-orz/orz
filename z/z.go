package z

import "github.com/labstack/echo/v4"

const Token = "X-Auth-Token"

func GetToken(c echo.Context) string {
	token := c.Request().Header.Get(Token)
	if len(token) > 0 {
		return token
	}
	token = c.QueryParam(Token)
	if token != "" {
		return token
	}
	cookie, err := c.Cookie(Token)
	if err != nil {
		return ""
	}
	return cookie.Value
}

var AccountId func(c echo.Context) string
