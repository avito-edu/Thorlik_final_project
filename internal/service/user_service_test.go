package service

import (
	"context"
	"testing"
	"time"

	"github.com/Thorlik/marketplace/internal/model/domain"
	"github.com/Thorlik/marketplace/internal/model/entity"
	"github.com/Thorlik/marketplace/internal/repository"
	"github.com/Thorlik/marketplace/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func newTestUserService(mockRepo *mocks.MockUserRepository) UserService {
	logger := zap.NewNop()
	return NewUserService(mockRepo, "test-secret", 24, logger)
}

func TestUserService_Register_Success(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	svc := newTestUserService(mockRepo)
	ctx := context.Background()

	user := &domain.User{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Ivan",
		LastName:  "Khorlsky",
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.User")).Return(&entity.User{
		ID:           1,
		Email:        "test@example.com",
		PasswordHash: "hashed",
		FirstName:    "Ivan",
		LastName:     "Khorlsky",
		Role:         "user",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil)

	result, err := svc.Register(ctx, user)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, "test@example.com", result.Email)
	assert.Equal(t, "Ivan", result.FirstName)
	assert.Equal(t, "Khorlsky", result.LastName)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Register_WeakPassword(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	svc := newTestUserService(mockRepo)
	ctx := context.Background()

	user := &domain.User{
		Email:     "test@example.com",
		Password:  "short",
		FirstName: "Ivan",
		LastName:  "Khorlsky",
	}

	result, err := svc.Register(ctx, user)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrWeakPassword)
}

