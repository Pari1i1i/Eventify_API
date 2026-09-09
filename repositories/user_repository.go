package repositories

import (
	"eventifyApi/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByID(id uint64) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	FindAll(page, limit int) ([]models.User, int64, error)
	UpdateRole(userID uint64, roleID uint8) error

	// Password reset tokens
	CreatePasswordResetToken(reset *models.PasswordResetToken) error
	FindPasswordResetToken(token string) (*models.PasswordResetToken, error)
	DeletePasswordResetToken(token string) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByID(id uint64) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Role").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Role").Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) FindAll(page, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	offset := (page - 1) * limit
	if err := r.db.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Preload("Role").Limit(limit).Offset(offset).Order("id DESC").Find(&users).Error
	return users, total, err
}

func (r *userRepository) UpdateRole(userID uint64, roleID uint8) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("role_id", roleID).Error
}

func (r *userRepository) CreatePasswordResetToken(reset *models.PasswordResetToken) error {
	return r.db.Create(reset).Error
}

func (r *userRepository) FindPasswordResetToken(token string) (*models.PasswordResetToken, error) {
	var pr models.PasswordResetToken
	err := r.db.Where("token = ?", token).First(&pr).Error
	if err != nil {
		return nil, err
	}
	return &pr, nil
}

func (r *userRepository) DeletePasswordResetToken(token string) error {
	return r.db.Where("token = ?", token).Delete(&models.PasswordResetToken{}).Error
}

