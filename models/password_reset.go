package models

import "time"

type PasswordResetToken struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement;type:bigint unsigned" json:"id"`
	Email     string    `gorm:"type:varchar(150);not null;index:idx_password_resets_email" json:"email"`
	Token     string    `gorm:"type:varchar(100);not null;uniqueIndex:idx_password_resets_token" json:"token"`
	ExpiresAt time.Time `gorm:"type:datetime;not null" json:"expires_at"`
	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}
