package handler

import (
	"html/template"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	core "gitlab.com/cynomous/school001/common"
	"gitlab.com/cynomous/school001/modules/report"
)

func init() {
	report.PriorityColumnList = []string{"revenue"}
	report.VisibleColumnList = report.VisibleColumn{
		ReportType: []string{
			"hit",
		},
		SubType: []string{
			"hit",
		},
	}
}

func ReportPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		// parse html template
		t, err := template.ParseFiles(core.App.Config.BasePath.Join("/pages/report.html"))
		if err != nil {
			c.String(404, "page not found")
			return
		}
		// binding data with template
		data := map[string]string{
			"service_name": core.App.Config.ServiceName,
		}
		err = t.Execute(c.Writer, data)
		if err != nil {
			core.App.Logger.Zap.Error("cannot call page", zap.Error(err))
		}
	}
}

func ReportPageData() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, core.App.Report.GetSummary(report.GetSummaryParams{
			FromDate: c.Query("fromdate"),
			ToDate:   c.Query("todate"),
		}))
	}
}
