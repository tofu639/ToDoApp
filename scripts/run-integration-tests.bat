@echo off
REM Script to run integration tests with Docker PostgreSQL

echo Starting PostgreSQL test database...
docker-compose -f docker-compose.test.yml up -d postgres-test

echo Waiting for PostgreSQL to be ready...
timeout /t 10 /nobreak > nul

REM Wait for PostgreSQL to be healthy
:wait_loop
docker-compose -f docker-compose.test.yml exec -T postgres-test pg_isready -U postgres -d todoapi_test > nul 2>&1
if %errorlevel% neq 0 (
    echo Waiting for PostgreSQL...
    timeout /t 2 /nobreak > nul
    goto wait_loop
)

echo PostgreSQL is ready!

REM Set test database URL
set TEST_DATABASE_URL=postgres://postgres:password@localhost:5433/todoapi_test?sslmode=disable

echo Running integration tests...
go test -v ./tests/integration/

REM Cleanup
echo Stopping test database...
docker-compose -f docker-compose.test.yml down

echo Integration tests completed!
pause