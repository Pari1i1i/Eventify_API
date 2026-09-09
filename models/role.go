package models

type Role struct {
	ID   uint8  `gorm:"primaryKey;autoIncrement;type:tinyint unsigned" json:"id"`
	Name string `gorm:"type:varchar(50);not null;uniqueIndex:roles_name_unique" json:"name"`
}

func (Role) TableName() string {
	return "roles"
}
