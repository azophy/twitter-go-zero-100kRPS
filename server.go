package main

import (
  "os"
  "fmt"
  "time"
  "context"
  "strconv"
	"net/http"
  "database/sql"
  "encoding/json"

  _ "github.com/lib/pq"

  "github.com/redis/go-redis/v9"
	"github.com/labstack/echo/v4"
  "github.com/labstack/echo-contrib/pprof"
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

var (
  backgroundCtx = context.Background()
  rdb *redis.Client

  db_conn *sql.DB
  readStmt *sql.Stmt
  writeStmt *sql.Stmt
)

func getPostHandler(c echo.Context) error {
  var posts []Post

  // retrieve from cache
  postJson, err := rdb.Get(backgroundCtx, "posts").Result()
  if err == nil {
    err = json.Unmarshal([]byte(postJson), &posts)
    if err != nil {
        fmt.Println(err.Error())
        //return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
        return err
    }

    return c.JSON(http.StatusOK, posts)
  }

  // lock before updating cache
  _, err = rdb.SetNX(backgroundCtx, "post:lock", "true", 1*time.Second).Result()
  if err != nil {
    time.Sleep(100 * time.Millisecond)
    return getPostHandler(c)
  }
  defer rdb.Del(backgroundCtx, "post:lock")

  // retrieve from DB
  rows, err := readStmt.Query()
  if err != nil {
      fmt.Println(err.Error())
      //return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
      return err
  }

  defer rows.Close()

  for rows.Next() {
    var post Post
    var timestamp string

    err := rows.Scan(&post.ID, &post.Username, &post.Content, &timestamp)
    if err != nil {
      fmt.Println(err.Error())
      return err
    }

    post.timestamp, _ = time.Parse(time.RFC3339, timestamp)
    posts = append(posts, post)
  }

  postJsonByte, err := json.Marshal(posts)
  if err != nil {
      fmt.Println(err.Error())
      //return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
      return err
  }

  err = rdb.Set(backgroundCtx, "posts", postJsonByte, 2 * time.Second).Err()
  if err != nil {
      fmt.Println(err.Error())
      //return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
      return err
  }

  return c.JSON(http.StatusOK, posts)
}

func main() {
	APP_PORT := getEnvOrDefault("APP_PORT",  "3000")
  DB_URI := getEnvOrDefault("DB_URI", "postgres://postgres:postgres@postgres/postgres?sslmode=disable")
  REDIS_URL := getEnvOrDefault("REDIS_URL", "redis://redis:6379/0?protocol=3")
  PROFILING_ENABLED := getEnvOrDefault("PROFILING_ENABLED", "false")
  // reference: https://pkg.go.dev/database/sql#DB.SetMaxOpenConns
  DB_MAX_OPEN_CONNECTION, _ := strconv.Atoi(getEnvOrDefault("DB_MAX_OPEN_CONNECTION", "100"))
  DB_MAX_IDLE_CONNECTION, _ := strconv.Atoi(getEnvOrDefault("DB_MAX_IDLE_CONNECTION", "10"))
  // in seconds
  DB_MAX_CONN_LIFETIME, _ := strconv.Atoi(getEnvOrDefault("DB_MAX_CONN_LIFETIME", "600"))

  // connect to redis
  fmt.Println("Connecting to redis...")
  redisOpts, err := redis.ParseURL(REDIS_URL)
  if err != nil {
      panic(err)
  }
  rdb = redis.NewClient(redisOpts)
  defer rdb.Close()

  status, err := rdb.Ping(backgroundCtx).Result()
  if err != nil {
      fmt.Println("Redis connection was refused")
  }
  fmt.Println("Redis connected with status: " + status)

  // connect to postgres
  fmt.Println("Connecting to database...")
  db_conn, err = sql.Open("postgres", DB_URI)
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
  readStmt, err = db_conn.Prepare("SELECT * FROM posts order by timestamp DESC LIMIT 10")
  if err != nil {
      fmt.Println(err.Error())
      return
  }
  writeStmt, err = db_conn.Prepare("insert into posts(username, content) values ($1, $2)")
  if err != nil {
      fmt.Println(err.Error())
      return
  }

  // route definitions
	e := echo.New()

  if (PROFILING_ENABLED == "true") {
    pprof.Register(e)
  }

	e.File("/", "public/index.html")

	e.GET("/api", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

  e.GET("/api/posts", getPostHandler)

	e.POST("/api/posts", func(c echo.Context) error {
    username := c.FormValue("username")
    content := c.FormValue("content")

    _, err = writeStmt.Exec(username, content)
    if err != nil {
        fmt.Println(err.Error())
        //return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
        return err
    }
		return c.HTML(http.StatusOK, "ok. return to <a href='/'>homepage</a>")
	})

	e.Logger.Fatal(e.Start(":" + APP_PORT))
}
