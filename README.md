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

## Performance Benchmarks

- JSON marshaling: ~1.3μs/op (small payload)
- JSON marshaling with indentation: ~2.3μs/op
- Token generation: ~350ns/op
- Password hashing (bcrypt): ~270ms/op