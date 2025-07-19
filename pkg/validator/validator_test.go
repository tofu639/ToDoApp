package validator

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test structs for validation
type TestUser struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,password"`
	Name     string `json:"name" validate:"required,min=2,max=50"`
}

type TestTodo struct {
	Title       string `json:"title" validate:"required,todo_title"`
	Description string `json:"description" validate:"max=1000"`
	Completed   bool   `json:"completed"`
}

type TestOptional struct {
	OptionalField string `json:"optional_field" validate:"omitempty,min=5"`
	RequiredField string `json:"required_field" validate:"required"`
}

func TestNew(t *testing.T) {
	validator := New()
	
	assert.NotNil(t, validator)
	assert.NotNil(t, validator.validate)
}

func TestValidator_ValidateStruct_Success(t *testing.T) {
	validator := New()
	
	tests := []struct {
		name   string
		input  interface{}
	}{
		{
			name: "valid user",
			input: TestUser{
				Email:    "test@example.com",
				Password: "validpassword123",
				Name:     "John Doe",
			},
		},
		{
			name: "valid todo",
			input: TestTodo{
				Title:       "Test Todo",
				Description: "This is a test todo",
				Completed:   false,
			},
		},
		{
			name: "valid optional with value",
			input: TestOptional{
				OptionalField: "valid value",
				RequiredField: "required",
			},
		},
		{
			name: "valid optional without value",
			input: TestOptional{
				OptionalField: "",
				RequiredField: "required",
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateStruct(tt.input)
			assert.NoError(t, err)
		})
	}
}

func TestValidator_ValidateStruct_Errors(t *testing.T) {
	validator := New()
	
	tests := []struct {
		name          string
		input         interface{}
		expectedField string
		expectedTag   string
	}{
		{
			name: "missing email",
			input: TestUser{
				Email:    "",
				Password: "validpassword123",
				Name:     "John Doe",
			},
			expectedField: "email",
			expectedTag:   "required",
		},
		{
			name: "invalid email",
			input: TestUser{
				Email:    "invalid-email",
				Password: "validpassword123",
				Name:     "John Doe",
			},
			expectedField: "email",
			expectedTag:   "email",
		},
		{
			name: "short password",
			input: TestUser{
				Email:    "test@example.com",
				Password: "short",
				Name:     "John Doe",
			},
			expectedField: "password",
			expectedTag:   "password",
		},
		{
			name: "long password",
			input: TestUser{
				Email:    "test@example.com",
				Password: strings.Repeat("a", 129),
				Name:     "John Doe",
			},
			expectedField: "password",
			expectedTag:   "password",
		},
		{
			name: "short name",
			input: TestUser{
				Email:    "test@example.com",
				Password: "validpassword123",
				Name:     "A",
			},
			expectedField: "name",
			expectedTag:   "min",
		},
		{
			name: "empty todo title",
			input: TestTodo{
				Title:       "",
				Description: "Description",
			},
			expectedField: "title",
			expectedTag:   "required",
		},
		{
			name: "whitespace only todo title",
			input: TestTodo{
				Title:       "   ",
				Description: "Description",
			},
			expectedField: "title",
			expectedTag:   "todo_title",
		},
		{
			name: "long todo title",
			input: TestTodo{
				Title:       strings.Repeat("a", 256),
				Description: "Description",
			},
			expectedField: "title",
			expectedTag:   "todo_title",
		},
		{
			name: "long description",
			input: TestTodo{
				Title:       "Valid Title",
				Description: strings.Repeat("a", 1001),
			},
			expectedField: "description",
			expectedTag:   "max",
		},
		{
			name: "short optional field",
			input: TestOptional{
				OptionalField: "abc",
				RequiredField: "required",
			},
			expectedField: "optional_field",
			expectedTag:   "min",
		},
		{
			name: "missing required field",
			input: TestOptional{
				OptionalField: "valid value",
				RequiredField: "",
			},
			expectedField: "required_field",
			expectedTag:   "required",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateStruct(tt.input)
			require.Error(t, err)
			
			validationErrors, ok := err.(ValidationErrors)
			require.True(t, ok)
			require.NotEmpty(t, validationErrors.Errors)
			
			// Check if the expected error is present
			found := false
			for _, validationError := range validationErrors.Errors {
				if validationError.Field == tt.expectedField && validationError.Tag == tt.expectedTag {
					found = true
					assert.NotEmpty(t, validationError.Message)
					break
				}
			}
			assert.True(t, found, "Expected validation error not found: field=%s, tag=%s", tt.expectedField, tt.expectedTag)
		})
	}
}

