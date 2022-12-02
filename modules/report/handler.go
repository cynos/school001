package report

import (
	"fmt"
	"path"
	"runtime"
	"text/template"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (r Report) ReportPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		// parse html template
		_, file, _, _ := runtime.Caller(0)
		t, err := template.ParseFiles(path.Join(path.Dir(file), "report.html"))
		if err != nil {
			c.String(404, "page not found")
			return
		}

		// binding data with template
		err = t.Execute(c.Writer, nil)
		if err != nil {
			r.Log.Error("cannot call page", zap.Error(err))
		}
	}
}

type GetSummaryParams struct {
	FromDate string
	ToDate   string
}

func (r Report) GetSummary(params GetSummaryParams) gin.H {
	var data = []map[string]interface{}{}

	if params.FromDate == "" || params.ToDate == "" {
		where := fmt.Sprintf("report_date at time zone '%s' > to_date(to_char(current_timestamp at time zone '%s', 'YYYY-MM-DD'), 'YYYY-MM-DD')  - interval '31 days'", r.TimeZone, r.TimeZone)
		err := r.DB.Table(table_report).Where(where).Order("report_date desc").Find(&data).Error
		if err != nil {
			r.Log.Error(err.Error())
			return gin.H{}
		}
		for i, v := range data {
			for k, vv := range v {
				if k == "report_date" {
					dt := vv.(time.Time)
					dt.In(r.TimeLocation)
					data[i][k] = dt.Format("2006-01-02T15:04:05Z")
				}
			}
		}
	} else {
		where := fmt.Sprintf("report_date between '%s' and '%s'", params.FromDate, params.ToDate)
		err := r.DB.Table(table_report).Where(where).Order("report_date desc").Find(&data).Error
		if err != nil {
			r.Log.Error(err.Error())
			return gin.H{}
		}
	}

	if len(PriorityColumnList) < 1 {
		PriorityColumnList = []string{VisibleColumnList.ReportType[0]}
	}

	return gin.H{
		"data":                 data,
		"visible_column_list":  VisibleColumnList,
		"priority_column_list": PriorityColumnList,
	}
}
