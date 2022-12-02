package table

import "gorm.io/gorm"

type Competence struct {
	gorm.Model
	Name   string `gorm:"type:varchar(100)"`
	NameID string `gorm:"type:varchar(50);index;unique"`
}

func (Competence) TableName() string {
	return "competence"
}

type Class struct {
	gorm.Model
	Name             string     `gorm:"type:varchar(100)"`
	NameID           string     `gorm:"type:varchar(50);index"`
	Competence       Competence `gorm:"foreignKey:CompetenceNameID;references:NameID"`
	CompetenceNameID string
}

func (Class) TableName() string {
	return "class"
}

type ClassMember struct {
	gorm.Model
	Class         Class
	ClassID       uint
	Users         Users `gorm:"foreignKey:UsersUsername;references:Username"`
	UsersUsername string
}

func (ClassMember) TableName() string {
	return "class_members"
}
