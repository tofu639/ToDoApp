package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	ErrValidationFailed = errors.New("validation failed")
)

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// Error implements the error interface
func (ve ValidationError) Error() string {
	return ve.Message
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// Error implements the error interface
func (ve ValidationErrors) Error() string {
	var messages []string
	for _, err := range ve.Errors {
		messages = append(messages, err.Message)
	}
	return strings.Join(messages, "; ")
}

// Validator wraps the go-playground validator with custom functionality
type Validator struct {
	validate *validator.Validate
}

// New creates a new validator instance
func New() *Validator {
	validate := validator.New()
	
	// Register custom tag name function to use JSON tags
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	
	// Register custom validators
	v := &Validator{validate: validate}
	v.registerCustomValidators()
	
	return v
}

// registerCustomValidators registers custom validation functions
func (v *Validator) registerCustomValidators() {
	// Register password strength validator
	v.validate.RegisterValidation("password", v.validatePassword)
	
	// Register todo title validator
	v.validate.RegisterValidation("todo_title", v.validateTodoTitle)
}

// validatePassword validates password strength
func (v *Validator) validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	
	// Minimum length check
	if len(password) < 8 {
		return false
	}
	
	// Maximum length check
	if len(password) > 128 {
		return false
	}
	
	// Additional password strength requirements can be added here
	// For now, we only check length as per the basic requirements
	
	return true
}

// validateTodoTitle validates todo title
func (v *Validator) validateTodoTitle(fl validator.FieldLevel) bool {
	title := strings.TrimSpace(fl.Field().String())
	
	// Must not be empty after trimming
	if len(title) == 0 {
		return false
	}
	
	// Maximum length check
	if len(title) > 255 {
		return false
	}
	
	return true
}

// ValidateStruct validates a struct and returns formatted errors
func (v *Validator) ValidateStruct(s interface{}) error {
	err := v.validate.Struct(s)
	if err == nil {
		return nil
	}
	
	var validationErrors []ValidationError
	
	if validatorErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validatorErrors {
			validationError := ValidationError{
				Field:   fieldError.Field(),
				Tag:     fieldError.Tag(),
				Value:   fmt.Sprintf("%v", fieldError.Value()),
				Message: v.getErrorMessage(fieldError),
			}
			validationErrors = append(validationErrors, validationError)
		}
	}
	
	return ValidationErrors{Errors: validationErrors}
}

// ValidateVar validates a single variable
func (v *Validator) ValidateVar(field interface{}, tag string) error {
	err := v.validate.Var(field, tag)
	if err == nil {
		return nil
	}
	
	if validatorErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validatorErrors {
			return ValidationError{
				Field:   "field",
				Tag:     fieldError.Tag(),
				Value:   fmt.Sprintf("%v", fieldError.Value()),
				Message: v.getErrorMessage(fieldError),
			}
		}
	}
	
	return ErrValidationFailed
}

// getErrorMessage returns a user-friendly error message for validation errors
func (v *Validator) getErrorMessage(fe validator.FieldError) string {
	field := fe.Field()
	tag := fe.Tag()
	param := fe.Param()
	
	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", field, param)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", field, param)
	case "password":
		return fmt.Sprintf("%s must be between 8 and 128 characters long", field)
	case "todo_title":
		return fmt.Sprintf("%s must not be empty and at most 255 characters long", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, param)
	case "numeric":
		return fmt.Sprintf("%s must be a number", field)
	case "alpha":
		return fmt.Sprintf("%s must contain only letters", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// GetValidator returns the underlying validator instance for advanced usage
func (v *Validator) GetValidator() *validator.Validate {
	return v.validate
}

// Global validator instance for convenience
var globalValidator = New()

// ValidateStruct validates a struct using the global validator
func ValidateStruct(s interface{}) error {
	return globalValidator.ValidateStruct(s)
}

// ValidateVar validates a variable using the global validator
func ValidateVar(field interface{}, tag string) error {
	return globalValidator.ValidateVar(field, tag)
}

// FormatValidationErrors formats validation errors for API responses
func FormatValidationErrors(err error) map[string]string {
	errors := make(map[string]string)
	
	if validationErrors, ok := err.(ValidationErrors); ok {
		for _, validationError := range validationErrors.Errors {
			errors[validationError.Field] = validationError.Message
		}
	} else if validationError, ok := err.(ValidationError); ok {
		errors[validationError.Field] = validationError.Message
	} else {
		errors["general"] = err.Error()
	}
	
	return errors
}