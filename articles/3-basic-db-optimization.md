- in my experience, 80% of case the bottleneck of app is not in the app itself, instead in the other service (database query, 3rd party APIs, etc)
- we'll try basic DB optimizations that wouldn't change many app logic or db structures first:

1. add docker setting for psql config file
- confirm by running `docker compose exec postgres psql -U postgres -c 'show max_connections;'`
2. add number of psql connection
3. analyze difference
4. add indexing
5. analyze
6. more db tuning
7. analyze
8. conclusions
