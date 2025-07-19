# Design Document

## Overview

The Todo API backend will be built using Go with the Gin web framework, following clean architecture principles and Go best practices. The system will implement a layered architecture with clear separation of concerns, dependency injection, and comprehensive error handling. The application will be containerized using Docker and use PostgreSQL for data persistence.

## Architecture

### High-Level Architecture

```mermaid
graph TB
    Client[Client Applications] --> API[Gin HTTP Server]
    API --> Auth[JWT Middleware]
    Auth --> Handler[HTTP Handlers]
    Handler --> Service[Business Logic Services]
    Service --> Repository[Data Repository Layer]
    Repository --> DB[(PostgreSQL Database)]
    
    Config[Environment Config] --> API
    Config --> DB
```

### Project Structure

```
todo-api-backend/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go              # Configuration management
│   ├── handler/
│   │   ├── auth.go                # Authentication handlers
│   │   ├── todo.go                # Todo CRUD handlers
│   │   └── handler.go             # Handler dependencies
│   ├── middleware/
│   │   ├── auth.go                # JWT authentication middleware
│   │   └── cors.go                # CORS middleware
│   ├── model/
│   │   ├── user.go                # User data model
│   │   ├── todo.go                # Todo data model
│   │   └── response.go            # API response models
│   ├── repository/
│   │   ├── user.go                # User database operations
│   │   ├── todo.go                # Todo database operations
│   │   └── repository.go          # Repository interfaces
│   ├── service/
│   │   ├── auth.go                # Authentication business logic
│   │   ├── todo.go                # Todo business logic
│   │   └── service.go             # Service interfaces
│   └── database/
│       ├── connection.go          # Database connection setup
│       └── migrations/
│           └── init.sql           # Database schema
├── pkg/
│   ├── jwt/
│   │   └── jwt.go                 # JWT token utilities
│   ├── password/
│   │   └── hash.go                # Password hashing utilities
│   └── validator/
│       └── validator.go           # Input validation utilities
├── docs/
│   ├── swagger.json               # OpenAPI specification
│   └── swagger.yaml               # OpenAPI specification (YAML)
├── docker/
│   └── Dockerfile                 # Multi-stage Docker build
├── scripts/
│   └── wait-for-postgres.sh       # Database readiness script
├── tests/
│   ├── handler/
│   ├── service/
│   └── integration/
├── docker-compose.yml
├── .env.example
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Components and Interfaces

### Core Interfaces

#### Repository Layer
```go
type UserRepository interface {
    Create(ctx context.Context, user *model.User) error
    GetByEmail(ctx context.Context, email string) (*model.User, error)
    GetByID(ctx context.Context, id uint) (*model.User, error)
}

type TodoRepository interface {
    Create(ctx context.Context, todo *model.Todo) error
    GetByID(ctx context.Context, id uint, userID uint) (*model.Todo, error)
    GetByUserID(ctx context.Context, userID uint) ([]*model.Todo, error)
    Update(ctx context.Context, todo *model.Todo) error
    Delete(ctx context.Context, id uint, userID uint) error
}
```

#### Service Layer
```go
type AuthService interface {
    Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error)
    Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error)
    ValidateToken(tokenString string) (*jwt.Claims, error)
}

type TodoService interface {
    Create(ctx context.Context, req *CreateTodoRequest, userID uint) (*model.Todo, error)
    GetByID(ctx context.Context, id uint, userID uint) (*model.Todo, error)
    GetByUserID(ctx context.Context, userID uint) ([]*model.Todo, error)
    Update(ctx context.Context, id uint, req *UpdateTodoRequest, userID uint) (*model.Todo, error)
    Delete(ctx context.Context, id uint, userID uint) error
}
```

### HTTP Handler Layer

The handler layer will use Gin's context for request/response handling and dependency injection for services:

```go
type Handler struct {
    authService AuthService
    todoService TodoService
    validator   *validator.Validate
}

func NewHandler(authService AuthService, todoService TodoService) *Handler
```

### Middleware Components

#### JWT Authentication Middleware
- Extracts JWT token from Authorization header
- Validates token signature and expiration
- Adds user ID to Gin context for downstream handlers
- Returns 401 for invalid/missing tokens

#### CORS Middleware
- Configures cross-origin resource sharing
- Allows specified origins, methods, and headers

## Data Models

### User Model
```go
type User struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    Email     string    `json:"email" gorm:"uniqueIndex;not null"`
    Password  string    `json:"-" gorm:"not null"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    Todos     []Todo    `json:"todos,omitempty" gorm:"foreignKey:UserID"`
}
```

### Todo Model
```go
type Todo struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    Title       string    `json:"title" gorm:"not null"`
    Description string    `json:"description"`
    Completed   bool      `json:"completed" gorm:"default:false"`
    UserID      uint      `json:"user_id" gorm:"not null"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    User        User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
}
```

### Request/Response Models
```go
type RegisterRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
    Token string     `json:"token"`
    User  *UserInfo  `json:"user"`
}

type CreateTodoRequest struct {
    Title       string `json:"title" validate:"required,min=1,max=255"`
    Description string `json:"description" validate:"max=1000"`
}

