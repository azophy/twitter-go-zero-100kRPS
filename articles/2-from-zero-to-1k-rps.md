# Iteration #1: zero to 1k RPS
outline:
- loadtest with WRK for each API
- breakpoint test with k6
- analyze
- profiling

content:
- when load testing, some things to consider:
  - spec resource. in the beginning we only need 1 CPU, but it would result in meager results
  - in my experience, I need 6 CPU for both app & tester to reach 100k RPS just for healthcheck endpoint
  - also check network. OS is limited to only 65k socket/connection. although if the request is fast enough we could handle 100k RPS in 1 machine.
  - not all requests are equal
  - running golang with 'go run' vs using compiled binary is different. with compiled binary in my experience we could even reach 250k RPS!
- start with wrk. with limited CPU & small wek params we only able to reach 1-5k RPS. after increasing the spec (both app & tester) + adding wrk params, we able to reach 100k.
  - dont forget to monitor resource. If using htop or the like its hard to monitor containers usages. Instead I used [ctop](https://github.com/bcicen/ctop). run with `ctop -a` to only monitor active containers.
  - `docker compose exec tester wrk -c 100 -t 10 --latency http://app:3000/api` -> small
  - `docker compose exec tester wrk -c 500 -t 50 --latency http://app:3000/api` -> large. notice the number of connection & num of threads
- are we done? lets check for DB-related APIs
  - first lets try inserting our web with some data
  - run wrk for our posts list endpoint
    - `docker compose exec tester wrk -c 500 -t 50 --latency http://app:3000/api/posts`
    - result, I only got ~1k RPS. if we see the server log, we also see something like: `pq: sorry, too many clients already`.
  - now lets test our posts insert endpoint
    - wrk support lua scripting for create POST from request
    - `docker compose exec tester wrk -c 500 -t 50 -s tests/wrk-post.lua --latency http://app:3000/api/posts`
    - this time, I only get ~200 RPS. way worse then the previous endpoint's result
  - its clear that there are lots to do before we could reach 100k RPS for real-world scenario
- after we get some data about our API, lets try profiling our API to get better understanding
  - reference:
  - first install package `github.com/labstack/echo-contrib/pprof`, then register in our echo app
  - re-run our app
  - try to generate some load. we could use the same WRK setup as previous. try to do several times to gate accurate results we could compare against each other
  - now in a new terminal run `go tool pprof -http=:5001 http://localhost:3000/debug/pprof/profile`. change the view into flamegraph to get more data.
    - in my case for our listing endpoint, DB operations took 29% of our time
    - while for our insert endpoint, DB operations took 64% of our time
- lets try to fix by first imposing connection limit to golang connection pool
  - adding connection limit to just 100 improve write performance to 1k RPS range
  - reading performance on the other hands tanked to around 200s RPS, which is pretty unexpected

- now lets do breakpoint test
  - `docker compose exec tester /app/k6 run tests/k6-breakpoint.js`
  - the result would be much smaller because of the efficiency of both tools (wrk vs k6). how they count it would also be much different
  - k6 need much larger memory compared to wrk
  - disabling the check yield higher request
  - somehow the latency of the app is higher compared to measured with wrk, thus it would be detected as slow request
    - consequently its hard to push the cpu usage of our web API using this approach
  - we could export the result. there are:
    - summary
    - detailed json. could be huge. 2 minutes breakpoint test yield 500MB file.

  // below need more elaborate checks. req/s is different compared to text summary. its consume way to much memory. loading the data also very costly.
  - the tester container already included `k6-dashboard` package, that give us beatiful dashboard for reaading k6 load testing data
  - `docker compose exec tester /app/k6 run --out web-dashboard tests/k6-breakpoint.js`
    - we could monitor in realtime at localhost:5665
    - to generate report automatically:
    - `docker compose exec tester /app/k6 run --out "web-dashboard=export=test-report-$(date '+%Y%m%d-%H%M%S').html" tests/k6-breakpoint.js`




