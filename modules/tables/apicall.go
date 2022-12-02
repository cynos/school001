package table

import "time"

type ApiCall struct {
	ID              uint `gorm:"primarykey"`
	CreatedAt       time.Time
	Track           string `gorm:"type:varchar(50);index"`
	Msisdn          string `gorm:"type:varchar(50);index"`
	Url             string
	RequestQuery    string
	RequestBody     string
	ResponseBody    string
	Status          int
	RequestHeaders  string
	ResponseHeaders string
	Latency         string
	Error           string
	MetaData        string
}

func (ApiCall) TableName() string {
	return "apicall"
}
