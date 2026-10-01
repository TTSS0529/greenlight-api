# Greenlight API

A RESTful movie catalog API built with Go and PostgreSQL.

The project focuses on practical backend engineering: HTTP API design, PostgreSQL integration, authentication and authorization, validation, concurrency, rate limiting, graceful shutdown, testing, containerized deployment, and CI.

## Features

* RESTful JSON API
* Movie CRUD operations
* Full-text movie search with PostgreSQL
* Filtering, sorting, and pagination
* Favorite movies
* User registration and account activation
* Stateful Bearer token authentication
* Permission-based authorization
* Password hashing with bcrypt
* Optimistic locking for concurrent updates
* Per-IP token-bucket rate limiting
* Structured JSON logging and panic recovery
* CORS support
* Graceful shutdown and background task management
* Application metrics with `expvar`
* PostgreSQL schema migrations
* Docker-based local deployment
* Unit and PostgreSQL integration tests
* GitHub Actions CI
* OpenAPI 3.1 API specification

## Getting Started

### Prerequisites

For Docker-based deployment:

* Docker
* Docker Compose

For running the application directly on the host:

* Go
* PostgreSQL
* `migrate` CLI

### Run with Docker

The easiest way to run the complete application stack is with Docker Compose.

The Docker deployment includes:

```text
Caddy
  │
  ▼
Go API
  │
  ▼
PostgreSQL
```

Create a `.env` file(for docker), for example:

```bash
POSTGRES_PASSWORD=pa55word
```

Build the images, start PostgreSQL, run migrations and seed data, and start the API:

```bash
make docker/deploy
```

The application is then available through Caddy.

To view the service logs:

```bash
make docker/logs
```

To stop the services:

```bash
make docker/down
```

To stop the services and remove the PostgreSQL volume:

```bash
make docker/down/v
```

To completely rebuild the environment from scratch:

```bash
make docker/reset
```

Other useful Docker commands:

```bash
make docker/build
make docker/up
make docker/redeploy
```

### Run locally with PostgreSQL

The application can also be run directly on the host without containerizing the Go API.

First, install and start PostgreSQL, then create a `.envrc` file containing the required database connection settings, for example:

```bash
export GREENLIGHT_DB_DSN='postgres://greenlight:pa55word@localhost/greenlight'
```

Apply the database migrations:

```bash
make db/migrations/up
```

Then start the API:

```bash
make run/api
```

The application will start the Go HTTP server directly on the host.

You can connect to the database using:

```bash
make db/psql
```

This development mode is useful when working on the Go application because changes can be run directly with `go run` without rebuilding the Docker image.

### Testing

Run the integration test environment:

```bash
make test
```

Run the unit test suite with the race detector:

```bash
go test -race -vet=off ./...
```

Run benchmarks:

```bash
make benchmark
```

Run the project's complete quality checks:

```bash
make audit
```

This runs formatting, `go vet`, `staticcheck`, and the test suite.

### Useful Make Commands

The available Make targets can be listed with:

```bash
make help
```

Common commands include:

| Command                 | Description                                  |
| ----------------------- | -------------------------------------------- |
| `make run/api`          | Run the Go API locally                       |
| `make db/psql`          | Connect to PostgreSQL                        |
| `make db/migrations/up` | Apply database migrations                    |
| `make test`             | Run PostgreSQL integration tests             |
| `make audit`            | Run formatting, static analysis, and tests   |
| `make benchmark`        | Run benchmarks                               |
| `make build/api`        | Build the API binary                         |
| `make docker/deploy`    | Build and start the Docker environment       |
| `make docker/down`      | Stop Docker services                         |
| `make docker/down/v`    | Stop services and remove volumes             |
| `make docker/redeploy`  | Rebuild and redeploy                         |
| `make docker/reset`     | Completely reset and rebuild the environment |
| `make docker/logs`      | Follow Docker service logs                   |

## API

| Method | Endpoint                    | Description                    | Auth           |
| ------ | --------------------------- | ------------------------------ | -------------- |
| GET    | `/v1/healthcheck`           | Health and version information | —              |
| GET    | `/v1/movies`                | List and search movies         | `movies:read`  |
| POST   | `/v1/movies`                | Create a movie                 | `movies:write` |
| GET    | `/v1/movies/:id`            | Get a movie                    | `movies:read`  |
| PATCH  | `/v1/movies/:id`            | Partially update a movie       | `movies:write` |
| DELETE | `/v1/movies/:id`            | Delete a movie                 | `movies:write` |
| POST   | `/v1/movies/:id/favorite`   | Add a favorite                 | `movies:read`  |
| DELETE | `/v1/movies/:id/favorite`   | Remove a favorite              | `movies:read`  |
| GET    | `/v1/favorites`             | List favorite movies           | `movies:read`  |
| POST   | `/v1/users`                 | Register a user                | —              |
| PUT    | `/v1/users/activated`       | Activate a user                | —              |
| POST   | `/v1/tokens/authentication` | Create an authentication token | —              |
| GET    | `/debug/vars`               | Application metrics            | —              |

Detailed API schemas and request/response definitions are available in [`openapi.yaml`](./openapi.yaml).

