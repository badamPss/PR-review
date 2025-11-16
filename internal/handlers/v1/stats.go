package v1

import (
	"net/http"
	"pr-review/internal/handlers/v1/dto"

	"github.com/labstack/echo/v4"
)

func (a *API) registerStatsHandlers(group *echo.Group) {
	group.GET("/stats", a.getStats)
}

func (a *API) getStats(c echo.Context) error {
	ctx := c.Request().Context()

	stats, err := a.service.GetStats(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, dto.FromModelStats(stats))
}
