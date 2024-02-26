package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

const (
	APP_PORT = "3000"
	DB_URI   = "postgres://postgres:postgres@postgres/postgres?sslmode=disable"
)

type Post struct {
	ID        string    `json:"id" xml:"id" form:"id" query:"id"`
	Username  string    `json:"username" xml:"username" form:"username" query:"username"`
	Content   string    `json:"content" xml:"content" form:"content" query:"content"`
	Timestamp time.Time `json:"timestamp" xml:"timestamp" form:"timestamp" query:"timestamp"`
}

func main() {
	fmt.Println("Connecting to database...")
	db_conn, err := sql.Open("postgres", DB_URI)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Starting migration...")
	migration := `
    CREATE TABLE IF NOT EXISTS posts (
      id SERIAL PRIMARY KEY,
      username VARCHAR (30) NOT NULL,
      content VARCHAR (150) NOT NULL,
      timestamp TIMESTAMP NOT NULL DEFAULT NOW()
    );
  `
	_, err = db_conn.Exec(migration)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	e := echo.New()

	e.File("/", "index.html")

	e.GET("/api", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	e.GET("/api/posts", func(c echo.Context) error {
		rows, err := db_conn.Query("SELECT * FROM posts ORDER BY timestamp DESC LIMIT 10")
		if err != nil {
			fmt.Println(err.Error())
			return err
		}

		defer rows.Close()
		var posts []Post

		for rows.Next() {
			var post Post
			var timestamp string

			err := rows.Scan(&post.ID, &post.Username, &post.Content, &timestamp)
			if err != nil {
				fmt.Println(err.Error())
				return err
			}

			post.Timestamp, _ = time.Parse(time.RFC3339, timestamp)
			posts = append(posts, post)
		}

		return c.JSON(http.StatusOK, posts)
	})

	e.POST("/api/posts", func(c echo.Context) error {
		username := c.FormValue("username")
		content := c.FormValue("content")

		_, err = db_conn.Exec("INSERT INTO posts(username, content) VALUES ($1, $2)", username, content)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		return c.HTML(http.StatusOK, "ok. return to <a href='/'>homepage</a>")
	})
	e.Logger.Fatal(e.Start(":" + APP_PORT))
}
