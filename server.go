package main

import (
  "os"
	//"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func main() {
	APP_PORT := os.Getenv("APP_PORT")
	if APP_PORT == "" {
		APP_PORT = "3000"
	}

	e := echo.New()

	e.File("/", "public/index.html")

	e.GET("/api/", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	e.Logger.Fatal(e.Start(":" + APP_PORT))
}
