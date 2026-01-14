package mocks

import (
	"context"

	"github.com/Thorlik/marketplace/internal/model/domain"
	"github.com/Thorlik/marketplace/internal/service"
	"github.com/stretchr/testify/mock"
)

type MockUserService struct {
	mock.Mock
}

func NewMockUserService() *MockUserService {
	return &MockUserService{}
}

func (m *MockUserService) Register(ctx context.Context, user *domain.User) (*domain.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) Login(ctx context.Context, email, password string) (*domain.User, string, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).(*domain.User), args.String(1), args.Error(2)
}

func (m *MockUserService) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) GetAll(ctx context.Context, page, pageSize int) ([]*domain.User, error) {
	args := m.Called(ctx, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.User), args.Error(1)
}

func (m *MockUserService) Update(ctx context.Context, id int64, firstName, lastName string) (*domain.User, error) {
	args := m.Called(ctx, id, firstName, lastName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) UpdateRole(ctx context.Context, id int64, role domain.UserRole) error {
	args := m.Called(ctx, id, role)
	return args.Error(0)
}

func (m *MockUserService) ChangePassword(ctx context.Context, id int64, oldPassword, newPassword string) error {
	args := m.Called(ctx, id, oldPassword, newPassword)
	return args.Error(0)
}

func (m *MockUserService) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserService) ValidateToken(tokenString string) (*service.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.Claims), args.Error(1)
}
