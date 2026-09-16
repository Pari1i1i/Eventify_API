package dto

import "time"

type RegisterRequest struct {
	Name     string  `json:"name" binding:"required,min=2,max=150" example:"Fachri Ramadhan"`
	Email    string  `json:"email" binding:"required,email,max=150" example:"fachri@example.com"`
	Phone    *string `json:"phone" binding:"omitempty,max=25" example:"081234567890"`
	Password string  `json:"password" binding:"required,min=6" example:"secret123"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"fachri@example.com"`
	Password string `json:"password" binding:"required" example:"secret123"`
}

type AuthResponse struct {
	Token string      `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User  UserProfile `json:"user"`
}

type UserProfile struct {
	ID        uint64    `json:"id" example:"1"`
	RoleID    uint8     `json:"role_id" example:"3"`
	RoleName  string    `json:"role_name" example:"customer"`
	Name      string    `json:"name" example:"Fachri Ramadhan"`
	Email     string    `json:"email" example:"fachri@example.com"`
	Phone     *string   `json:"phone" example:"081234567890"`
	Status    string    `json:"status" example:"active"` // active or inactive
	CreatedAt time.Time `json:"created_at"`
}

type UpdateProfileRequest struct {
	Name  string  `json:"name" binding:"required,min=2,max=150" example:"Fachri Ramadhan"`
	Phone *string `json:"phone" binding:"omitempty,max=25" example:"081234567899"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required" example:"old_secret"`
	NewPassword string `json:"new_password" binding:"required,min=6" example:"new_secret123"`
}

type UpdateUserRoleRequest struct {
	RoleID uint8 `json:"role_id" binding:"required,oneof=1 2 3" example:"2"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email" example:"customer@example.com"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required" example:"a1b2c3d4e5f6..."`
	NewPassword string `json:"new_password" binding:"required,min=6" example:"new_secret123"`
}

// UpdateUserStatusRequest is used to update a user's status (active/inactive)
type UpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active inactive" example:"inactive"`
}
