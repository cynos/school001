package table

import (
	"gitlab.com/cynomous/school001/modules/report"
)

type Report struct {
	report.ReportModel
	Hit int `gorm:"default:0"`
}
