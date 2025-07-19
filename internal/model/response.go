package model

// AuthResponse represents the response for authentication endpoints
type AuthResponse struct {
	Token string    `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User  *UserInfo `json:"user"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string            `json:"error" example:"validation_failed"`
	Message string            `json:"message" example:"Invalid input data"`
	Details map[string]string `json:"details,omitempty"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Message string      `json:"message" example:"Operation completed successfully"`
	Data    interface{} `json:"data,omitempty"`
}

// TodoListResponse represents the response for listing todos
type TodoListResponse struct {
	Todos []*Todo `json:"todos"`
	Count int     `json:"count" example:"5"`
}

// HealthResponse represents the response for health check endpoint
type HealthResponse struct {
	Status   string `json:"status" example:"ok"`
	Database string `json:"database" example:"connected"`
	Time     string `json:"time" example:"2024-01-01T12:00:00Z"`
}