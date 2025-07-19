#!/bin/bash

# Script to run integration tests with Docker PostgreSQL

set -e

echo "Starting PostgreSQL test database..."
docker-compose -f docker-compose.test.yml up -d postgres-test

echo "Waiting for PostgreSQL to be ready..."
sleep 10

# Wait for PostgreSQL to be healthy
until docker-compose -f docker-compose.test.yml exec -T postgres-test pg_isready -U postgres -d todoapi_test; do
  echo "Waiting for PostgreSQL..."
  sleep 2
done

echo "PostgreSQL is ready!"

# Set test database URL
export TEST_DATABASE_URL="postgres://postgres:password@localhost:5433/todoapi_test?sslmode=disable"

echo "Running integration tests..."
go test -v ./tests/integration/

# Cleanup
echo "Stopping test database..."
docker-compose -f docker-compose.test.yml down

echo "Integration tests completed!"