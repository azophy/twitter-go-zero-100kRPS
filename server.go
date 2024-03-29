package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
  //"strings"

	"github.com/labstack/echo-contrib/pprof"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

const (
  POST_WORKER_TIMEOUT = 1000 // in milliseconds
  POST_WORKER_CHUCK_SIZE = 100
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
	DB_MAX_OPEN_CONNECTION, _ := strconv.Atoi(getEnvOrDefault("DB_MAX_OPEN_CONNECTION", "300"))
	DB_MAX_IDLE_CONNECTION, _ := strconv.Atoi(getEnvOrDefault("DB_MAX_IDLE_CONNECTION", "200"))
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

  // prebuild all write queries
  var writeStmts []sql.Stmt
  writeQuery := "insert into posts(username, content) values"
  for i:=0; i<=POST_WORKER_CHUCK_SIZE; i++ {
    if (i>0) {
      writeQuery += ", "
    }
    writeQuery += fmt.Sprintf("($%d, $%d)", (i*2+1), (i*2+2))
    stmt, err := db_conn.Prepare(writeQuery)
    if err != nil {
      fmt.Println(err.Error())
      return
    }
    writeStmts = append(writeStmts, *stmt)
  }

	//writeStmt, err := db_conn.Prepare("insert into posts(username, content) values ($1, $2)")
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
    paramLength := len(params)
    //query := "insert into posts(username, content) values "
    for _, item := range params {
      //query += fmt.Sprintf("($%d, $%d), ", (idx*2+1), (idx*2+2))
      queryParams = append(queryParams, item...)
    }
    _, err := writeStmts[paramLength-1].Exec(queryParams...)
		if err != nil {
			fmt.Println(err.Error())
			//return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			return err
		}

    return nil
  }

  postChannel := make(chan []any)

  postWorker := func() {
    var postBuffer [][]any
    for { // run as background process
      //fmt.Println("starting cycle")
      loop:
        for { // waiting for input/timeout
          select {
            case item := <- postChannel:
              postBuffer = append(postBuffer, item)
              //fmt.Println("get data:", item)
              if (len(postBuffer) >= POST_WORKER_CHUCK_SIZE) {
                //fmt.Println("postBuffer full:", postBuffer)
                break loop
              }
            case <-time.After(time.Millisecond * POST_WORKER_TIMEOUT):
              //fmt.Printf("timeout. no activities under %d milliseconds: %v\n", POST_WORKER_TIMEOUT, postBuffer)
              break loop
          }
        }
      // end of cycle
      if (len(postBuffer) > 0) {
        go postToDb(postBuffer)
        postBuffer = [][]any{}
        //fmt.Println("clearing buffer:", postBuffer)
      }
    }
  }

  go postWorker()
  go postWorker()
  go postWorker()

	e.POST("/api/posts", func(c echo.Context) error {
		username := c.FormValue("username")
		content := c.FormValue("content")

    postChannel <- []any{username, content}

		if err != nil {
			fmt.Println(err.Error())
			//return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			return err
		}
		return c.HTML(http.StatusOK, "ok. return to <a href='/'>homepage</a>")
	})

	e.Logger.Fatal(e.Start(":" + APP_PORT))
}
