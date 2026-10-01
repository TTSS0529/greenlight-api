## chapter 1 Introduction

Endpoints as follows:
| Method | URL Pattern                 | Action                                          |
| ------ | --------------------------- | ----------------------------------------------- |
| GET    | `/v1/healthcheck`           | Show application health and version information |
| GET    | `/v1/movies`                | Show the details of all movies                  |
| POST   | `/v1/movies`                | Create a new movie                              |
| GET    | `/v1/movies/:id`            | Show the details of a specific movie            |
| PATCH  | `/v1/movies/:id`            | Update the details of a specific movie          |
| DELETE | `/v1/movies/:id`            | Delete a specific movie                         |
| POST   | `/v1/movies/:id/favorite`   | Add a movie to the user's favorites             |
| DELETE | `/v1/movies/:id/favorite`   | Remove a movie from the user's favorites        |
| GET    | `/v1/favorites`             | Show the user's favorite movies                 |
| POST   | `/v1/users`                 | Register a new user                             |
| PUT    | `/v1/users/activated`       | Activate a specific user                        |
| POST   | `/v1/tokens/authentication` | Generate a new authentication token             |
| GET    | `/debug/vars`               | Display application metrics                     |

## chapter 2 Getting Started

### Go Modules: Module Path

Initialize a module with:
```bash
go mod init github.com/username/project
```
- The argument is the module path, which identifies the Go module.
- It is usually the GitHub repository path for projects hosted on GitHub.
- The module path is also used as the base for importing packages:
    `import "github.com/username/project/internal/data"`

**Key point**: The module path is an import path, not the URL where the application is deployed.

### `internal` Packages

`internal` is a special directory recognized by the Go toolchain.

Packages under `internal` can only be imported by code within the
parent directory tree of `internal`.

This is a **language/toolchain rule**, not merely a convention.

`pkg` and `cmd`, on the other hand, are conventions and have no
special import restrictions.

### `http.ServeMux` Routing Enhancements

Since Go 1.22, `http.ServeMux` supports:

- HTTP method matching
- Path wildcards: `/movies/{id}`
- Path parameters via `r.PathValue("id")`
- Automatic `405 Method Not Allowed` responses

Therefore, for many REST APIs, the standard library can replace a third-party router such as `httprouter`.

> `httprouter` is still useful when advanced routing features or specific performance characteristics are required.

### Concepts

#### Idempotent

An operation is **idempotent** if performing it multiple times has the same effect on the final state as performing it once.

> Repeating the same operation does not cause additional state changes.

**HTTP examples:**
- `GET` → idempotent
- `PUT` → idempotent
- `DELETE` → idempotent
- `POST` → generally not idempotent

**Key point:** Idempotency concerns the **final state**, not necessarily identical response messages.

## chapter 3 Sending JSON Responses

### `json.MarshalIndent()` vs `json.Marshal()`

- `json.MarshalIndent()` produces more human-readable, formatted JSON.
- **Trade-off:** better readability, but slightly worse performance due to additional formatting work and larger output.
- Use `json.Marshal()` when performance and compact output matter; use `json.MarshalIndent()` when readability is more important.

## chapter 6 SQL Migrations

### Why Use SQL Migrations?

- **Version control** — Track database schema changes alongside source code.
- **Reproducibility** — Recreate the same schema in any environment.
- **Incremental changes** — Apply schema changes step by step.
- **Consistency** — Keep development, testing, and production schemas in sync.
- **Rollback** — Revert schema changes when necessary.

> SQL migrations make database schema changes versioned, reproducible, and consistent.

## chapter 7 CRUD Operations

### mocking models

In internal/data/models.go, everything is contained in:
```
type Models struct {
	Movies      MovieModel
	Permissions PermissionModel
	Tokens      TokenModel
	Users       UserModel
	Favorites   FavoriteModel
}
```

For mocking models, we can replace specific models with interface.
Using Movies as example:
```
type Models struct {
    Movies interface {
        Insert(movie *Movie) error
        Get(id int64) (*Movie, error)
        Update(movie *Movie) error
        Delete(id int64) error
    }
}
```

With helper function:
```
func NewMockModels() Models {
    return Models{
        Movies: MockMovieModel{},
    }
}
```

## chapter 8 Advanced CRUD Operations

### PATCH vs. Database UPDATE

- `PATCH` means **partial update at the API level**: only provided fields are changed.
- The database can still **UPDATE all columns** after merging the changes in Go.

```text
PATCH → GET full resource → modify provided fields → UPDATE all columns
```

- **True database-level partial update** can be implemented by dynamically building the `UPDATE` statement based on the fields provided:

```text
PATCH → build dynamic SQL → UPDATE only changed columns
```

- This avoids updating unchanged columns but adds complexity.

**Key point:** PATCH semantics do not require SQL to update only the changed columns.

### Optimistic Locking

Prevent concurrent updates from overwriting each other by using a `version` column:

- Read the current `version`.
- Update only if the version is unchanged.
- Increment `version` on a successful update.
- If no row is returned, another request has already modified the movie → **edit conflict**.

