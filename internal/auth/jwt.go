package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

type JWTService interface {
	GenerateToken(userID uuid.UUID, email string) (string, error)
	ValidateToken(tokenString string) (*Claims, error)
}

type jwtService struct {
	secretKey  string
	expiration int
}

func NewJWTService(secretKey string, expiration int) JWTService {
	return &jwtService{
		secretKey:  secretKey,
		expiration: expiration,
	}
}

func (s *jwtService) GenerateToken(userID uuid.UUID, email string) (string, error) {
	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.expiration) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func (s *jwtService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
