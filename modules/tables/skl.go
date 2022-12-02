package table

import "gorm.io/gorm"

type SKL struct {
	gorm.Model
	Graduated     bool
	Score         float32
	Users         Users `gorm:"foreignKey:UsersUsername;references:Username"`
	UsersUsername string
	Class         Class
	ClassID       uint
}

func (SKL) TableName() string {
	return "skl"
}

type SKLDetails struct {
	gorm.Model
	NISN                      string
	NIS                       string
	PaketKeahlian             string
	NamaPeserta               string
	TTL                       string
	NamaOrtu                  string
	NilaiAgama                string
	NilaiPKN                  string
	NilaiBindo                string
	NilaiMTK                  string
	NilaiSI                   string
	NilaiBing                 string
	NilaiSenbud               string
	NilaiPenjas               string
	NilaiBsun                 string
	NilaiPLH                  string
	NilaiSimdig               string
	NilaiFisika               string
	NilaiKimia                string
	NilaiDasarProgramKeahlian string
	NilaiPaketKeahlian        string
}
