package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Thorlik/marketplace/internal/model/entity"
	"github.com/Thorlik/marketplace/internal/pkg/db"
	"go.uber.org/zap"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user with this email already exists")
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) (*entity.User, error)
	GetByID(ctx context.Context, id int64) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entity.User, error)
	Update(ctx context.Context, user *entity.User) (*entity.User, error)
	UpdateRole(ctx context.Context, id int64, role string) error
	UpdatePassword(ctx context.Context, id int64, passwordHash string) error
	Delete(ctx context.Context, id int64) error
	Count(ctx context.Context) (int64, error)
}

type userRepository struct {
	txManager *db.TransactionManager
	qManager  *db.QueryManager
	logger    *zap.Logger
}

func NewUserRepository(txManager *db.TransactionManager, qManager *db.QueryManager, logger *zap.Logger) UserRepository {
	return &userRepository{
		txManager: txManager,
		qManager:  qManager,
		logger:    logger,
	}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	r.logger.Info("Creating user", zap.String("email", user.Email))

	var createdUser entity.User
	err := r.txManager.WithTransaction(ctx, func(tx *db.Transaction) error {
		query := `
			INSERT INTO users (email, password_hash, first_name, last_name, role)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, email, password_hash, first_name, last_name, role, created_at, updated_at
		`
		row := tx.QueryRow(ctx, query, user.Email, user.PasswordHash, user.FirstName, user.LastName, user.Role)
		if err := row.Scan(&createdUser.ID, &createdUser.Email, &createdUser.PasswordHash,
			&createdUser.FirstName, &createdUser.LastName, &createdUser.Role,
			&createdUser.CreatedAt, &createdUser.UpdatedAt); err != nil {
			if isUniqueViolation(err) {
				return ErrUserAlreadyExists
			}
			r.logger.Error("Failed to create user", zap.Error(err))
			return fmt.Errorf("failed to create user: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	r.logger.Info("User created successfully", zap.Int64("id", createdUser.ID))
	return &createdUser, nil
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	r.logger.Debug("Getting user by ID", zap.Int64("id", id))

	query := `
		SELECT id, email, password_hash, first_name, last_name, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	row := r.qManager.QueryRow(ctx, query, id)

	var user entity.User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FirstName,
		&user.LastName, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		r.logger.Error("Failed to get user", zap.Error(err), zap.Int64("id", id))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	r.logger.Debug("Getting user by email", zap.String("email", email))

	query := `
		SELECT id, email, password_hash, first_name, last_name, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	row := r.qManager.QueryRow(ctx, query, email)

	var user entity.User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FirstName,
		&user.LastName, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		r.logger.Error("Failed to get user", zap.Error(err), zap.String("email", email))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *userRepository) GetAll(ctx context.Context, limit, offset int) ([]*entity.User, error) {
	r.logger.Debug("Getting all users", zap.Int("limit", limit), zap.Int("offset", offset))

	query := `
		SELECT id, email, password_hash, first_name, last_name, role, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.qManager.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*entity.User
	for rows.Next() {
		var user entity.User
		if err := rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FirstName,
			&user.LastName, &user.Role, &user.CreatedAt, &user.UpdatedAt); err != nil {
			r.logger.Error("Failed to scan user", zap.Error(err))
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	return users, nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) (*entity.User, error) {
	r.logger.Info("Updating user", zap.Int64("id", user.ID))

	var updatedUser entity.User
	err := r.txManager.WithTransaction(ctx, func(tx *db.Transaction) error {
		query := `
			UPDATE users
			SET first_name = $1, last_name = $2
			WHERE id = $3
			RETURNING id, email, password_hash, first_name, last_name, role, created_at, updated_at
		`
		row := tx.QueryRow(ctx, query, user.FirstName, user.LastName, user.ID)
		if err := row.Scan(&updatedUser.ID, &updatedUser.Email, &updatedUser.PasswordHash,
			&updatedUser.FirstName, &updatedUser.LastName, &updatedUser.Role,
			&updatedUser.CreatedAt, &updatedUser.UpdatedAt); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrUserNotFound
			}
			r.logger.Error("Failed to update user", zap.Error(err))
			return fmt.Errorf("failed to update user: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &updatedUser, nil
}

func (r *userRepository) UpdateRole(ctx context.Context, id int64, role string) error {
	r.logger.Info("Updating user role", zap.Int64("id", id), zap.String("role", role))

	err := r.txManager.WithTransaction(ctx, func(tx *db.Transaction) error {
		query := `UPDATE users SET role = $1 WHERE id = $2`
		result, err := tx.Exec(ctx, query, role, id)
		if err != nil {
			return fmt.Errorf("failed to update role: %w", err)
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return ErrUserNotFound
		}

		return nil
	})

	return err
}

func (r *userRepository) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {
	r.logger.Info("Updating user password", zap.Int64("id", id))

	err := r.txManager.WithTransaction(ctx, func(tx *db.Transaction) error {
		query := `UPDATE users SET password_hash = $1 WHERE id = $2`
		result, err := tx.Exec(ctx, query, passwordHash, id)
		if err != nil {
			return fmt.Errorf("failed to update password: %w", err)
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return ErrUserNotFound
		}

		return nil
	})

	return err
}

func (r *userRepository) Delete(ctx context.Context, id int64) error {
	r.logger.Info("Deleting user", zap.Int64("id", id))

	err := r.txManager.WithTransaction(ctx, func(tx *db.Transaction) error {
		query := `DELETE FROM users WHERE id = $1`
		result, err := tx.Exec(ctx, query, id)
		if err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return ErrUserNotFound
		}

		return nil
	})

	return err
}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM users`
	row := r.qManager.QueryRow(ctx, query)

	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

func isUniqueViolation(err error) bool {
	return err != nil && (contains(err.Error(), "unique") || contains(err.Error(), "duplicate"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
