- in my experience, 80% of case the bottleneck of app is not in the app itself, instead in the other service (database query, 3rd party APIs, etc)
- we'll try basic DB optimizations that wouldn't change many app logic or db structures first:

1. add docker setting for psql config file
2. add number of psql connection
  - confirm by running `docker compose exec postgres psql -U postgres -c 'show max_connections;'`
3. analyze difference
  - increasing number of connection increase the number of requests for write operations around 3x (from ~1k to 3k)
  - interestingly the read operations doesn't seems to be affected. If anything, it seemed to be decreasing
  - pprof results: db query for read is taking 40% or requests time
4. add indexing
  - basic intro to 'explain' query
5. analyze
  - write doesn't seems to affected much
  - read however, now we got ~35k RPS!!!
  - pprof results: db query for read now taking only 23% or requests time
6. more db tuning
7. analyze
8. conclusions
