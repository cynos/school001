package common

import (
	"github.com/chmike/securecookie"
	"gorm.io/gorm"

	l "gitlab.com/cynomous/school001/modules/log"
	"gitlab.com/cynomous/school001/modules/ratio"
	"gitlab.com/cynomous/school001/modules/report"
)

type Application struct {
	Config          *Config
	DB              *gorm.DB
	Logger          *l.Logger
	Report          *report.Report
	Ratio           map[string]*ratio.RatioCore
	SecureCookie    map[string]*securecookie.Obj
	SecureCookieKey []byte
	// SecureCookie *securecookie.SecureCookie
}
