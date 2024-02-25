package main

import (
  "os"
	"net/http"

	"github.com/labstack/echo/v4"
)

const (
	APP_PORT = "3000"
)

func main() {
	e := echo.New()

	e.File("/", "public/index.html")

	e.GET("/api", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	e.Logger.Fatal(e.Start(":" + APP_PORT))
}