type UpdateTodoRequest struct {
    Title       *string `json:"title,omitempty" validate:"omitempty,min=1,max=255"`
    Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`
    Completed   *bool   `json:"completed,omitempty"`
}
```

## Error Handling

### Error Response Structure
```go
type ErrorResponse struct {
    Error   string            `json:"error"`
    Message string            `json:"message"`
    Details map[string]string `json:"details,omitempty"`
}
```

### HTTP Status Code Mapping
- 200 OK: Successful GET, PUT operations
- 201 Created: Successful POST operations
- 204 No Content: Successful DELETE operations
- 400 Bad Request: Validation errors, malformed requests
- 401 Unauthorized: Authentication failures, invalid tokens
- 404 Not Found: Resource not found or access denied
- 409 Conflict: Duplicate email registration
- 500 Internal Server Error: Unexpected server errors

### Error Handling Strategy
1. **Validation Errors**: Use go-playground/validator for input validation
2. **Database Errors**: Wrap database errors with context
3. **Authentication Errors**: Clear error messages without exposing sensitive info
4. **Logging**: Structured logging with different levels (info, warn, error)
5. **Recovery**: Gin recovery middleware for panic handling

## Database Design

### PostgreSQL Schema
```sql
-- Users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Todos table
CREATE TABLE todos (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    completed BOOLEAN DEFAULT FALSE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_todos_user_id ON todos(user_id);
CREATE INDEX idx_todos_completed ON todos(completed);
CREATE INDEX idx_users_email ON users(email);
```

### Database Connection Strategy
- Use GORM as ORM for database operations
- Connection pooling with configurable pool size
- Health check endpoint for database connectivity
- Graceful shutdown with connection cleanup
- Migration support for schema versioning

## Security Considerations

### Authentication & Authorization
- JWT tokens with configurable expiration (default: 24 hours)
- Secure password hashing using bcrypt (cost factor: 12)
- Token validation on every protected endpoint
- User context isolation (users can only access their own todos)

### Input Validation
- Comprehensive input validation using struct tags
- SQL injection prevention through GORM parameterized queries
- XSS prevention through proper JSON encoding
- Rate limiting considerations for production deployment

### Environment Security
- Sensitive configuration via environment variables
- JWT secret key from secure environment variable
- Database credentials from environment variables
- No hardcoded secrets in source code

## Configuration Management

### Environment Variables
```go
type Config struct {
    Port           string `env:"PORT" envDefault:"8080"`
    DatabaseURL    string `env:"DATABASE_URL" envDefault:"postgres://user:password@localhost/todoapi?sslmode=disable"`
    JWTSecret      string `env:"JWT_SECRET" envDefault:"your-secret-key"`
    JWTExpiration  int    `env:"JWT_EXPIRATION" envDefault:"24"`
    Environment    string `env:"ENVIRONMENT" envDefault:"development"`
    LogLevel       string `env:"LOG_LEVEL" envDefault:"info"`
}
```

## Testing Strategy

### Unit Testing
- **Handler Tests**: Mock services, test HTTP request/response handling
- **Service Tests**: Mock repositories, test business logic
- **Repository Tests**: Use test database or mocks for data layer testing
- **Middleware Tests**: Test JWT validation, CORS headers, error handling

### Integration Testing
- **API Integration Tests**: Test complete request flows
- **Database Integration Tests**: Test with real PostgreSQL instance
- **Authentication Flow Tests**: End-to-end auth testing

### Test Structure
```go
// Example handler test structure
func TestTodoHandler_Create(t *testing.T) {
    tests := []struct {
        name           string
        requestBody    interface{}
        mockSetup      func(*mocks.TodoService)
        expectedStatus int
        expectedBody   string
    }{
        // Test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Testing Tools
- **testify**: Assertions and mocking
- **httptest**: HTTP testing utilities
- **dockertest**: Integration testing with real databases
- **gomock**: Mock generation for interfaces

## API Documentation

### OpenAPI/Swagger Specification
- Complete API documentation with request/response schemas
- Interactive documentation with Swagger UI
- Authentication requirements clearly documented
- Example requests and responses for all endpoints

### Documentation Generation
- Use swaggo/swag for automatic documentation generation from code comments
- Serve documentation at `/swagger/` endpoint
- Include authentication examples and error response formats

## Deployment & DevOps

### Docker Configuration
- **Multi-stage Dockerfile**: Separate build and runtime stages
- **Minimal runtime image**: Use Alpine Linux for small image size
- **Health checks**: Container health check endpoints
- **Non-root user**: Run application as non-root for security

### Docker Compose Setup
- **API Service**: Go application container
- **Database Service**: PostgreSQL container with persistent volume
- **Environment Configuration**: Shared environment variables
- **Network Configuration**: Internal network for service communication
- **Volume Management**: Persistent data storage for PostgreSQL

### Production Considerations
- **Graceful Shutdown**: Handle SIGTERM signals properly
- **Health Check Endpoints**: `/health` and `/ready` endpoints
- **Metrics**: Prometheus metrics for monitoring
- **Logging**: Structured JSON logging for production
- **Configuration Validation**: Validate required environment variables on startup