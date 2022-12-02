package report

import (
	"fmt"
	"reflect"
	"time"

	"go.uber.org/zap"
)

func (r Report) Summarizer(interval int) {
	for {
		now := time.Now().In(r.TimeLocation).Format("2006-01-02")
		sql := fmt.Sprintf("select * from %s where report_date = '%s'", table_report, now)
		rows, err := r.DB.Raw(sql).Rows()
		if err != nil {
			r.Log.Error("report date not yet found", zap.String("date", now))
			continue
		}

		data := map[string]interface{}{
			"report_date": now,
		}

		for rows.Next() {
			row := make(map[string]interface{})
			if err := r.DB.ScanRows(rows, &row); err != nil {
				r.Log.Error(err.Error())
				continue
			}
			for k, v := range row {
				if vreflect := reflect.ValueOf(v); vreflect.IsValid() {
					if k == "id" {
						continue
					}
					switch vreflect.Kind() {
					case reflect.Int, reflect.Int32, reflect.Int64:
						var sum int64
						if prev := reflect.ValueOf(data[k]); prev.IsValid() {
							sum = prev.Int() + vreflect.Int()
						} else {
							sum = vreflect.Int()
						}
						data[k] = sum
					}
				}
			}
		}

		trs := ReportSummary{}.TableName()
		exc := r.DB.Exec(fmt.Sprintf("select id from %s where report_date = '%s'", trs, now))
		if exc.RowsAffected == 0 {
			r.DB.Table(trs).Create(data)
		} else {
			r.DB.Table(trs).Where("report_date = ?", now).Updates(data)
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
}
