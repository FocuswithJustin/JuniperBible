// Package errors provides standardized error types and helpers for the Mimicry codebase.
package errors

import (
	"errors"
	"fmt"
)

// Sentinel errors for common cases
var (
	// ErrNotFound indicates a resource was not found
	ErrNotFound = errors.New("not found")
	// ErrInvalidInput indicates invalid input or validation failure
	ErrInvalidInput = errors.New("invalid input")
	// ErrAlreadyExists indicates a resource already exists
	ErrAlreadyExists = errors.New("already exists")
	// ErrUnauthorized indicates insufficient permissions
	ErrUnauthorized = errors.New("unauthorized")
	// ErrInternal indicates an internal system error
	ErrInternal = errors.New("internal error")
	// ErrUnsupported indicates an unsupported operation or format
	ErrUnsupported = errors.New("unsupported")
)

// NotFoundError represents a resource not found error with context
type NotFoundError struct {
	Resource string // Type of resource (e.g., "plugin", "artifact", "capsule")
	ID       string // Identifier of the resource
	Err      error  // Underlying error, if any
}

func (e *NotFoundError) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
	}
	return fmt.Sprintf("%s not found", e.Resource)
}

func (e *NotFoundError) Unwrap() error {
	if e.Err != nil {
		return e.Err
	}
	return ErrNotFound
}

// ValidationError represents an input validation error with context
type ValidationError struct {
	Field   string // Field name that failed validation
	Value   string // Value that failed validation (may be redacted)
	Message string // Human-readable error message
	Err     error  // Underlying error, if any
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation failed: %s", e.Message)
}

func (e *ValidationError) Unwrap() error {
	if e.Err != nil {
		return e.Err
	}
	return ErrInvalidInput
}

// PermissionError represents an authorization/permission error
type PermissionError struct {
	Operation string // Operation that was attempted
	Resource  string // Resource being accessed
	Reason    string // Why permission was denied
	Err       error  // Underlying error, if any
}

func (e *PermissionError) Error() string {
	if e.Operation != "" && e.Resource != "" {
		return fmt.Sprintf("permission denied: cannot %s %s: %s", e.Operation, e.Resource, e.Reason)
	}
	return fmt.Sprintf("permission denied: %s", e.Reason)
}

func (e *PermissionError) Unwrap() error {
	if e.Err != nil {
		return e.Err
	}
	return ErrUnauthorized
}

// IOError represents an I/O operation error with context
type IOError struct {
	Operation string // Operation being performed (e.g., "read", "write", "open")
	Path      string // File/resource path involved
	Err       error  // Underlying error
}

func (e *IOError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("failed to %s %s: %v", e.Operation, e.Path, e.Err)
	}
	return fmt.Sprintf("failed to %s: %v", e.Operation, e.Err)
}

func (e *IOError) Unwrap() error {
	return e.Err
}

// ParseError represents a parsing or deserialization error
type ParseError struct {
	Format  string // Format being parsed (e.g., "JSON", "XML", "manifest")
	Path    string // File path, if applicable
	Message string // Error details
	Err     error  // Underlying error, if any
}

func (e *ParseError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("failed to parse %s at %s: %s", e.Format, e.Path, e.Message)
	}
	return fmt.Sprintf("failed to parse %s: %s", e.Format, e.Message)
}

func (e *ParseError) Unwrap() error {
	if e.Err != nil {
		return e.Err
	}
	return ErrInvalidInput
}

// UnsupportedError represents an unsupported feature or format
type UnsupportedError struct {
	Feature string // Feature or format that is unsupported
	Reason  string // Why it's not supported
	Err     error  // Underlying error, if any
}

func (e *UnsupportedError) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("unsupported %s: %s", e.Feature, e.Reason)
	}
	return fmt.Sprintf("unsupported %s", e.Feature)
}

func (e *UnsupportedError) Unwrap() error {
	if e.Err != nil {
		return e.Err
	}
	return ErrUnsupported
}

// Helper functions for creating common errors

// NewNotFound creates a NotFoundError
func NewNotFound(resource, id string) *NotFoundError {
	return &NotFoundError{
		Resource: resource,
		ID:       id,
	}
}

// NewValidation creates a ValidationError
func NewValidation(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// NewPermission creates a PermissionError
func NewPermission(operation, resource, reason string) *PermissionError {
	return &PermissionError{
		Operation: operation,
		Resource:  resource,
		Reason:    reason,
	}
}

// NewIO creates an IOError
func NewIO(operation, path string, err error) *IOError {
	return &IOError{
		Operation: operation,
		Path:      path,
		Err:       err,
	}
}

// NewParse creates a ParseError
func NewParse(format, path, message string) *ParseError {
	return &ParseError{
		Format:  format,
		Path:    path,
		Message: message,
	}
}

// NewUnsupported creates an UnsupportedError
func NewUnsupported(feature, reason string) *UnsupportedError {
	return &UnsupportedError{
		Feature: feature,
		Reason:  reason,
	}
}

// Wrap adds context to an error. If err is nil, returns nil.
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// Wrapf adds formatted context to an error. If err is nil, returns nil.
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	message := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s: %w", message, err)
}

// Is wraps errors.Is for convenience
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As wraps errors.As for convenience
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}
