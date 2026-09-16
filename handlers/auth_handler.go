package handlers

import (
	"net/http"
	"strconv"

	"eventifyApi/dto"
	"eventifyApi/services"
	"eventifyApi/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService services.AuthService
}

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register godoc
// @Summary Register a new customer
// @Description Register a new user with default customer role
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "User Registration Details"
// @Success 201 {object} utils.APIResponse{data=dto.AuthResponse} "User registered successfully"
// @Failure 400 {object} utils.APIResponse "Bad request"
// @Failure 500 {object} utils.APIResponse "Internal server error"
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	resp, err := h.authService.Register(req)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusCreated, "User registered successfully", resp)
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and return JWT bearer token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login Credentials"
// @Success 200 {object} utils.APIResponse{data=dto.AuthResponse} "Login successful"
// @Failure 400 {object} utils.APIResponse "Invalid credentials or request"
// @Failure 401 {object} utils.APIResponse "Unauthorized"
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	resp, err := h.authService.Login(req)
	if err != nil {
		utils.JSONError(c, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Login successful", resp)
}

// GetProfile godoc
// @Summary Get logged-in user profile
// @Description Retrieve current authenticated user profile
// @Tags Authentication
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=dto.UserProfile} "Profile retrieved"
// @Failure 401 {object} utils.APIResponse "Unauthorized"
// @Failure 500 {object} utils.APIResponse "Internal server error"
// @Router /api/v1/auth/me [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.GetUint64("user_id")

	profile, err := h.authService.GetProfile(userID)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Failed to retrieve profile: "+err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Profile retrieved successfully", profile)
}

// UpdateProfile godoc
// @Summary Update profile
// @Description Update current authenticated user profile information
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdateProfileRequest true "Update Profile Body"
// @Success 200 {object} utils.APIResponse{data=dto.UserProfile} "Profile updated successfully"
// @Failure 400 {object} utils.APIResponse "Invalid request"
// @Failure 401 {object} utils.APIResponse "Unauthorized"
// @Router /api/v1/auth/me [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetUint64("user_id")

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	profile, err := h.authService.UpdateProfile(userID, req)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Profile updated successfully", profile)
}

// ChangePassword godoc
// @Summary Change user password
// @Description Change password for the current authenticated user
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.ChangePasswordRequest true "Change Password Body"
// @Success 200 {object} utils.APIResponse "Password updated successfully"
// @Failure 400 {object} utils.APIResponse "Current password incorrect or validation error"
// @Failure 401 {object} utils.APIResponse "Unauthorized"
// @Router /api/v1/auth/change-password [put]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := c.GetUint64("user_id")

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := h.authService.ChangePassword(userID, req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Password changed successfully", nil)
}

// GetAllUsers godoc
// @Summary List all users (Admin only)
// @Description Paginated list of registered users
// @Tags Admin - Users
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} utils.APIResponse{data=utils.PaginatedData{items=[]dto.UserProfile}} "Users retrieved"
// @Failure 401 {object} utils.APIResponse "Unauthorized"
// @Failure 403 {object} utils.APIResponse "Forbidden"
// @Router /api/v1/admin/users [get]
func (h *AuthHandler) GetAllUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	data, err := h.authService.GetAllUsers(page, limit)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Users retrieved successfully", data)
}

// UpdateUserRole godoc
// @Summary Change user role (Admin only)
// @Description Update role of a specific user (1: Admin, 2: Panitia, 3: Customer)
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param request body dto.UpdateUserRoleRequest true "Role update payload"
// @Success 200 {object} utils.APIResponse "Role updated successfully"
// @Failure 400 {object} utils.APIResponse "Bad request"
// @Failure 401 {object} utils.APIResponse "Unauthorized"
// @Failure 403 {object} utils.APIResponse "Forbidden"
// @Router /api/v1/admin/users/{id}/role [put]
func (h *AuthHandler) UpdateUserRole(c *gin.Context) {
	targetUserID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	var req dto.UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := h.authService.UpdateUserRole(targetUserID, req.RoleID); err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "User role updated successfully", nil)
}

// UpdateUserStatus godoc
// @Summary Update user status (Admin only)
// @Description Update status of a specific user (active or inactive)
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param request body dto.UpdateUserStatusRequest true "Status update payload"
// @Success 200 {object} utils.APIResponse "User status updated successfully"
// @Failure 400 {object} utils.APIResponse "Bad request"
// @Failure 401 {object} utils.APIResponse "Unauthorized"
// @Failure 403 {object} utils.APIResponse "Forbidden"
// @Router /api/v1/admin/users/{id}/status [put]
func (h *AuthHandler) UpdateUserStatus(c *gin.Context) {
	targetUserID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	var req dto.UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := h.authService.UpdateUserStatus(targetUserID, req.Status); err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "User status updated successfully", nil)
}

// ForgotPassword godoc
// @Summary Request password reset token
// @Description Generates a secure reset token sent to email (simulated return for API development)
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.ForgotPasswordRequest true "Forgot Password Request"
// @Success 200 {object} utils.APIResponse "Password reset instructions sent"
// @Failure 400 {object} utils.APIResponse "Bad request"
// @Router /api/v1/auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	token, err := h.authService.ForgotPassword(req)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	// In real-world, token is emailed. Here we return simulated instructions / token for easy API testing
	utils.JSONSuccess(c, http.StatusOK, "If your email is registered, password reset instructions have been sent.", gin.H{
		"reset_token": token,
		"hint":        "Use this token with POST /api/v1/auth/reset-password",
	})
}

// ResetPassword godoc
// @Summary Reset password using token
// @Description Reset account password using the token received from forgot-password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.ResetPasswordRequest true "Reset Password Request"
// @Success 200 {object} utils.APIResponse "Password has been successfully reset"
// @Failure 400 {object} utils.APIResponse "Invalid or expired token"
// @Router /api/v1/auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := h.authService.ResetPassword(req); err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Password has been reset successfully. You may now login.", nil)
}

