# Todo API Backend

A comprehensive Todo API backend built with Go and the Gin web framework. This project demonstrates modern Go development practices, clean architecture, JWT authentication, and containerized deployment.

## Features

- **User Authentication**: Secure registration and login with JWT tokens
- **Todo Management**: Full CRUD operations for todo items
- **Clean Architecture**: Layered architecture with clear separation of concerns
- **Database Integration**: PostgreSQL with GORM ORM
- **API Documentation**: Interactive Swagger/OpenAPI documentation
- **Containerization**: Docker and Docker Compose support
- **Comprehensive Testing**: Unit and integration tests
- **Security**: Password hashing, JWT validation, input validation

## Tech Stack

- **Language**: Go 1.23+
- **Web Framework**: Gin
- **Database**: PostgreSQL
- **ORM**: GORM
- **Authentication**: JWT (golang-jwt/jwt)
- **Validation**: go-playground/validator
- **Documentation**: Swagger/OpenAPI (swaggo)
- **Testing**: testify
- **Containerization**: Docker & Docker Compose

## Quick Start

### Prerequisites

- Go 1.23 or higher
- PostgreSQL 13+ (or use Docker Compose)
- Docker and Docker Compose (optional, for containerized setup)

### Option 1: Docker Compose (Recommended)

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd todo-api-backend
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env file with your configuration
   ```

3. **Start the application**
   ```bash
   docker-compose up -d
   ```

4. **Verify the setup**
   ```bash
   curl http://localhost:8080/health
   ```

The API will be available at `http://localhost:8080` and Swagger documentation at `http://localhost:8080/swagger/index.html`.

### Option 2: Local Development

1. **Install dependencies**
   ```bash
   go mod download
   ```

2. **Set up PostgreSQL database**
   ```bash
   # Create database
   createdb todoapi
   
   # Run migrations
   psql -d todoapi -f scripts/init-db.sql
   ```

3. **Configure environment**
   ```bash
   cp .env.example .env
   # Edit DATABASE_URL and other settings in .env
   ```

4. **Run the application**
   ```bash
   go run cmd/server/main.go
   ```

## Configuration

The application uses environment variables for configuration. Copy `.env.example` to `.env` and modify as needed:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `ENVIRONMENT` | Environment (development/production) | `development` |
| `LOG_LEVEL` | Logging level (debug/info/warn/error) | `info` |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://user:password@localhost:5432/todoapi?sslmode=disable` |
| `JWT_SECRET` | JWT signing secret (min 32 chars) | `your-super-secret-jwt-key-change-this-in-production` |
| `JWT_EXPIRATION` | JWT token expiration (hours) | `24` |
| `ALLOWED_ORIGINS` | CORS allowed origins | `*` |

**⚠️ Security Note**: Always use strong, unique values for `JWT_SECRET` in production.

## API Documentation

### Interactive Documentation
Visit `http://localhost:8080/swagger/index.html` for interactive API documentation.

### Authentication Endpoints

#### Register User
```bash
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

#### Login User
```bash
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "user@example.com"
  }
}
```

### Todo Endpoints

All todo endpoints require authentication. Include the JWT token in the Authorization header:
```
Authorization: Bearer <your-jwt-token>
```

#### Create Todo
```bash
POST /api/v1/todos
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Complete project documentation",
  "description": "Write comprehensive README and API docs"
}
```

#### Get All Todos
```bash
GET /api/v1/todos
Authorization: Bearer <token>
```

#### Get Todo by ID
```bash
GET /api/v1/todos/{id}
Authorization: Bearer <token>
```

#### Update Todo
```bash
PUT /api/v1/todos/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Updated title",
  "description": "Updated description",
  "completed": true
}
```

#### Delete Todo
```bash
DELETE /api/v1/todos/{id}
Authorization: Bearer <token>
```

### Health Check
```bash
GET /health
```

## Development

### Project Structure
```
todo-api-backend/
├── cmd/server/           # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── handler/         # HTTP handlers
│   ├── middleware/      # HTTP middleware
│   ├── model/          # Data models
│   ├── repository/     # Data access layer
│   ├── service/        # Business logic
│   └── database/       # Database connection
├── pkg/                # Reusable packages
│   ├── jwt/           # JWT utilities
│   ├── password/      # Password hashing
│   └── validator/     # Input validation
├── docs/              # API documentation
├── scripts/           # Database scripts
├── tests/             # Test files
└── docker-compose.yml # Container orchestration
```

### Building the Application

#### Local Build
```bash
go build -o bin/server cmd/server/main.go
./bin/server
```

#### Docker Build
```bash
docker build -t todo-api-backend .
docker run -p 8080:8080 --env-file .env todo-api-backend
```

### Database Management

#### Run Migrations
```bash
# Using psql
psql -d todoapi -f scripts/init-db.sql

