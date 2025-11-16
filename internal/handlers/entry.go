package handlers

import (
	"github.com/labstack/echo/v4"
)

const (
	APIEndpointName = ""
)

type HTTPController interface {
	RegisterHandlers(e *echo.Group)
}

func Register(e *echo.Echo, controllers ...HTTPController) {
	api := e.Group(APIEndpointName)
	for _, c := range controllers {
		c.RegisterHandlers(api)
	}
}
