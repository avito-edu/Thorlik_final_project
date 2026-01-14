package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Thorlik/marketplace/internal/app/dto"
	"github.com/Thorlik/marketplace/internal/model/domain"
	"github.com/Thorlik/marketplace/internal/service"
	"github.com/Thorlik/marketplace/internal/service/mocks"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestUserHandler(mockService *mocks.MockUserService) *UserHandler {
	logger := zap.NewNop()
	return NewUserHandler(mockService, logger)
}

func addClaimsToContext(r *http.Request, claims *service.Claims) *http.Request {
	ctx := context.WithValue(r.Context(), "claims", claims)
	return r.WithContext(ctx)
}
func TestUserHandler_Register_Success(t *testing.T) {
	mockService := mocks.NewMockUserService()
	handler := newTestUserHandler(mockService)

	reqBody := dto.RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Ivan",
		LastName:  "Khorolsky",
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("Register", mock.Anything, mock.AnythingOfType("*domain.User")).Return(&domain.User{
		ID:        1,
		Email:     "test@example.com",
		FirstName: "Ivan",
		LastName:  "Khorolsky",
		Role:      domain.RoleUser,
		CreatedAt: time.Now(),
	}, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var response dto.UserResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, int64(1), response.ID)
	assert.Equal(t, "test@example.com", response.Email)
	mockService.AssertExpectations(t)
}