# Or using Docker
docker-compose exec postgres psql -U user -d todoapi -f /docker-entrypoint-initdb.d/init-db.sql
```

#### Database Schema
The application uses two main tables:
- `users`: User accounts with email and hashed passwords
- `todos`: Todo items linked to users with foreign key relationship

## Testing

### Running Tests

#### Unit Tests
```bash
# Run all unit tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### Integration Tests
```bash
# Run integration tests (requires test database)
go test ./tests/integration/...

# Using Docker for integration tests
docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit
```

### Test Structure
- `tests/handler/`: HTTP handler unit tests
- `tests/service/`: Business logic unit tests
- `tests/repository/`: Data layer unit tests
- `tests/middleware/`: Middleware unit tests
- `tests/integration/`: End-to-end integration tests

## Deployment

### Docker Deployment

#### Production Docker Compose
```bash
# Set production environment variables
export JWT_SECRET="your-production-jwt-secret-at-least-32-characters"
export ALLOWED_ORIGINS="https://yourdomain.com"

# Deploy
docker-compose up -d
```

#### Environment-Specific Configurations
- **Development**: Use `.env` file with debug logging
- **Production**: Use environment variables with info/warn logging
- **Testing**: Use separate test database and configuration

### Health Checks
The application provides health check endpoints:
- `/health`: Basic health check
- `/ready`: Readiness check (includes database connectivity)

### Monitoring
- Structured JSON logging in production
- HTTP request logging with Gin middleware
- Database connection health monitoring

## Security Considerations

### Authentication & Authorization
- JWT tokens with configurable expiration
- Secure password hashing using bcrypt (cost factor: 12)
- User context isolation (users can only access their own todos)

### Input Validation
- Comprehensive input validation using struct tags
- SQL injection prevention through GORM parameterized queries
- XSS prevention through proper JSON encoding

### Production Security
- Use strong JWT secrets (minimum 32 characters)
- Configure CORS for specific origins
- Use HTTPS in production
- Regular security updates for dependencies

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines
- Follow Go best practices and idioms
- Write tests for new functionality
- Update documentation for API changes
- Use conventional commit messages
- Ensure all tests pass before submitting PR

## Troubleshooting

### Common Issues

#### Database Connection Issues
```bash
# Check if PostgreSQL is running
docker-compose ps postgres

# Check database logs
docker-compose logs postgres

# Test database connection
docker-compose exec postgres psql -U user -d todoapi -c "SELECT 1;"
```

#### JWT Token Issues
- Ensure `JWT_SECRET` is set and at least 32 characters
- Check token expiration time
- Verify Authorization header format: `Bearer <token>`

#### Port Conflicts
```bash
# Check if port 8080 is in use
lsof -i :8080

# Use different port
export PORT=8081
```

### Logs and Debugging
```bash
# View application logs
docker-compose logs api

# Follow logs in real-time
docker-compose logs -f api

# Enable debug logging
export LOG_LEVEL=debug
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For support and questions:
- Create an issue in the repository
- Check the [API documentation](http://localhost:8080/swagger/index.html)
- Review the troubleshooting section above

---

**Note**: This is a demonstration project showcasing Go backend development best practices. For production use, ensure proper security configurations, monitoring, and deployment practices.