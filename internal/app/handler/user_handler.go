package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Thorlik/marketplace/internal/app/dto"
	"github.com/Thorlik/marketplace/internal/model/domain"
	"github.com/Thorlik/marketplace/internal/service"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type UserHandler struct {
	service service.UserService
	logger  *zap.Logger
}

func NewUserHandler(service service.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		service: service,
		logger:  logger,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Registration data"
// @Success 201 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/register [post]
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling register request")

	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", h.logger)
		return
	}

	user := &domain.User{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      domain.RoleUser,
	}

	createdUser, err := h.service.Register(r.Context(), user)
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			RespondWithError(w, http.StatusConflict, err.Error(), h.logger)
			return
		}
		if errors.Is(err, service.ErrWeakPassword) {
			RespondWithError(w, http.StatusBadRequest, err.Error(), h.logger)
			return
		}
		h.logger.Error("Failed to register user", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to register user", h.logger)
		return
	}

	response := dto.UserResponse{
		ID:        createdUser.ID,
		Email:     createdUser.Email,
		FirstName: createdUser.FirstName,
		LastName:  createdUser.LastName,
		Role:      string(createdUser.Role),
	}

	RespondWithJSON(w, http.StatusCreated, response, h.logger)
}

// Login godoc
// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/login [post]
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling login request")

	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", h.logger)
		return
	}

	user, token, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			RespondWithError(w, http.StatusUnauthorized, err.Error(), h.logger)
			return
		}
		h.logger.Error("Failed to login", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to login", h.logger)
		return
	}

	response := dto.LoginResponse{
		Token: token,
		User: dto.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      string(user.Role),
		},
	}

	RespondWithJSON(w, http.StatusOK, response, h.logger)
}

// GetProfile godoc
// @Summary Get current user profile
// @Description Get authenticated user's profile information
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/me [get]
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*service.Claims)

	user, err := h.service.GetByID(r.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			RespondWithError(w, http.StatusNotFound, "User not found", h.logger)
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to get profile", h.logger)
		return
	}

	response := dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      string(user.Role),
	}

	RespondWithJSON(w, http.StatusOK, response, h.logger)
}

// UpdateProfile godoc
// @Summary Update current user profile
// @Description Update authenticated user's profile information
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdateUserRequest true "Profile update data"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/me [put]
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*service.Claims)

	var req dto.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", h.logger)
		return
	}

	user, err := h.service.Update(r.Context(), claims.UserID, req.FirstName, req.LastName)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			RespondWithError(w, http.StatusNotFound, "User not found", h.logger)
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to update profile", h.logger)
		return
	}

	response := dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      string(user.Role),
	}

	RespondWithJSON(w, http.StatusOK, response, h.logger)
}

// ChangePassword godoc
// @Summary Change user password
// @Description Change authenticated user's password
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.ChangePasswordRequest true "Password change data"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/me/password [put]
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*service.Claims)

	var req dto.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", h.logger)
		return
	}

	err := h.service.ChangePassword(r.Context(), claims.UserID, req.OldPassword, req.NewPassword)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			RespondWithError(w, http.StatusBadRequest, "Invalid old password", h.logger)
			return
		}
		if errors.Is(err, service.ErrWeakPassword) {
			RespondWithError(w, http.StatusBadRequest, err.Error(), h.logger)
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to change password", h.logger)
		return
	}

	RespondWithJSON(w, http.StatusOK, dto.SuccessResponse{Message: "Password changed successfully"}, h.logger)
}

// GetUser godoc
// @Summary Get user by ID (Admin)
// @Description Get user information by ID (admin only)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID", h.logger)
		return
	}

	user, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			RespondWithError(w, http.StatusNotFound, "User not found", h.logger)
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to get user", h.logger)
		return
	}

	response := dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      string(user.Role),
	}

	RespondWithJSON(w, http.StatusOK, response, h.logger)
}

// GetAllUsers godoc
// @Summary Get all users (Admin)
// @Description Get list of all users (admin only)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {array} dto.UserResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users [get]
func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	users, err := h.service.GetAll(r.Context(), page, pageSize)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get users", h.logger)
		return
	}

	var responses []dto.UserResponse
	for _, u := range users {
		responses = append(responses, dto.UserResponse{
			ID:        u.ID,
			Email:     u.Email,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			Role:      string(u.Role),
		})
	}

	RespondWithJSON(w, http.StatusOK, responses, h.logger)
}

// UpdateUserRole godoc
// @Summary Update user role (Admin)
// @Description Update user's role (admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param request body dto.UpdateUserRoleRequest true "New role"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/{id}/role [put]
func (h *UserHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID", h.logger)
		return
	}

	var req dto.UpdateUserRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", h.logger)
		return
	}

	role := domain.UserRole(req.Role)
	if role != domain.RoleUser && role != domain.RoleModerator && role != domain.RoleAdmin {
		RespondWithError(w, http.StatusBadRequest, "Invalid role", h.logger)
		return
	}

	if err := h.service.UpdateRole(r.Context(), id, role); err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			RespondWithError(w, http.StatusNotFound, "User not found", h.logger)
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to update role", h.logger)
		return
	}

	RespondWithJSON(w, http.StatusOK, dto.SuccessResponse{Message: "Role updated successfully"}, h.logger)
}

// DeleteUser godoc
// @Summary Delete user (Admin)
// @Description Delete a user by ID (admin only)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID", h.logger)
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			RespondWithError(w, http.StatusNotFound, "User not found", h.logger)
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete user", h.logger)
		return
	}

	RespondWithJSON(w, http.StatusOK, dto.SuccessResponse{Message: "User deleted successfully"}, h.logger)
}
