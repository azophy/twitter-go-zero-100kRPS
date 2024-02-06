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
# run docker compose
docker compose up -d

# run go-echo server
docker compose exec app go run server.go

# after that you could access the site
curl localhost:3000
````

## Guide outline
1. create basic web app. setup docker di for golang & postgres (with resource limitation). setup basic FE. db connection & migration. API for healtcheck, listing & posting. 
2. loadtest with WRK for each API. breakpoint test with k6. analyze. profiling
3. add caching. talk about race condition. use mutex. tune postgresql. loadtest & profiling
4. add like feature. create like table. add API for liking & displaying. update FE. loadtest & profiling
5. db optimization. materialized view. denormalizarestion. loadtest & profiling.
6. implement connection pool & read replicas. loadtest & profiling

- increase posttgfesql perf: https://stackoverflow.com/questions/5131266/increase-postgresql-write-speed-at-the-cost-of-likely-data-loss#5138794
- psql read-write replica: https://gist.github.com/JosimarCamargo/40f8636563c6e9ececf603e94c3affa7
