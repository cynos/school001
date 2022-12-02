package table

import (
	"time"
)

type Incoming struct {
	ID           uint      `gorm:"primary_key"`
	CreatedAt    time.Time `gorm:"index"`
	Track        string    `gorm:"type:varchar(25);index"`
	Path         string    `gorm:"type:varchar(50)"`
	CountryCode  string    `gorm:"type:varchar(2)"`
	IP           string    `gorm:"type:varchar(50)"`
	XForward     string    `gorm:"type:varchar(50)"`
	XRealIP      string    `gorm:"type:varchar(50)"`
	Method       string    `gorm:"type:varchar(10)"`
	RequestQuery string
	RequestBody  string
	ResponseBody string
	StatusCode   int
	Latency      string
	UserAgent    string
	Headers      string

	Callback    func(*Incoming) `gorm:"-"` // only can used when save mode is true
	HasResponse bool            `gorm:"-"`
	Save        bool            `gorm:"-"`
}

func (Incoming) TableName() string {
	return "incoming"
}
