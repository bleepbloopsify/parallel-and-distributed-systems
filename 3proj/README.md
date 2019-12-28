# Name

Leon Chou


# Building

## Backend

`$ cd backend && go build .`

## Frontend

`$ cd frontend && go build .`

## Docker Compose

`$ docker-compose build`

# Running


## Backend

`$ cd backend && ./backend -listen=<listenAddr>`

## Frontend

`$ cd frontend && ./frontend -listen=<listenAddr> -backend=<backendAddr>`

## Docker Compose

`$ docker-compose up`

# Testing

If building and running each individual item, you can specify which ports to hit.

## Vegeta
The benchmark.cfg file is configured to be used with docker-compose port configuration.

A sample report has been generated using this command:

`$ vegeta attack -body body.txt -targets benchmark.cfg -output report.bin --workers=50 -max-workers=50 -duration=60s -rate=0`

This expects you to have spun up the server using docker-compose, because of the port mappings.

We don't expect users to create many extra geese, so that isn't reflected here. Otherwise Vegeta shows a pretty straightforward report

`$ vegeta report report.bin`

```
Requests      [total, rate, throughput]  45827, 763.78, 763.10
Duration      [total, attack, wait]      1m0.053955974s, 59.99993536s, 54.020614ms
Latencies     [mean, 50, 95, 99, max]    65.313442ms, 62.855752ms, 121.371383ms, 150.359961ms, 271.757197ms
Bytes In      [total, mean]              136640610, 2981.66
Bytes Out     [total, mean]              733232, 16.00
Success       [ratio]                    100.00%
Status Codes  [code:count]               200:45827
Error Set:
```

## Browser
The docker-compose configuration is configured to spin up two frontend servers and a single backend server.
You can test using vegeta.


# Decisions

I have mutexes on each goose and a mutex across the store for adding geese. Removal of geese is easy as long as I'm holding the geese's mutex because the ID is autoincrement transient.

Performance was measured in p two nines, so we see a 150.359961ms latency across the 99th percentile, which is fast enough for most users. Our highest latency is also lower than 300ms, which is still smaller than a third of a second.

Failure detection is implemented with a goroutine loop and a specific request on the backend. This allows us to make sure the backend is working.

Timeout is once every 5 seconds.
