package services

import (
	"errors"
	"time"

	"eventifyApi/config"
	"eventifyApi/dto"
	"eventifyApi/models"
	"eventifyApi/repositories"
	"eventifyApi/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthService interface {
	Register(req dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(req dto.LoginRequest) (*dto.AuthResponse, error)
	GetProfile(userID uint64) (*dto.UserProfile, error)
	UpdateProfile(userID uint64, req dto.UpdateProfileRequest) (*dto.UserProfile, error)
	ChangePassword(userID uint64, req dto.ChangePasswordRequest) error
	GetAllUsers(page, limit int) (*utils.PaginatedData, error)
	UpdateUserRole(userID uint64, roleID uint8) error
	UpdateUserStatus(userID uint64, status string) error
	ForgotPassword(req dto.ForgotPasswordRequest) (string, error)
	ResetPassword(req dto.ResetPasswordRequest) error
}

type authService struct {
	userRepo repositories.UserRepository
	cfg      *config.Config
}

func NewAuthService(userRepo repositories.UserRepository, cfg *config.Config) AuthService {
	return &authService{userRepo: userRepo, cfg: cfg}
}

func (s *authService) Register(req dto.RegisterRequest) (*dto.AuthResponse, error) {
	existing, _ := s.userRepo.FindByEmail(req.Email)
	if existing != nil {
		return nil, errors.New("email is already registered")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	status := req.Status
	if status == "" {
		status = "active"
	}
	// Validate status
	if status != "active" && status != "inactive" {
		return nil, errors.New("status must be either 'active' or 'inactive'")
	}

	user := &models.User{
		RoleID:   3, // default customer
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: hashedPassword,
		Status:   status,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	token, err := utils.GenerateToken(user.ID, user.Email, user.RoleID, s.cfg.JWTSecret, s.cfg.JWTExpirationHours)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		User: dto.UserProfile{
			ID:        user.ID,
			RoleID:    user.RoleID,
			RoleName:  "customer",
			Name:      user.Name,
			Email:     user.Email,
			Phone:     user.Phone,
			CreatedAt: user.CreatedAt,
		},
	}, nil
}

func (s *authService) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return nil, errors.New("invalid email or password")
	}

	// Check if user is active
	if user.Status != "active" {
		if user.Status == "suspended" {
			return nil, errors.New("akun Anda telah disuspend oleh administrator")
		}
		return nil, errors.New("account is inactive")
	}

	token, err := utils.GenerateToken(user.ID, user.Email, user.RoleID, s.cfg.JWTSecret, s.cfg.JWTExpirationHours)
	if err != nil {
		return nil, err
	}

	roleName := "customer"
	if user.Role != nil {
		roleName = user.Role.Name
	}

	return &dto.AuthResponse{
		Token: token,
		User: dto.UserProfile{
			ID:        user.ID,
			RoleID:    user.RoleID,
			RoleName:  roleName,
			Name:      user.Name,
			Email:     user.Email,
			Phone:     user.Phone,
			Status:    user.Status,
			CreatedAt: user.CreatedAt,
		},
	}, nil
}

func (s *authService) GetProfile(userID uint64) (*dto.UserProfile, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	roleName := "customer"
	if user.Role != nil {
		roleName = user.Role.Name
	}

	return &dto.UserProfile{
		ID:        user.ID,
		RoleID:    user.RoleID,
		RoleName:  roleName,
		Name:      user.Name,
		Email:     user.Email,
		Phone:     user.Phone,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *authService) UpdateProfile(userID uint64, req dto.UpdateProfileRequest) (*dto.UserProfile, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	user.Name = req.Name
	user.Phone = req.Phone
	if req.Status != "" {
		// Validate status
		if req.Status != "active" && req.Status != "inactive" {
			return nil, errors.New("status must be either 'active' or 'inactive'")
		}
		user.Status = req.Status
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return s.GetProfile(userID)
}

func (s *authService) ChangePassword(userID uint64, req dto.ChangePasswordRequest) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	if !utils.CheckPasswordHash(req.OldPassword, user.Password) {
		return errors.New("current password does not match")
	}

	newHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return errors.New("failed to hash new password")
	}

	user.Password = newHash
	return s.userRepo.Update(user)
}

func (s *authService) GetAllUsers(page, limit int) (*utils.PaginatedData, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	users, total, err := s.userRepo.FindAll(page, limit)
	if err != nil {
		return nil, err
	}

	var userProfiles []dto.UserProfile
	for _, u := range users {
		roleName := "customer"
		if u.Role != nil {
			roleName = u.Role.Name
		}
		userProfiles = append(userProfiles, dto.UserProfile{
			ID:        u.ID,
			RoleID:    u.RoleID,
			RoleName:  roleName,
			Name:      u.Name,
			Email:     u.Email,
			Phone:     u.Phone,
			Status:    u.Status,
			CreatedAt: u.CreatedAt,
		})
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &utils.PaginatedData{
		Items:      userProfiles,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *authService) UpdateUserRole(userID uint64, roleID uint8) error {
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	return s.userRepo.UpdateRole(userID, roleID)
}

func (s *authService) UpdateUserStatus(userID uint64, status string) error {
	// Validate status
	if status != "active" && status != "inactive" && status != "suspended" {
		return errors.New("status must be 'active', 'inactive', or 'suspended'")
	}
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	return s.userRepo.UpdateStatus(userID, status)
}

func (s *authService) ForgotPassword(req dto.ForgotPasswordRequest) (string, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		// Return generic success to avoid user enumeration
		return "", nil
	}

	token := uuid.New().String()
	pr := &models.PasswordResetToken{
		Email:     user.Email,
		Token:     token,
		ExpiresAt: time.Now().Add(15 * time.Minute), // 15 mins expiry
	}

	if err := s.userRepo.CreatePasswordResetToken(pr); err != nil {
		return "", err
	}

	return token, nil
}

func (s *authService) ResetPassword(req dto.ResetPasswordRequest) error {
	pr, err := s.userRepo.FindPasswordResetToken(req.Token)
	if err != nil {
		return errors.New("invalid or expired password reset token")
	}

	if time.Now().After(pr.ExpiresAt) {
		_ = s.userRepo.DeletePasswordResetToken(req.Token)
		return errors.New("password reset token has expired")
	}

	user, err := s.userRepo.FindByEmail(pr.Email)
	if err != nil {
		return errors.New("user account not found")
	}

	newHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return errors.New("failed to hash new password")
	}

	user.Password = newHash
	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	_ = s.userRepo.DeletePasswordResetToken(req.Token)
	return nil
}

