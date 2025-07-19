-- Initialize database for Todo API
-- This script runs when the PostgreSQL container starts for the first time

-- Create database if it doesn't exist (handled by POSTGRES_DB env var)
-- The database is automatically created by the postgres image

-- Set timezone
SET timezone = 'UTC';

-- Create extensions if needed
-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- The actual table creation will be handled by GORM migrations
-- This file is mainly for any initial setup that needs to happen
-- before the application starts

-- Log initialization
SELECT 'Todo API database initialized successfully' as message;