This is called **optimistic locking:** assume conflicts are rare and detect them when updating, rather than locking the row beforehand.

### context timeout

**context timeout** applies to the whole DB operation, including waiting for a connection from sql.DB's pool.

So `QueryRowContext()` may return `context.DeadlineExceeded` even before the SQL query
starts executing in PostgreSQL.

## chapter 9 Filtering, Sorting and Pagination

### search flow

```
text
  ↓
to_tsvector()
  ↓
tsvector + GIN index
  ↓
to_tsquery()
  ↓
@@
  ↓
matching rows
  ↓
pagination(LIMIT and OFFSET clauses)
```

### Pagination: `COUNT(*) OVER()` edge case

> `COUNT(*) OVER()` cannot provide `totalRecords` when
> `LIMIT/OFFSET` produces zero rows, because there is no row
> to scan the count from.

For a production API, a separate `COUNT(*)` query can be used
if metadata is required even for an empty page.

## chapter 10 Structured Logging and Error Handling

### Logging

We implement our own custom logger.
If you don't want to do this, you can use the third-prty packages, such as `zerolog`.

### Panic Recovery

- `recoverPanic()` catches panics in the **handler goroutine** and prevents the server from crashing.
- `recover()` only works within the **same goroutine**.
- Panics in goroutines spawned by a handler must be recovered **inside those goroutines**.

## chapter 11 Rate Limiting

### Token-Bucket Rate Limiter

A rate limiter that controls request frequency using a **bucket of tokens**.

- Tokens are added at a fixed rate.
- Each request consumes one token.
- If no token is available, the request is rejected and return `429 Too Many Requests`.
- The bucket has a maximum capacity(full at beginning), allowing short **bursts** of requests.

**Key idea:** steady request rate + limited bursts.

### Token-Bucket Rate Limiting Middleware

- Uses `golang.org/x/time/rate` to implement a **token-bucket rate limiter**.
- Maintains a separate limiter for each client IP:
  - `rps`: token refill rate.
  - `burst`: maximum burst size.
- Each request:
  1. Gets the client's IP.
  2. Creates a limiter if the IP is new.
  3. Updates `lastSeen`.
  4. Calls `limiter.Allow()`.
  5. If no token is available → return `429 Too Many Requests`.
  6. Otherwise → pass the request to `next`.

### Client Cleanup

A background goroutine runs every minute and removes clients
that have been inactive for more than 3 minutes.

### Pros

- Simple and easy to implement.
- Per-IP limiting isolates clients from each other.
- Token bucket allows controlled bursts.
- Inactive clients are removed to prevent unbounded memory growth.
- `sync.Mutex` protects the shared `clients` map.

### Cons / Limitations

- **In-memory**: limits are per server instance; multiple instances don't share limits.
- **IP-based**: many users behind the same NAT/proxy may share one limiter.
- The cleanup goroutine has a **lifecycle issue**: every middleware instance starts a long-lived goroutine with no shutdown mechanism.
- Holding the mutex while calling `Allow()` is simple but adds lock contention under high concurrency.

### Better Production Design

Extract the logic into a dedicated `rateLimiter` type with:

- `clients map[string]*client`
- `sync.Mutex`
- `Start()` / `Close()` or `context.Context`
- cleanup goroutine with a `time.Ticker`

For multiple API instances, use a **distributed rate limiter** such as Redis instead of an in-memory map.

## chapter 12 Graceful Shutdown

### Signals treatement

- `SIGINT` and `SIGTERM` catched
- `SIGQUIT` left with its default behavior
- `SIGKILL` not catchable

### Why use a buffered channel for signal handling?

- `signal.Notify()` sends OS signals to the channel.
- A buffer of 1 allows the signal to be stored even if the receiver is not ready at exactly the same moment.
- Only one signal is needed because the application exits after receiving the first SIGINT or SIGTERM.

## chapter 13 User Model Setup and Registration

### hash methode

use `golang.org/x/crypto/bcrypt` package to hash user passwords before storing them in the databse

### User Enumeration

**User enumeration**: revealing whether an email/username exists.

#### Risks
- Privacy leakage
- Enables targeted attacks/phishing
- Enables credential stuffing using leaked passwords

#### Mitigation
- Return **ambiguous/identical responses** whether user exists or not.
- Avoid **timing differences** between the two cases.

#### Trade-off
Better security/privacy vs. **more complexity and worse UX**.

**Takeaway:** Prevent enumeration when user privacy or account value justifies the added friction.

## chapter 14 Sending Emails

- use the **Mailtrap** `SMTP service` to send and monitor test emails during
development.
- use the third-party `go-mail/mail(github.com/go-mail/mail/v2@v2.3.0)` package to help send email:
```golang
//go:embed "templates"
var templateFS embed.FS
```
- Use a background goroutine for non-critical tasks like sending emails; track it with `sync.WaitGroup` so graceful shutdown waits for all background tasks to finish.

