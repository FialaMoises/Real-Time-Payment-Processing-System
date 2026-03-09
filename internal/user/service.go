package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"github.com/yourusername/real-time-payments/internal/auth"
	apperrors "github.com/yourusername/real-time-payments/pkg/errors"
)

type Service interface {
	Register(ctx context.Context, req *CreateUserRequest) (*User, error)
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
}

type service struct {
	repo       Repository
	jwtService auth.JWTService
	jwtExp     int
}

func NewService(repo Repository, jwtService auth.JWTService, jwtExp int) Service {
	return &service{
		repo:       repo,
		jwtService: jwtService,
		jwtExp:     jwtExp,
	}
}

func (s *service) Register(ctx context.Context, req *CreateUserRequest) (*User, error) {
	// Check if email already exists
	existingUser, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil && err != apperrors.ErrNotFound {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if existingUser != nil {
		return nil, apperrors.ErrDuplicateEntry
	}

	// Check if CPF already exists
	existingCPF, err := s.repo.GetByCPF(ctx, req.CPF)
	if err != nil && err != apperrors.ErrNotFound {
		return nil, fmt.Errorf("failed to check CPF: %w", err)
	}
	if existingCPF != nil {
		return nil, apperrors.New("DUPLICATE_CPF", "CPF already registered", 409)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FullName:     req.FullName,
		CPF:          req.CPF,
		Phone:        req.Phone,
		Status:       "ACTIVE",
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Don't return password hash
	user.PasswordHash = ""

	return user, nil
}

func (s *service) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		if err == apperrors.ErrNotFound {
			return nil, apperrors.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, apperrors.ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := s.jwtService.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Don't return password hash
	user.PasswordHash = ""

	return &LoginResponse{
		AccessToken: token,
		ExpiresIn:   s.jwtExp,
		User:        user,
	}, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Don't return password hash
	user.PasswordHash = ""

	return user, nil
}
