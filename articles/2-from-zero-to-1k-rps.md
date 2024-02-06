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
- are we done? lets check for DB-related APIs
- now lets do breakpoint test