func TestValidator_ValidateVar(t *testing.T) {
	validator := New()
	
	tests := []struct {
		name      string
		value     interface{}
		tag       string
		shouldErr bool
	}{
		{
			name:      "valid email",
			value:     "test@example.com",
			tag:       "email",
			shouldErr: false,
		},
		{
			name:      "invalid email",
			value:     "invalid-email",
			tag:       "email",
			shouldErr: true,
		},
		{
			name:      "valid password",
			value:     "validpassword123",
			tag:       "password",
			shouldErr: false,
		},
		{
			name:      "invalid password",
			value:     "short",
			tag:       "password",
			shouldErr: true,
		},
		{
			name:      "valid required field",
			value:     "value",
			tag:       "required",
			shouldErr: false,
		},
		{
			name:      "empty required field",
			value:     "",
			tag:       "required",
			shouldErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateVar(tt.value, tt.tag)
			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_CustomValidators(t *testing.T) {
	validator := New()
	
	t.Run("password validator", func(t *testing.T) {
		tests := []struct {
			password string
			valid    bool
		}{
			{"validpass", true},
			{"12345678", true},
			{"short", false},
			{"", false},
			{strings.Repeat("a", 128), true},
			{strings.Repeat("a", 129), false},
		}
		
		for _, tt := range tests {
			err := validator.ValidateVar(tt.password, "password")
			if tt.valid {
				assert.NoError(t, err, "Password should be valid: %s", tt.password)
			} else {
				assert.Error(t, err, "Password should be invalid: %s", tt.password)
			}
		}
	})
	
	t.Run("todo_title validator", func(t *testing.T) {
		tests := []struct {
			title string
			valid bool
		}{
			{"Valid Title", true},
			{"A", true},
			{"", false},
			{"   ", false},
			{strings.Repeat("a", 255), true},
			{strings.Repeat("a", 256), false},
		}
		
		for _, tt := range tests {
			err := validator.ValidateVar(tt.title, "todo_title")
			if tt.valid {
				assert.NoError(t, err, "Title should be valid: %s", tt.title)
			} else {
				assert.Error(t, err, "Title should be invalid: %s", tt.title)
			}
		}
	})
}

func TestValidationErrors_Error(t *testing.T) {
	errors := ValidationErrors{
		Errors: []ValidationError{
			{Field: "email", Message: "email is required"},
			{Field: "password", Message: "password must be at least 8 characters"},
		},
	}
	
	errorMsg := errors.Error()
	assert.Contains(t, errorMsg, "email is required")
	assert.Contains(t, errorMsg, "password must be at least 8 characters")
}

func TestGlobalValidatorFunctions(t *testing.T) {
	user := TestUser{
		Email:    "test@example.com",
		Password: "validpassword123",
		Name:     "John Doe",
	}
	
	err := ValidateStruct(user)
	assert.NoError(t, err)
	
	err = ValidateVar("test@example.com", "email")
	assert.NoError(t, err)
	
	err = ValidateVar("invalid-email", "email")
	assert.Error(t, err)
}

func TestFormatValidationErrors(t *testing.T) {
	t.Run("ValidationErrors type", func(t *testing.T) {
		validationErrors := ValidationErrors{
			Errors: []ValidationError{
				{Field: "email", Message: "email is required"},
				{Field: "password", Message: "password is too short"},
			},
		}
		
		formatted := FormatValidationErrors(validationErrors)
		
		assert.Equal(t, "email is required", formatted["email"])
		assert.Equal(t, "password is too short", formatted["password"])
	})
	
	t.Run("ValidationError type", func(t *testing.T) {
		validationError := ValidationError{
			Field:   "name",
			Message: "name is required",
		}
		
		formatted := FormatValidationErrors(validationError)
		
		assert.Equal(t, "name is required", formatted["name"])
	})
	
	t.Run("generic error", func(t *testing.T) {
		err := errors.New("generic error")
		
		formatted := FormatValidationErrors(err)
		
		assert.Equal(t, "generic error", formatted["general"])
	})
}

func TestValidator_GetErrorMessage(t *testing.T) {
	validator := New()
	
	// Test with a struct that will generate validation errors
	user := TestUser{
		Email:    "",
		Password: "short",
		Name:     strings.Repeat("a", 51),
	}
	
	err := validator.ValidateStruct(user)
	require.Error(t, err)
	
	validationErrors, ok := err.(ValidationErrors)
	require.True(t, ok)
	require.NotEmpty(t, validationErrors.Errors)
	
	// Check that error messages are user-friendly
	for _, validationError := range validationErrors.Errors {
		assert.NotEmpty(t, validationError.Message)
		assert.NotContains(t, validationError.Message, "Key:")
		assert.NotContains(t, validationError.Message, "Error:")
	}
}