# greenlight-api
A RESTful movie catalog API built with Go, PostgreSQL, JWT authentication, rate limiting and background processing.

## Testing Strategy

This project uses integration tests for database-related functionality.
The tests run against a dedicated PostgreSQL container with migrations applied before execution.

I chose integration tests for the data layer because they verify:
- SQL queries
- database schema compatibility
- migrations
- constraints

In a larger production system, I would combine integration tests with unit tests using mocks for isolated business logic.
