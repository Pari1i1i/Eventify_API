package models

import "time"

type User struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement;type:bigint unsigned" json:"id"`
	RoleID        uint8     `gorm:"type:tinyint unsigned;not null;default:3;index:users_role_id_foreign" json:"role_id"`
	Name          string    `gorm:"type:varchar(150);not null" json:"name"`
	Email         string    `gorm:"type:varchar(150);not null;uniqueIndex:users_email_unique" json:"email"`
	Phone         *string   `gorm:"type:varchar(25)" json:"phone"`
	Password      string    `gorm:"type:varchar(255);not null" json:"-"`
	RememberToken *string   `gorm:"type:varchar(100)" json:"-"`
	Status        string    `gorm:"type:varchar(10);not null;default:'active'" json:"status"`
	CreatedAt     time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP" json:"updated_at"`

	Role *Role `gorm:"foreignKey:RoleID;constraint:OnUpdate:CASCADE" json:"role,omitempty"`
}

func (User) TableName() string {
	return "users"
}