## Architecture

```text
HTTP Request
     │
     ▼
Middleware
 ├── Rate Limiting
 ├── Authentication
 ├── CORS
 ├── Panic Recovery
 └── Metrics
     │
     ▼
HTTP Handlers
     │
     ▼
Data Models
     │
     ▼
PostgreSQL
```

The application is organized into:

```text
cmd/api/          HTTP handlers, middleware, routing, server setup
internal/data/    Database models and queries
internal/validator/
                  Request validation
internal/jsonlog/ Structured logging
internal/mailer/  Email delivery and templates
migrations/       PostgreSQL schema migrations
```

## Key Engineering Decisions

### PATCH and optimistic locking

`PATCH` implements partial updates at the API level. The existing movie is loaded, supplied fields are merged, and the complete record is validated before updating the database.

Concurrent updates are protected with optimistic locking:

```sql
UPDATE movies
SET ..., version = version + 1
WHERE id = $1 AND version = $2
```

If no row is updated, the API returns `409 Conflict`.

### PostgreSQL full-text search

Movie title search uses PostgreSQL full-text search with `tsvector`, `plainto_tsquery`, and a GIN index.

Filtering, sorting, and pagination are handled at the database level using `WHERE`, `ORDER BY`, `LIMIT`, and `OFFSET`.

### Authentication and authorization

Authentication uses stateful Bearer tokens.

```text
POST /v1/tokens/authentication
        │
        ▼
   Authenticate user
        │
        ▼
  Generate random token
        │
        ▼
 Store SHA-256 hash in DB
        │
        ▼
 Return plaintext token
```

Protected endpoints authenticate the token against the database and then check user permissions such as:

* `movies:read`
* `movies:write`

Passwords are hashed with bcrypt and are never stored in plaintext.

### Rate limiting

The API uses a token-bucket rate limiter from `golang.org/x/time/rate`.

Each client IP has its own limiter:

* configurable requests-per-second rate
* configurable burst capacity
* inactive clients removed periodically
* `429 Too Many Requests` when the bucket is exhausted

The current implementation is intentionally in-memory and therefore per-server-instance. A distributed deployment would require shared rate-limiting state, such as Redis.

### Graceful shutdown

The server handles `SIGINT` and `SIGTERM` and allows in-flight requests and background tasks to finish before shutdown.

Background email delivery is tracked with `sync.WaitGroup`.

### Error handling

The API uses consistent JSON error responses:

```json
{
  "error": "the requested resource could not be found"
}
```

Validation errors return field-specific messages:

```json
{
  "error": {
    "title": "must be provided",
    "year": "must not be in the future"
  }
}
```

The API uses appropriate HTTP status codes including `400`, `401`, `403`, `404`, `409`, `422`, `429`, and `500`.

## Testing

The project contains unit tests and PostgreSQL integration tests.

Integration tests run against a dedicated PostgreSQL Docker environment with migrations applied.

Because the integration tests share database state, package-level test parallelism is disabled:

```bash
go test -p 1 -race -tags=integration ./...
```

The test suite covers:

* HTTP handlers and middleware
* validation
* database models and SQL queries
* authentication and tokens
* permissions
* PostgreSQL integration
* race detection

For a larger production system, stronger test isolation could be introduced using separate databases, schemas, or transaction-based rollback.

## CI

GitHub Actions runs:

* `gofmt`
* `go vet`
* `staticcheck`
* unit tests
* race detector
* PostgreSQL integration tests

## Deployment Architecture

The application can be run locally with Docker Compose.

The local stack contains:

```text
Caddy
  │
  ▼
Go API
  │
  ▼
PostgreSQL
```

Caddy acts as a reverse proxy in front of the Go application.

In production, Caddy can also terminate TLS and forward HTTP traffic to the Go application over the internal network.

## Performance Benchmarks

Selected local microbenchmarks:

| Operation                        | Approx. Result |
| -------------------------------- | -------------: |
| JSON marshaling                  |     ~1.3 μs/op |
| JSON marshaling with indentation |     ~2.3 μs/op |
| Token generation                 |     ~350 ns/op |
| bcrypt password hashing          |     ~270 ms/op |

These measurements are environment-dependent and are included as development benchmarks rather than production performance guarantees.

## Project Structure

```text
.
├── cmd/
│   └── api/
│       ├── routes.go
│       ├── middleware.go
│       ├── movies.go
│       ├── favorites.go
│       ├── user.go
│       ├── tokens.go
│       └── ...
├── internal/
│   ├── data/
│   ├── jsonlog/
│   ├── mailer/
│   └── validator/
├── migrations/
├── openapi.yaml
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── README.md
```

## What I Learned

This project was built to understand how a production-oriented Go backend is structured rather than only how to make HTTP handlers work.

Key areas include:

* Go HTTP server and middleware design
* PostgreSQL and SQL migrations
* Database connection pooling and context timeouts
* REST API design
* Authentication and authorization
* Concurrent programming and graceful shutdown
* Rate limiting
* Structured logging and error handling
* Automated testing and race detection
* Docker-based deployment
* CI and static analysis
* OpenAPI API documentation
