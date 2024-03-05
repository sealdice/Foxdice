package serve

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func bindAPI(api *echo.Group) {
	api.GET("hello", func(c echo.Context) error {
		return c.JSON(http.StatusOK, nil)
	})
}
