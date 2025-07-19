# Integration Tests

This directory contains comprehensive integration tests for the Todo API backend. These tests verify the complete functionality of the API by testing actual HTTP requests against a real database.

## Overview

The integration tests cover:

- **Complete Authentication Flow**: User registration, login, and JWT token validation
- **Todo CRUD Operations**: Create, read, update, and delete operations with real database interactions
- **User Isolation**: Ensuring users can only access their own todos
- **Input Validation**: Testing various validation scenarios and error handling
- **JWT Authentication**: Token validation, expiration, and malformed token handling
- **Database Interactions**: Foreign key constraints, transactions, and concurrent access
- **API Error Handling**: Malformed requests, missing headers, and edge cases
- **Complete User Workflows**: End-to-end user journeys from registration to todo management

## Test Structure

The integration tests use the `testify/suite` package to provide:

- **Setup/Teardown**: Automatic database setup and cleanup
- **Test Isolation**: Each test runs with a clean state
- **Shared Resources**: Common test data and utilities
- **Organized Test Cases**: Grouped by functionality

## Running Integration Tests

### Prerequisites

1. **Docker**: Required for running the test PostgreSQL database
2. **Go**: Version 1.19 or higher

### Option 1: Using Test Scripts (Recommended)

#### On Linux/macOS:
```bash
chmod +x scripts/run-integration-tests.sh
./scripts/run-integration-tests.sh
```

#### On Windows:
```cmd
scripts\run-integration-tests.bat
```

### Option 2: Manual Setup

1. **Start Test Database**:
   ```bash
   docker-compose -f docker-compose.test.yml up -d postgres-test
   ```

2. **Wait for Database to be Ready**:
   ```bash
   # Wait about 10 seconds for PostgreSQL to start
   docker-compose -f docker-compose.test.yml exec postgres-test pg_isready -U postgres -d todoapi_test
   ```

3. **Set Environment Variable**:
   ```bash
   # Linux/macOS
   export TEST_DATABASE_URL="postgres://postgres:password@localhost:5433/todoapi_test?sslmode=disable"
   
   # Windows (PowerShell)
   $env:TEST_DATABASE_URL="postgres://postgres:password@localhost:5433/todoapi_test?sslmode=disable"
   
   # Windows (CMD)
   set TEST_DATABASE_URL=postgres://postgres:password@localhost:5433/todoapi_test?sslmode=disable
   ```

4. **Run Tests**:
   ```bash
   go test -v ./tests/integration/
   ```

5. **Cleanup**:
   ```bash
   docker-compose -f docker-compose.test.yml down
   ```

### Option 3: Using Existing PostgreSQL

If you have a PostgreSQL instance running, you can use it directly:

```bash
# Create test database
createdb todoapi_test

# Set environment variable
export TEST_DATABASE_URL="postgres://username:password@localhost:5432/todoapi_test?sslmode=disable"

# Run tests
go test -v ./tests/integration/
```

## Test Configuration

### Environment Variables

- `TEST_DATABASE_URL`: PostgreSQL connection string for integration tests
  - If not set, tests will be skipped with an informative message
  - Example: `postgres://postgres:password@localhost:5433/todoapi_test?sslmode=disable`

### Database Requirements

- PostgreSQL 12 or higher
- Empty database (tests will create and clean up tables)
- Connection permissions for creating/dropping tables

## Test Coverage

### Authentication Tests
- ✅ User registration with valid data
- ✅ Registration with duplicate email (409 Conflict)
- ✅ Registration with invalid email format
- ✅ Registration with weak password
- ✅ User login with valid credentials
- ✅ Login with invalid credentials (401 Unauthorized)

### Todo CRUD Tests
- ✅ Create todo with valid data
- ✅ Create todo without authentication (401 Unauthorized)
- ✅ Create todo with invalid data (400 Bad Request)
- ✅ Get all todos for authenticated user
- ✅ Get specific todo by ID
- ✅ Get non-existent todo (404 Not Found)
- ✅ Update todo with valid data
- ✅ Update non-existent todo (404 Not Found)
- ✅ Delete todo
- ✅ Delete non-existent todo (404 Not Found)

### Security Tests
- ✅ User isolation (users cannot access other users' todos)
- ✅ JWT token validation
- ✅ Access without token (401 Unauthorized)
- ✅ Access with invalid token (401 Unauthorized)
- ✅ Access with expired token (401 Unauthorized)
- ✅ Malformed authorization header

### Database Integration Tests
- ✅ Foreign key constraints
- ✅ Transaction rollback behavior
- ✅ Concurrent access handling

### Error Handling Tests
- ✅ Malformed JSON requests
- ✅ Missing Content-Type header
- ✅ Large request bodies

### End-to-End Workflow Tests
- ✅ Complete user journey (register → create todos → list → update → delete)

## Test Data Management

### Test User
- Email: `test@example.com`
- Password: `password123` (bcrypt hashed)
- Created automatically for each test suite

### Database Cleanup
- All test data is cleaned up after each test
- Database tables are recreated for each test suite
- No persistent data between test runs

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   ```
   Error: failed to connect to database
   ```
   - Ensure PostgreSQL is running
   - Check the `TEST_DATABASE_URL` environment variable
   - Verify database credentials and permissions

2. **Tests Skipped**
   ```
   SKIP: TEST_DATABASE_URL not set
   ```
   - Set the `TEST_DATABASE_URL` environment variable
   - Use the provided test scripts for automatic setup

3. **Port Already in Use**
   ```
   Error: port 5433 already in use
   ```
   - Stop any existing PostgreSQL containers
   - Change the port in `docker-compose.test.yml`
   - Update the `TEST_DATABASE_URL` accordingly

4. **Permission Denied**
   ```
   Error: permission denied for database
   ```
   - Ensure the database user has CREATE/DROP permissions
   - Use a superuser account for testing

### Debug Mode

To run tests with more verbose output:

```bash
go test -v -race ./tests/integration/
```

To run specific test cases:

```bash
go test -v ./tests/integration/ -run TestAuthenticationFlow
```

## Contributing

When adding new integration tests:

1. Follow the existing test structure and naming conventions
2. Ensure tests are isolated and don't depend on each other
3. Clean up any test data created
4. Add appropriate assertions and error checking
5. Update this README if adding new test categories

## Performance Considerations

- Tests use a separate database to avoid conflicts
- Database operations are optimized for test speed
- Connection pooling is configured for test workloads
- Tests run in parallel where possible

## Security Notes

- Test database uses default credentials (not for production)
- JWT tokens use test secrets (not for production)
- All test data is ephemeral and cleaned up
- No sensitive data should be used in tests