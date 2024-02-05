package main

import (
  "os"
	//"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// helper for setting env vars
func getEnvOrDefault(envName string, defaultValue string) string {
  val := os.Getenv(envName)
  if val != "" {
    return val
  } else {
    return defaultValue
  }
}

func main() {
	APP_PORT := getEnvOrDefault("APP_PORT",  "3000")

	e := echo.New()

	e.File("/", "public/index.html")

	e.GET("/api", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	e.Logger.Fatal(e.Start(":" + APP_PORT))
}
