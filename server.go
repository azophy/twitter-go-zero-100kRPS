package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
  "strings"

	"github.com/labstack/echo-contrib/pprof"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

type Post struct {
	ID        string    `json:"id" xml:"id" form:"id" query:"id"`
	Username  string    `json:"username" xml:"username" form:"username" query:"username"`
	Content   string    `json:"content" xml:"content" form:"content" query:"content"`
	timestamp time.Time `json:"timestamp" xml:"timestamp" form:"timestamp" query:"timestamp"`
}

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
	APP_PORT := getEnvOrDefault("APP_PORT", "3000")
	DB_URI := getEnvOrDefault("DB_URI", "postgres://postgres:postgres@postgres/postgres?sslmode=disable")
	PROFILING_ENABLED := getEnvOrDefault("PROFILING_ENABLED", "false")
	// reference: https://pkg.go.dev/database/sql#DB.SetMaxOpenConns
	DB_MAX_OPEN_CONNECTION, _ := strconv.Atoi(getEnvOrDefault("DB_MAX_OPEN_CONNECTION", "100"))
	DB_MAX_IDLE_CONNECTION, _ := strconv.Atoi(getEnvOrDefault("DB_MAX_IDLE_CONNECTION", "10"))
	// in seconds
	DB_MAX_CONN_LIFETIME, _ := strconv.Atoi(getEnvOrDefault("DB_MAX_CONN_LIFETIME", "600"))

	postsCache := Cache[[]Post]{}

	fmt.Println("Connecting to database...")
	db_conn, err := sql.Open("postgres", DB_URI)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	db_conn.SetMaxOpenConns(DB_MAX_OPEN_CONNECTION)
	db_conn.SetMaxIdleConns(DB_MAX_IDLE_CONNECTION)
	db_conn.SetConnMaxLifetime(time.Duration(DB_MAX_CONN_LIFETIME) * time.Second)

	fmt.Println("Starting migration...")
	migration := `
    CREATE TABLE IF NOT EXISTS posts (
      id SERIAL PRIMARY KEY,
      username VARCHAR (30) NOT NULL,
      content VARCHAR (150) NOT NULL,
      timestamp TIMESTAMP NOT NULL DEFAULT NOW()
    );
    CREATE INDEX IF NOT EXISTS idx_posts_timestamp ON posts(timestamp);
  `
	_, err = db_conn.Exec(migration)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// define prepared statements
	readStmt, err := db_conn.Prepare("SELECT * FROM posts order by timestamp DESC LIMIT 10")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	//writeStmt, err := db_conn.Prepare("insert into posts(username, content) values ($1, $2)")
	//if err != nil {
		//fmt.Println(err.Error())
		//return
	//}

	// route definitions
	e := echo.New()

	if PROFILING_ENABLED == "true" {
		pprof.Register(e)
	}

	e.File("/", "public/index.html")

	e.GET("/api", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	getPostsFromDb := func() ([]Post, error) {
		var posts []Post

		rows, err := readStmt.Query()
		if err != nil {
			fmt.Println(err.Error())
			//return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			return posts, err
		}

		defer rows.Close()
		for rows.Next() {
			var post Post
			var timestamp string

			err := rows.Scan(&post.ID, &post.Username, &post.Content, &timestamp)
			if err != nil {
				fmt.Println(err.Error())
				return posts, err
			}

			post.timestamp, _ = time.Parse(time.RFC3339, timestamp)
			posts = append(posts, post)
		}

		return posts, nil
	}

	e.GET("/api/posts", func(c echo.Context) error {
		posts, err := postsCache.Fetch(500*time.Millisecond, getPostsFromDb)
		//posts, err := getPostsFromDb()
		if err != nil {
			fmt.Println(err.Error())
			//return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			return err
		}

		return c.JSON(http.StatusOK, posts)
	})

  postToDb := func(params [][]any) error {
    var queryParams []any
    query := "insert into posts(username, content) values "
    for idx, item := range params {
      query += fmt.Sprintf("($%d, $%d), ", (idx*2+1), (idx*2+2))
      queryParams = append(queryParams, item...)
    }
    _, err := db_conn.Exec(strings.TrimRight(query, ", "), queryParams...)
		if err != nil {
			fmt.Println(err.Error())
			//return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			return err
		}

    return nil
  }

	e.POST("/api/posts", func(c echo.Context) error {
		username := c.FormValue("username")
		content := c.FormValue("content")

		err = postToDb([][]any{ {username, content} })
		if err != nil {
			fmt.Println(err.Error())
			//return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			return err
		}
		return c.HTML(http.StatusOK, "ok. return to <a href='/'>homepage</a>")
	})

	e.Logger.Fatal(e.Start(":" + APP_PORT))
}
