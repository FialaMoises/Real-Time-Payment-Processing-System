package user

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	apperrors "github.com/yourusername/real-time-payments/pkg/errors"
)

type Repository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByCPF(ctx context.Context, cpf string) (*User, error)
	Update(ctx context.Context, user *User) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (email, password_hash, full_name, cpf, phone, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.CPF,
		user.Phone,
		user.Status,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err.Error() == "pq: duplicate key value violates unique constraint \"users_email_key\"" {
			return apperrors.ErrDuplicateEntry
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	query := `
		SELECT id, email, password_hash, full_name, cpf, phone, status, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.CPF,
		&user.Phone,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

func (r *repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, password_hash, full_name, cpf, phone, status, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.CPF,
		&user.Phone,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

func (r *repository) GetByCPF(ctx context.Context, cpf string) (*User, error) {
	query := `
		SELECT id, email, password_hash, full_name, cpf, phone, status, created_at, updated_at
		FROM users
		WHERE cpf = $1
	`

	user := &User{}
	err := r.db.QueryRowContext(ctx, query, cpf).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.CPF,
		&user.Phone,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by CPF: %w", err)
	}

	return user, nil
}

func (r *repository) Update(ctx context.Context, user *User) error {
	query := `
		UPDATE users
		SET full_name = $1, phone = $2, status = $3
		WHERE id = $4
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		user.FullName,
		user.Phone,
		user.Status,
		user.ID,
	).Scan(&user.UpdatedAt)

	if err == sql.ErrNoRows {
		return apperrors.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}
