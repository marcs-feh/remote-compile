package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func setupRoutes(e *echo.Echo){
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "hello")
	})
}


