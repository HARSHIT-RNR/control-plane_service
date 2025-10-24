package password

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Hasher defines the interface for password hashing
type Hasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) error
}

type bcryptHasher struct {
	cost int
}

// NewBcryptHasher creates a new bcrypt password hasher
func NewBcryptHasher() Hasher {
	return &bcryptHasher{
		cost: bcrypt.DefaultCost,
	}
}

// Hash generates a bcrypt hash of the password
func (h *bcryptHasher) Hash(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// Compare verifies a password against its hash
func (h *bcryptHasher) Compare(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return fmt.Errorf("password mismatch")
	}
	return nil
}
