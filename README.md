TWITTER CLONE IN GO: FROM ZERO TO 100K RPS
==========================================

This is a personal exercise project for me familiarizing with Go API development, while also exercising creating high performance API. I'll try to go along documenting my learning, while also aiming & maintaining the 100k RPS metric. This project would also assuming fullstack application setup, although I would focused more on the backend/API stuffs.

## The Stack
For this project, I would assume this stack:
- Go with Echo framework
- PostgreSQL as database storage
- Docker Compose for orchestration
- Load Testing with wrk & k6
- MVP.css & vanilla javascript for the frontend
- Other minor libraries & tools I'll add along with this project

## Running
```
# using go
go run server.go

# using docker compose
docker compose up -d

# after that you could access the site
curl localhost:3000
````
