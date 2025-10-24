package helpers

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ToGRPCError converts an error to a gRPC status error
func ToGRPCError(err error) error {
	if err == nil {
		return nil
	}

	// Map common errors to gRPC codes
	switch {
	case ContainsString(err.Error(), "not found"):
		return status.Errorf(codes.NotFound, err.Error())
	case ContainsString(err.Error(), "invalid"):
		return status.Errorf(codes.InvalidArgument, err.Error())
	case ContainsString(err.Error(), "unauthorized"), ContainsString(err.Error(), "invalid credentials"):
		return status.Errorf(codes.Unauthenticated, err.Error())
	case ContainsString(err.Error(), "permission denied"), ContainsString(err.Error(), "access denied"):
		return status.Errorf(codes.PermissionDenied, err.Error())
	case ContainsString(err.Error(), "already exists"), ContainsString(err.Error(), "duplicate"):
		return status.Errorf(codes.AlreadyExists, err.Error())
	default:
		return status.Errorf(codes.Internal, err.Error())
	}
}

// ContainsString checks if a string contains a substring (case-insensitive)
func ContainsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && contains(s, substr)))
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
