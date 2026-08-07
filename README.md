# greenlight-api
A RESTful movie catalog API built with Go, PostgreSQL, authentication, rate limiting and background processing.

## Testing Strategy

This project uses integration tests for database-related functionality.
The tests run against a dedicated PostgreSQL container with migrations applied before execution.

I chose integration tests for the data layer because they verify:
- SQL queries
- database schema compatibility
- migrations
- constraints

In a larger production system, I would combine integration tests with unit tests using mocks for isolated business logic.

## Integration Test Isolation

The project uses PostgreSQL integration tests running in a dedicated Docker test environment.

Integration tests share a test database and reset database state with `TRUNCATE` before each test. During development, intermittent failures were discovered due to Go's default package-level test parallelism.

Different Go packages could run tests concurrently while accessing the same PostgreSQL database. For example, one package could truncate tables while another package was inserting records, causing inconsistent database states and errors such as foreign key constraint violations.

The issue was caused by shared database state between concurrently running integration tests, rather than application logic.

To ensure deterministic test execution, integration tests are currently executed with:

```bash
go test -p 1 -race -tags=integration ./...
```

The -p 1 option limits Go test execution to one package at a time, preventing different packages from modifying the same test database concurrently.

For a larger-scale project, stronger isolation strategies could be introduced, such as:
- separate databases for different test packages
- isolated PostgreSQL schemas per test
- transaction-based rollback

The current approach prioritizes reliability and simplicity because the integration test suite is still small and the execution time remains acceptable.

## Performance Benchmarks

- JSON marshaling: ~1.3μs/op (small payload)
- JSON marshaling with indentation: ~2.3μs/op
- Token generation: ~350ns/op
- Password hashing (bcrypt): ~270ms/op

## Continuous Integration

Every push runs:

- gofmt
- go vet
- staticcheck
- unit tests
- race detector
- PostgreSQL integration tests