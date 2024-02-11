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
  - explain result: now it uses the index
  - pprof results: db query for read now taking only 23% or requests time
6. more db tuning
  ```
  synchronous_commit = off
  ```
7. analyze
  - now the write almost reached ~30k RPS !
  - read doesn't affected much
  - other tuning params you could try:
  ```
  shared_buffers = 1GB
  wal_writer_delay = 1s
  wal_buffers = 15MB
  ```
  - coba profiling write, sekarang write sekitar 16%
8. prepared statement
  - https://go.dev/doc/database/prepared-statements
  - hasil profiling sebelumnya, ada step 'preparedTo' yang makan 17% waktu eksekusi juga. coba kita optimize ini
9. Analyze
  - read sekarang sampai 55k RPS!
  - write sampai 53k RPS
  - karena ini optimasi di kode, gak ada perubahan dari segi EXPLAIN ANALYZE
  - hasil profiling udah gak ada pemanggilan step prepare lagi
10. conclusions
  - perhatikan spec
  - cek ukuran db juga. di aku ketika jumlah data sektiar 3jt an, write mulai turun
