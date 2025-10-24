package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims represents the JWT claims
type Claims struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

// Generator defines the interface for token generation
type Generator interface {
	GenerateAccessToken(userID, tenantID, email string) (string, error)
	GenerateRefreshToken(userID, tenantID, email string) (string, error)
}

// Validator defines the interface for token validation
type Validator interface {
	ValidateToken(tokenString string) (*Claims, error)
}

type jwtGenerator struct {
	secretKey            string
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

// NewJWTGenerator creates a new JWT token generator
func NewJWTGenerator(secretKey string, accessTokenDuration, refreshTokenDuration time.Duration) Generator {
	return &jwtGenerator{
		secretKey:            secretKey,
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
	}
}

func (g *jwtGenerator) GenerateAccessToken(userID, tenantID, email string) (string, error) {
	return g.generateToken(userID, tenantID, email, g.accessTokenDuration)
}

func (g *jwtGenerator) GenerateRefreshToken(userID, tenantID, email string) (string, error) {
	return g.generateToken(userID, tenantID, email, g.refreshTokenDuration)
}

func (g *jwtGenerator) generateToken(userID, tenantID, email string, duration time.Duration) (string, error) {
	claims := Claims{
		UserID:   userID,
		TenantID: tenantID,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(g.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

type jwtValidator struct {
	secretKey string
}

// NewJWTValidator creates a new JWT token validator
func NewJWTValidator(secretKey string) Validator {
	return &jwtValidator{
		secretKey: secretKey,
	}
}

func (v *jwtValidator) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(v.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
