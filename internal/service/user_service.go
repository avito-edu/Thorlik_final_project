package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Thorlik/marketplace/internal/model"
	"github.com/Thorlik/marketplace/internal/model/domain"
	"github.com/Thorlik/marketplace/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user with this email already exists")
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrWeakPassword       = errors.New("password must be at least 6 characters")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
)

type UserService interface {
	Register(ctx context.Context, user *domain.User) (*domain.User, error)
	Login(ctx context.Context, email, password string) (*domain.User, string, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	GetAll(ctx context.Context, page, pageSize int) ([]*domain.User, error)
	Update(ctx context.Context, id int64, firstName, lastName string) (*domain.User, error)
	UpdateRole(ctx context.Context, id int64, role domain.UserRole) error
	ChangePassword(ctx context.Context, id int64, oldPassword, newPassword string) error
	Delete(ctx context.Context, id int64) error
	ValidateToken(tokenString string) (*Claims, error)
}

type Claims struct {
	UserID int64           `json:"user_id"`
	Email  string          `json:"email"`
	Role   domain.UserRole `json:"role"`
	jwt.RegisteredClaims
}

type userService struct {
	repo      repository.UserRepository
	jwtSecret string
	jwtExpiry time.Duration
	logger    *zap.Logger
}

func NewUserService(repo repository.UserRepository, jwtSecret string, jwtExpiryHours int, logger *zap.Logger) UserService {
	return &userService{
		repo:      repo,
		jwtSecret: jwtSecret,
		jwtExpiry: time.Duration(jwtExpiryHours) * time.Hour,
		logger:    logger,
	}
}

func (s *userService) Register(ctx context.Context, user *domain.User) (*domain.User, error) {
	s.logger.Info("Registering new user", zap.String("email", user.Email))

	if len(user.Password) < 6 {
		return nil, ErrWeakPassword
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	if user.Role == "" {
		user.Role = domain.RoleUser
	}

	entityUser := model.UserDomainToEntity(user, string(passwordHash))
	createdUser, err := s.repo.Create(ctx, entityUser)
	if err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return nil, ErrUserAlreadyExists
		}
		s.logger.Error("Failed to create user", zap.Error(err))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Info("User registered successfully", zap.Int64("id", createdUser.ID))
	return model.UserEntityToDomain(createdUser), nil
}

func (s *userService) Login(ctx context.Context, email, password string) (*domain.User, string, error) {
	s.logger.Info("User login attempt", zap.String("email", email))

	entityUser, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, "", ErrInvalidCredentials
		}
		s.logger.Error("Failed to get user by email", zap.Error(err))
		return nil, "", fmt.Errorf("failed to get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(entityUser.PasswordHash), []byte(password)); err != nil {
		s.logger.Warn("Invalid password attempt", zap.String("email", email))
		return nil, "", ErrInvalidCredentials
	}

	user := model.UserEntityToDomain(entityUser)
	token, err := s.generateToken(user)
	if err != nil {
		s.logger.Error("Failed to generate token", zap.Error(err))
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	s.logger.Info("User logged in successfully", zap.Int64("id", user.ID))
	return user, token, nil
}

func (s *userService) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	entityUser, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return model.UserEntityToDomain(entityUser), nil
}

func (s *userService) GetAll(ctx context.Context, page, pageSize int) ([]*domain.User, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	entityUsers, err := s.repo.GetAll(ctx, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	users := make([]*domain.User, 0, len(entityUsers))
	for _, eu := range entityUsers {
		users = append(users, model.UserEntityToDomain(eu))
	}

	return users, nil
}

func (s *userService) Update(ctx context.Context, id int64, firstName, lastName string) (*domain.User, error) {
	s.logger.Info("Updating user", zap.Int64("id", id))

	existingUser, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	existingUser.FirstName = firstName
	existingUser.LastName = lastName

	updatedUser, err := s.repo.Update(ctx, existingUser)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return model.UserEntityToDomain(updatedUser), nil
}

func (s *userService) UpdateRole(ctx context.Context, id int64, role domain.UserRole) error {
	s.logger.Info("Updating user role", zap.Int64("id", id), zap.String("role", string(role)))

	if err := s.repo.UpdateRole(ctx, id, string(role)); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to update role: %w", err)
	}

	return nil
}

func (s *userService) ChangePassword(ctx context.Context, id int64, oldPassword, newPassword string) error {
	s.logger.Info("Changing user password", zap.Int64("id", id))

	if len(newPassword) < 6 {
		return ErrWeakPassword
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return ErrInvalidCredentials
	}

	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.repo.UpdatePassword(ctx, id, string(newPasswordHash)); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

func (s *userService) Delete(ctx context.Context, id int64) error {
	s.logger.Info("Deleting user", zap.Int64("id", id))

	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (s *userService) generateToken(user *domain.User) (string, error) {
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *userService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, ErrUnauthorized
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrUnauthorized
}