## chapter 15 User Activation

### Activation Token Hashing

- Activation tokens are 128-bit high-entropy random strings, making them impractical to guess, using `crypto/rand` package.
- Therefore, SHA-256 is sufficient; unlike passwords, they do not need a slow hashing algorithm such as bcrypt.

---

- Activation tokens must only be transmitted over HTTPS in production.
- With Caddy as a reverse proxy, Caddy can terminate TLS and forward requests to the Go API over the internal network.

## 16 Authentication

### HTTP Basic Authentication

**How it works**
- The client sends a username and password in the `Authorization` HTTP header.
- Credentials are encoded using **Base64**.
- The server decodes the credentials and verifies them on every request.

**Advantages**
- **Simple** — easy to implement on both client and server.
- **Widely supported** — built into HTTP clients, browsers, and many tools.
- **Stateless** — the server does not need to maintain a login session.

**Disadvantages**
- **Not encrypted by itself** — Base64 is encoding, not encryption.
- **Credentials are sent with every request**, increasing the impact of interception.
- **Password hashing on every request** — the server must verify the password on every request, and secure password hashing algorithms such as bcrypt are intentionally computationally expensive.
- **Requires HTTPS** in production to protect credentials in transit.
- **Poor user experience** — less flexible than modern authentication mechanisms.
- **Harder to revoke selectively** — changing a password affects all clients using it.

### Token Authentication

The client first authenticates with the server and receives a token. 
The token is then sent with subsequent requests, usually in the `Authorization` header.

#### Stateful Tokens

The server stores the token and its associated user information in a database or cache.

**Advantages**
- **Easy to revoke** — the server can invalidate a token at any time.
- **Easy to manage sessions** — the server has full control over active tokens.
- **Small tokens** — the token itself does not need to contain much information.

**Disadvantages**
- **Requires server-side storage** — tokens must be stored and managed.
- **Database/cache lookup on each request** — adds I/O and latency.
- **More difficult to scale** — shared token storage may be needed across multiple servers.

#### Stateless Tokens

The server does not store the token. The token contains the necessary information and is cryptographically signed so the server can verify it.

**Advantages**
- **No server-side session storage** — easier to scale horizontally.
- **No database lookup required** — the server can verify the token locally.
- **Good performance** — authentication can be handled without external storage.

**Disadvantages**
- **Harder to revoke** — a valid token normally remains valid until it expires.
- **Larger tokens** — more information may need to be stored in the token.
- **Token invalidation is more complex** — revocation often requires additional mechanisms such as short expiration times or token blacklists.

#### Key Difference

> **Stateful:** the server stores and checks the token.
>
> **Stateless:** the server does not store the token and verifies it using the token itself.

### API-key authentication

- like the stateful token approach, but the keys are permanent keys

### OAuth 2.0 / OpenID Connect

## 17 Permission-based Authorization

The required permissions will align with our API endpoints like so:
| Method | URL Pattern      | Required permission |
| ------ | ---------------- | ------------------- |
| GET    | `/v1/movies`     | movies:read         |
| POST   | `/v1/movies`     | movies:write        |
| GET    | `/v1/movies/:id` | movies:read         |
| PATCH  | `/v1/movies/:id` | movies:write        |
| DELETE | `/v1/movies/:id` | movies:write        |

## 18 Cross Origin Requests

HTTP requests made by a web page to a server with a different origin (protocol, domain, or port); browsers restrict these requests by default and require mechanisms such as CORS to allow them.

the same-origin policy is a web browser thing only.

We can leverage the fact that preflight requests always have three components:
the HTTP method OPTIONS, an Origin header, and an Access-Control-Request-Method
header. If any one of these pieces is missing, we know that it is not a preflight request.

## 19 Metrics

use `hey` to generate the load
use `httpsnoop` to get status codes our responses had

## 20 Building, Versioning and Quality Control

`.envrc` file for makefile to use variable `GREENLIGHT_DB_DSN`

`go install honnef.co/go/tools/cmd/staticcheck@latest` makefile uses `staticcheck` to carry out some additional static analytic checks.

### Module Proxies

A **module proxy** is a server that stores and serves Go modules so that the Go toolchain can download them reliably.

By default, Go uses `https://proxy.golang.org`.

#### Benefits

- Faster and more reliable module downloads
- Caches module versions
- Reduces dependency on source repositories
- Provides immutable module versions

You can configure it with the `GOPROXY` environment variable:

```bash
GOPROXY=https://proxy.golang.org,direct
```

## 21 Deployment and Hosting

In the book, the author uses the VPS, but here I deploy locally using docker.

### Caddy

- Caddy sits in front of the Go application as a reverse proxy.
- In production, Caddy can terminate TLS and handle HTTPS.
- Go only needs to run an HTTP server behind Caddy.
- This separates application concerns from TLS/certificate management.
- Local Docker deployment does not need public HTTPS.
- use `realip` package to get the real IP address

## beyond the book

### tests

- unit tests
- integration tests with another docker compose

### CI

- Github Action