func TestUserHandler_Register_InvalidBody(t *testing.T) {
	mockService := mocks.NewMockUserService()
	handler := newTestUserHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUserHandler_Register_UserAlreadyExists(t *testing.T) {
	mockService := mocks.NewMockUserService()
	handler := newTestUserHandler(mockService)

	reqBody := dto.RegisterRequest{
		Email:     "existing@example.com",
		Password:  "password123",
		FirstName: "Ivan",
		LastName:  "Khorolsky",
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("Register", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil, service.ErrUserAlreadyExists)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_Register_WeakPassword(t *testing.T) {
	mockService := mocks.NewMockUserService()
	handler := newTestUserHandler(mockService)

	reqBody := dto.RegisterRequest{
		Email:     "test@example.com",
		Password:  "short",
		FirstName: "Ivan",
		LastName:  "Khorolsky",
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("Register", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil, service.ErrWeakPassword)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_Login_Success(t *testing.T) {
	mockService := mocks.NewMockUserService()
	handler := newTestUserHandler(mockService)

	reqBody := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("Login", mock.Anything, "test@example.com", "password123").Return(&domain.User{
		ID:        1,
		Email:     "test@example.com",
		FirstName: "Ivan",
		LastName:  "Khorolsky",
		Role:      domain.RoleUser,
	}, "jwt-token", nil)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Login(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response dto.LoginResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "jwt-token", response.Token)
	assert.Equal(t, int64(1), response.User.ID)
	mockService.AssertExpectations(t)
}

func TestUserHandler_Login_InvalidBody(t *testing.T) {
	mockService := mocks.NewMockUserService()
	handler := newTestUserHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Login(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUserHandler_Login_InvalidCredentials(t *testing.T) {
	mockService := mocks.NewMockUserService()
	handler := newTestUserHandler(mockService)

	reqBody := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("Login", mock.Anything, "test@example.com", "wrongpassword").Return(nil, "", service.ErrInvalidCredentials)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Login(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_GetProfile_Success(t *testing.T) {
	mockService := mocks.NewMockUserService()
	handler := newTestUserHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "test@example.com", Role: domain.RoleUser}

	mockService.On("GetByID", mock.Anything, int64(1)).Return(&domain.User{
		ID:        1,
		Email:     "test@example.com",
		FirstName: "Ivan",
		LastName:  "Khorolsky",
		Role:      domain.RoleUser,
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/users/profile", nil)
	req = addClaimsToContext(req, claims)
	rr := httptest.NewRecorder()

	handler.GetProfile(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response dto.UserResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, int64(1), response.ID)
	mockService.AssertExpectations(t)
}

func TestUserHandler_GetProfile_NotFound(t *testing.T) {
	mockService := mocks.NewMockUserService()
	handler := newTestUserHandler(mockService)

	claims := &service.Claims{UserID: 999, Email: "test@example.com", Role: domain.RoleUser}

	mockService.On("GetByID", mock.Anything, int64(999)).Return(nil, service.ErrUserNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/users/profile", nil)
	req = addClaimsToContext(req, claims)
	rr := httptest.NewRecorder()

	handler.GetProfile(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_UpdateProfile_Success(t *testing.T) {
	mockService := mocks.NewMockUserService()
	handler := newTestUserHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "test@example.com", Role: domain.RoleUser}

	reqBody := dto.UpdateUserRequest{
		FirstName: "Ivan",
		LastName:  "Ivanov",
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("Update", mock.Anything, int64(1), "Ivan", "Ivanov").Return(&domain.User{
		ID:        1,
		Email:     "test@example.com",
		FirstName: "Ivan",
		LastName:  "Ivanov",
		Role:      domain.RoleUser,
	}, nil)

	req := httptest.NewRequest(http.MethodPut, "/api/users/profile", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = addClaimsToContext(req, claims)
	rr := httptest.NewRecorder()

	handler.UpdateProfile(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response dto.UserResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "Ivan", response.FirstName)
	assert.Equal(t, "Ivanov", response.LastName)
	mockService.AssertExpectations(t)
}

func TestUserHandler_UpdateProfile_InvalidBody(t *testing.T) {
	mockService := mocks.NewMockUserService()
	handler := newTestUserHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "test@example.com", Role: domain.RoleUser}

	req := httptest.NewRequest(http.MethodPut, "/api/users/profile", bytes.NewBuffer([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	req = addClaimsToContext(req, claims)
	rr := httptest.NewRecorder()

	handler.UpdateProfile(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUserHandler_ChangePassword_Success(t *testing.T) {
	mockService := mocks.NewMockUserService()
	handler := newTestUserHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "test@example.com", Role: domain.RoleUser}

	reqBody := dto.ChangePasswordRequest{
		OldPassword: "oldpass123",
		NewPassword: "newpass123",
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("ChangePassword", mock.Anything, int64(1), "oldpass123", "newpass123").Return(nil)

	req := httptest.NewRequest(http.MethodPut, "/api/users/password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = addClaimsToContext(req, claims)
	rr := httptest.NewRecorder()

	handler.ChangePassword(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_ChangePassword_InvalidOldPassword(t *testing.T) {
	mockService := mocks.NewMockUserService()
	handler := newTestUserHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "test@example.com", Role: domain.RoleUser}

	reqBody := dto.ChangePasswordRequest{
		OldPassword: "wrongold",
		NewPassword: "newpass123",
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("ChangePassword", mock.Anything, int64(1), "wrongold", "newpass123").Return(service.ErrInvalidCredentials)

	req := httptest.NewRequest(http.MethodPut, "/api/users/password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = addClaimsToContext(req, claims)
	rr := httptest.NewRecorder()

	handler.ChangePassword(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_ChangePassword_WeakNewPassword(t *testing.T) {
	mockService := mocks.NewMockUserService()
	handler := newTestUserHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "test@example.com", Role: domain.RoleUser}

	reqBody := dto.ChangePasswordRequest{
		OldPassword: "oldpass123",
		NewPassword: "short",
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("ChangePassword", mock.Anything, int64(1), "oldpass123", "short").Return(service.ErrWeakPassword)

	req := httptest.NewRequest(http.MethodPut, "/api/users/password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = addClaimsToContext(req, claims)
	rr := httptest.NewRecorder()

	handler.ChangePassword(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_GetUser_Success(t *testing.T) {
	mockService := mocks.NewMockUserService()
	handler := newTestUserHandler(mockService)

	mockService.On("GetByID", mock.Anything, int64(1)).Return(&domain.User{
		ID:        1,
		Email:     "test@example.com",
		FirstName: "Ivan",
		LastName:  "Khorolsky",
		Role:      domain.RoleUser,
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/users/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler.GetUser(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_GetUser_InvalidID(t *testing.T) {
	mockService := mocks.NewMockUserService()
	handler := newTestUserHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/api/users/invalid", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "invalid"})
	rr := httptest.NewRecorder()

	handler.GetUser(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUserHandler_GetUser_NotFound(t *testing.T) {
	mockService := mocks.NewMockUserService()
	handler := newTestUserHandler(mockService)

	mockService.On("GetByID", mock.Anything, int64(999)).Return(nil, service.ErrUserNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/users/999", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "999"})
	rr := httptest.NewRecorder()

	handler.GetUser(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_GetAllUsers_Success(t *testing.T) {
	mockService := mocks.NewMockUserService()
	handler := newTestUserHandler(mockService)

	mockService.On("GetAll", mock.Anything, 1, 20).Return([]*domain.User{
		{ID: 1, Email: "user1@example.com", FirstName: "Ivan", LastName: "Khorolsky", Role: domain.RoleUser},
		{ID: 2, Email: "user2@example.com", FirstName: "Ivan", LastName: "Ivanov", Role: domain.RoleUser},
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/users?page=1&page_size=20", nil)
	rr := httptest.NewRecorder()

	handler.GetAllUsers(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response []dto.UserResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)
	assert.Len(t, response, 2)
	mockService.AssertExpectations(t)
}

func TestUserHandler_UpdateUserRole_Success(t *testing.T) {
	mockService := mocks.NewMockUserService()
	handler := newTestUserHandler(mockService)

	reqBody := dto.UpdateUserRoleRequest{Role: "moderator"}
	body, _ := json.Marshal(reqBody)

	mockService.On("UpdateRole", mock.Anything, int64(1), domain.RoleModerator).Return(nil)

	req := httptest.NewRequest(http.MethodPut, "/api/users/1/role", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler.UpdateUserRole(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_UpdateUserRole_InvalidRole(t *testing.T) {
	mockService := mocks.NewMockUserService()
	handler := newTestUserHandler(mockService)

	reqBody := dto.UpdateUserRoleRequest{Role: "superadmin"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/users/1/role", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler.UpdateUserRole(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUserHandler_UpdateUserRole_UserNotFound(t *testing.T) {
	mockService := mocks.NewMockUserService()
	handler := newTestUserHandler(mockService)

	reqBody := dto.UpdateUserRoleRequest{Role: "moderator"}
	body, _ := json.Marshal(reqBody)

	mockService.On("UpdateRole", mock.Anything, int64(999), domain.RoleModerator).Return(service.ErrUserNotFound)

	req := httptest.NewRequest(http.MethodPut, "/api/users/999/role", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": "999"})
	rr := httptest.NewRecorder()

	handler.UpdateUserRole(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_DeleteUser_Success(t *testing.T) {
	mockService := mocks.NewMockUserService()
	handler := newTestUserHandler(mockService)

	mockService.On("Delete", mock.Anything, int64(1)).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/users/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler.DeleteUser(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_DeleteUser_InvalidID(t *testing.T) {
	mockService := mocks.NewMockUserService()
	handler := newTestUserHandler(mockService)

	req := httptest.NewRequest(http.MethodDelete, "/api/users/invalid", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "invalid"})
	rr := httptest.NewRecorder()

	handler.DeleteUser(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUserHandler_DeleteUser_NotFound(t *testing.T) {
	mockService := mocks.NewMockUserService()
	handler := newTestUserHandler(mockService)

	mockService.On("Delete", mock.Anything, int64(999)).Return(service.ErrUserNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/api/users/999", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "999"})
	rr := httptest.NewRecorder()

	handler.DeleteUser(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockService.AssertExpectations(t)
}
