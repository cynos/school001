package table

import (
	"time"

	"gorm.io/gorm"
)

type Users struct {
	gorm.Model
	Name         string `gorm:"type:varchar(255)"`
	Email        string `gorm:"type:varchar(100);index"`
	Username     string `gorm:"type:varchar(50);index;unique"`
	PasswordHash string
	LastLogin    time.Time
	Token        string
	Role         string
}

func (Users) TableName() string {
	return "users"
}