func TestUserService_Register_UserAlreadyExists(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	svc := newTestUserService(mockRepo)
	ctx := context.Background()

	user := &domain.User{
		Email:     "existing@example.com",
		Password:  "password123",
		FirstName: "Ivan",
		LastName:  "Khorlsky",
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.User")).Return(nil, repository.ErrUserAlreadyExists)

	result, err := svc.Register(ctx, user)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrUserAlreadyExists)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Register_SetsDefaultRole(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	svc := newTestUserService(mockRepo)
	ctx := context.Background()

	user := &domain.User{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Ivan",
		LastName:  "Khorlsky",
		Role:      "",
	}

	mockRepo.On("Create", ctx, mock.MatchedBy(func(u *entity.User) bool {
		return u.Role == string(domain.RoleUser)
	})).Return(&entity.User{
		ID:        1,
		Email:     "test@example.com",
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil)

	result, err := svc.Register(ctx, user)

	require.NoError(t, err)
	assert.Equal(t, domain.RoleUser, result.Role)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Login_Success(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	svc := newTestUserService(mockRepo)
	ctx := context.Background()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	mockRepo.On("GetByEmail", ctx, "test@example.com").Return(&entity.User{
		ID:           1,
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
		FirstName:    "Ivan",
		LastName:     "Khorlsky",
		Role:         "user",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil)

	user, token, err := svc.Login(ctx, "test@example.com", "password123")

	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotEmpty(t, token)
	assert.Equal(t, int64(1), user.ID)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Login_UserNotFound(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	svc := newTestUserService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetByEmail", ctx, "nonexistent@example.com").Return(nil, repository.ErrUserNotFound)

	user, token, err := svc.Login(ctx, "nonexistent@example.com", "password123")

	assert.Nil(t, user)
	assert.Empty(t, token)
	assert.ErrorIs(t, err, ErrInvalidCredentials)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Login_InvalidPassword(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	svc := newTestUserService(mockRepo)
	ctx := context.Background()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctPassword"), bcrypt.DefaultCost)

	mockRepo.On("GetByEmail", ctx, "test@example.com").Return(&entity.User{
		ID:           1,
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
		Role:         "user",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil)

	user, token, err := svc.Login(ctx, "test@example.com", "wrongPassword")

	assert.Nil(t, user)
	assert.Empty(t, token)
	assert.ErrorIs(t, err, ErrInvalidCredentials)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetByID_Success(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	svc := newTestUserService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetByID", ctx, int64(1)).Return(&entity.User{
		ID:        1,
		Email:     "test@example.com",
		FirstName: "Ivan",
		LastName:  "Khorlsky",
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil)

	user, err := svc.GetByID(ctx, 1)

	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, int64(1), user.ID)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetByID_NotFound(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	svc := newTestUserService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetByID", ctx, int64(999)).Return(nil, repository.ErrUserNotFound)

	user, err := svc.GetByID(ctx, 999)

	assert.Nil(t, user)
	assert.ErrorIs(t, err, ErrUserNotFound)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetAll_Success(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	svc := newTestUserService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetAll", ctx, 20, 0).Return([]*entity.User{
		{ID: 1, Email: "user1@example.com", Role: "user", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: 2, Email: "user2@example.com", Role: "user", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}, nil)

	users, err := svc.GetAll(ctx, 1, 20)

	require.NoError(t, err)
	assert.Len(t, users, 2)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetAll_InvalidPagination(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	svc := newTestUserService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetAll", ctx, 20, 0).Return([]*entity.User{}, nil)

	_, err := svc.GetAll(ctx, -1, 200)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Update_Success(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	svc := newTestUserService(mockRepo)
	ctx := context.Background()

	existingUser := &entity.User{
		ID:        1,
		Email:     "test@example.com",
		FirstName: "Ivan",
		LastName:  "Khorlsky",
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo.On("GetByID", ctx, int64(1)).Return(existingUser, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*entity.User")).Return(&entity.User{
		ID:        1,
		Email:     "test@example.com",
		FirstName: "Ivan",
		LastName:  "Ivanov",
		Role:      "user",
		CreatedAt: existingUser.CreatedAt,
		UpdatedAt: time.Now(),
	}, nil)

	user, err := svc.Update(ctx, 1, "Ivan", "Ivanov")

	require.NoError(t, err)
	assert.Equal(t, "Ivan", user.FirstName)
	assert.Equal(t, "Ivanov", user.LastName)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Update_NotFound(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	svc := newTestUserService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetByID", ctx, int64(999)).Return(nil, repository.ErrUserNotFound)

	user, err := svc.Update(ctx, 999, "Ivan", "Ivanov")

	assert.Nil(t, user)
	assert.ErrorIs(t, err, ErrUserNotFound)
	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateRole_Success(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	svc := newTestUserService(mockRepo)
	ctx := context.Background()

	mockRepo.On("UpdateRole", ctx, int64(1), "moderator").Return(nil)

	err := svc.UpdateRole(ctx, 1, domain.RoleModerator)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateRole_NotFound(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	svc := newTestUserService(mockRepo)
	ctx := context.Background()

	mockRepo.On("UpdateRole", ctx, int64(999), "moderator").Return(repository.ErrUserNotFound)

	err := svc.UpdateRole(ctx, 999, domain.RoleModerator)

	assert.ErrorIs(t, err, ErrUserNotFound)
	mockRepo.AssertExpectations(t)
}

func TestUserService_ChangePassword_Success(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	svc := newTestUserService(mockRepo)
	ctx := context.Background()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("oldPassword"), bcrypt.DefaultCost)

	mockRepo.On("GetByID", ctx, int64(1)).Return(&entity.User{
		ID:           1,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil)
	mockRepo.On("UpdatePassword", ctx, int64(1), mock.AnythingOfType("string")).Return(nil)

	err := svc.ChangePassword(ctx, 1, "oldPassword", "newPassword123")

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUserService_ChangePassword_WeakPassword(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	svc := newTestUserService(mockRepo)
	ctx := context.Background()

	err := svc.ChangePassword(ctx, 1, "oldPassword", "short")

	assert.ErrorIs(t, err, ErrWeakPassword)
}

func TestUserService_ChangePassword_WrongOldPassword(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	svc := newTestUserService(mockRepo)
	ctx := context.Background()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctOldPassword"), bcrypt.DefaultCost)

	mockRepo.On("GetByID", ctx, int64(1)).Return(&entity.User{
		ID:           1,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil)

	err := svc.ChangePassword(ctx, 1, "wrongOldPassword", "newPassword123")

	assert.ErrorIs(t, err, ErrInvalidCredentials)
	mockRepo.AssertExpectations(t)
}

func TestUserService_ChangePassword_UserNotFound(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	svc := newTestUserService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetByID", ctx, int64(999)).Return(nil, repository.ErrUserNotFound)

	err := svc.ChangePassword(ctx, 999, "oldPassword", "newPassword123")

	assert.ErrorIs(t, err, ErrUserNotFound)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Delete_Success(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	svc := newTestUserService(mockRepo)
	ctx := context.Background()

	mockRepo.On("Delete", ctx, int64(1)).Return(nil)

	err := svc.Delete(ctx, 1)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Delete_NotFound(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	svc := newTestUserService(mockRepo)
	ctx := context.Background()

	mockRepo.On("Delete", ctx, int64(999)).Return(repository.ErrUserNotFound)

	err := svc.Delete(ctx, 999)

	assert.ErrorIs(t, err, ErrUserNotFound)
	mockRepo.AssertExpectations(t)
}

func TestUserService_ValidateToken_Success(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	svc := newTestUserService(mockRepo)
	ctx := context.Background()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	mockRepo.On("GetByEmail", ctx, "test@example.com").Return(&entity.User{
		ID:           1,
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
		Role:         "user",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil)

	_, token, _ := svc.Login(ctx, "test@example.com", "password123")

	claims, err := svc.ValidateToken(token)

	require.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, int64(1), claims.UserID)
	assert.Equal(t, "test@example.com", claims.Email)
}

func TestUserService_ValidateToken_InvalidToken(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	svc := newTestUserService(mockRepo)

	claims, err := svc.ValidateToken("invalid-token")

	assert.Nil(t, claims)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

func TestUserService_ValidateToken_EmptyToken(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	svc := newTestUserService(mockRepo)

	claims, err := svc.ValidateToken("")

	assert.Nil(t, claims)
	assert.ErrorIs(t, err, ErrUnauthorized)
}
