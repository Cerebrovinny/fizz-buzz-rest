# FizzBuzz REST API

FizzBuzz sequences with request statistics and health checks.

## Highlights

- Custom FizzBuzz output by supplying your own divisors, strings, and upper limit
- Built-in statistics that surface the most popular parameter combination
- Structured logging, graceful shutdown, CORS controls, and health endpoints ready for production

## Quick Start

### Docker Compose (recommended)

```bash
docker-compose up -d        # start
docker-compose logs -f      # stream logs
docker-compose down         # stop
```

Service listens on http://localhost:8080.

### Docker

```bash
docker build -t fizzbuzz-api .
docker run -d -p 8080:8080 --name fizzbuzz fizzbuzz-api
```

### Local Go

```bash
go run ./cmd/server          # run directly
# or
make build && ./bin/server   # build and run
```

Run `make help` to see all automation targets (fmt, lint, test, docker, etc.).

## API

| Method | Path          | Description                                     |
| ------ | ------------- | ----------------------------------------------- |
| GET    | `/fizzbuzz`   | Generate a sequence with custom parameters      |
| GET    | `/statistics` | Return the most frequently requested parameters |
| GET    | `/health`     | Liveness/readiness probe                        |

### FizzBuzz

- Required query parameters: `int1`, `int2`, `limit`, `str1`, `str2`
- All numeric values must be greater than 0; strings must be non-empty

```bash
curl "http://localhost:8080/fizzbuzz?int1=3&int2=5&limit=15&str1=fizz&str2=buzz"
```

```json
{
  "result": [
    "1",
    "2",
    "fizz",
    "4",
    "buzz",
    "fizz",
    "7",
    "8",
    "fizz",
    "buzz",
    "11",
    "fizz",
    "13",
    "14",
    "fizzbuzz"
  ]
}
```

### Statistics

Returns the parameter set with the highest request count (tracked in-memory).

```bash
curl http://localhost:8080/statistics
```

```json
{
  "params": {
    "int1": 3,
    "int2": 5,
    "limit": 15,
    "str1": "fizz",
    "str2": "buzz"
  },
  "hits": 42
}
```

### Health

```bash
curl http://localhost:8080/health
```

Returns `200 OK` when the service is ready to receive traffic.

## Configuration

| Variable               | Default | Purpose                                      |
| ---------------------- | ------- | -------------------------------------------- |
| `PORT`                 | `8080`  | HTTP listener port                           |
| `LOG_LEVEL`            | `info`  | `debug`, `info`, `warn`, or `error`          |
| `LOG_FORMAT`           | `json`  | `json` for production, `text` for local runs |
| `READ_TIMEOUT`         | `15s`   | Server read timeout                          |
| `WRITE_TIMEOUT`        | `15s`   | Server write timeout                         |
| `CORS_ALLOWED_ORIGINS` | `*`     | Comma-separated list of allowed origins      |

Override variables in your shell, `.env`, or `docker-compose.yml` as needed.

## Development

- `go test ./...` (or `make test`) to run the test suite
- `go test -race ./...` to run with the race detector
- `make lint` and `make fmt` to enforce style and linting
- Run benchmarks with `go test -bench=. -benchmem ./internal/fizzbuzz`
