package errors

import (
	"database/sql"
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// --- Custom Application-Specific Errors ---
// These errors are used by the service layer to signal specific business rule failures.
// The gRPC handler layer then maps these to appropriate gRPC status codes.

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrRoleNotFound       = errors.New("role not found")
	ErrDuplicateEmail     = errors.New("email address is already in use")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrPermissionDenied   = errors.New("permission denied")
	ErrTokenInvalid       = errors.New("token is invalid or expired")
)

// ValidationError is a custom error type for input validation.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
}

// MapErrorToGRPCStatus maps an application error to a gRPC status code.
// This is a crucial part of the gRPC handler layer.
func MapErrorToGRPCStatus(err error) error {
	if err == nil {
		return nil
	}

	// --- Specific Error Mapping ---
	switch {
	// Not Found Errors
	case errors.Is(err, sql.ErrNoRows):
		return status.Error(codes.NotFound, "resource not found")
	case errors.Is(err, ErrUserNotFound):
		return status.Error(codes.NotFound, ErrUserNotFound.Error())
	case errors.Is(err, ErrRoleNotFound):
		return status.Error(codes.NotFound, ErrRoleNotFound.Error())

	// Already Exists / Conflict Errors
	case errors.Is(err, ErrDuplicateEmail):
		return status.Error(codes.AlreadyExists, ErrDuplicateEmail.Error())

	// Authentication/Authorization Errors
	case errors.Is(err, ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, ErrInvalidCredentials.Error())
	case errors.Is(err, ErrPermissionDenied):
		return status.Error(codes.PermissionDenied, ErrPermissionDenied.Error())
	case errors.Is(err, ErrTokenInvalid):
		return status.Error(codes.Unauthenticated, ErrTokenInvalid.Error())

	// Invalid Argument Errors
	case errors.As(err, new(*ValidationError)):
		var validationErr *ValidationError
		errors.As(err, &validationErr)
		return status.Error(codes.InvalidArgument, validationErr.Error())
	}

	// Default to an internal error for unhandled cases
	return status.Error(codes.Internal, "an unexpected error occurred")
}
