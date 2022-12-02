package report

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"
)

func (r Report) Counter(mode string, where *Where, value int, fields ...string) error {
	var err error
	if !strings.Contains("+-", mode) {
		err = fmt.Errorf("invalid mode %s", mode)
		r.Log.Error(err.Error())
		return err
	}

	clause := map[string]interface{}{
		"report_date": gorm.Expr("to_date(to_char(current_timestamp at time zone ?, 'YYYY-MM-DD'), 'YYYY-MM-DD')", r.TimeLocation.String()),
	}

	// get column type
	col, err := r.DB.Migrator().ColumnTypes(r.TableStruct[table_report])
	if err != nil {
		r.Log.Error(err.Error())
		return err
	}

	// checking where clause
	if where != nil {
		list := []string{}
		v := reflect.TypeOf(r.TableStruct[table_report])
		for i := 0; i < v.NumField(); i++ {
			if v.Field(i).Type.Name() == "ReportModel" {
				for ii := 0; ii < v.Field(i).Type.NumField(); ii++ {
					if v.Field(i).Type.Field(ii).Type.Kind() == reflect.String {
						list = append(list, ToUnderScore(v.Field(i).Type.Field(ii).Name))
					}
				}
			} else {
				y := v.Field(i)
				if y.Type.Kind() == reflect.String {
					list = append(list, ToUnderScore(y.Name))
				}
			}
		}
		for _, item := range list {
			v, found := where.Find(item)
			if found {
				if v == "" {
					clause[item] = "default"
				} else {
					clause[item] = v
				}
			} else {
				clause[item] = "default"
			}
		}
	} else {
		// default value of where clause
		v := reflect.TypeOf(r.TableStruct[table_report])
		for i := 0; i < v.NumField(); i++ {
			y := v.Field(i)
			if y.Type.Kind() == reflect.String {
				clause[ToUnderScore(y.Name)] = "default"
			}
		}
	}

	// checking fields
	for _, v := range fields {
		count := 0
		for _, vv := range col {
			if vv.Name() == v {
				count += 1
			}
		}
		if count < 1 {
			err = fmt.Errorf("field %v not found in table %s", v, table_report)
			r.Log.Error(err.Error())
			return err
		}
	}

	data := clause
	sm := make(map[string]interface{})
	if err := r.DB.Model(r.TableStruct[table_report]).Where(clause).Last(&sm).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		for _, field := range fields {
			if value > 0 {
				data[field] = value
			}
		}
		err = r.DB.Model(r.TableStruct[table_report]).Create(data).Error
		if err != nil {
			r.Log.Error(err.Error())
			return err
		}
		return nil
	}

	for _, field := range fields {
		// skip if the value less than 1
		if value < 1 {
			continue
		}
		// check if column type of database is not integer
		if reflect.TypeOf(sm[field]).Kind() != reflect.Int {
			continue
		}
		// prevent when decrement the value less than 0
		if mode == Decrement {
			if sm[field].(int) < 0 {
				r.Log.Info(fmt.Sprintf("decrement report failed, value of column %s cannot be minus, current value is %v", field, sm[field]))
				continue
			}
		}
		data[field] = gorm.Expr(fmt.Sprintf("%s %s %d", field, mode, value))
	}

	if err := r.DB.Model(r.TableStruct[table_report]).Where("id = ?", sm["id"]).Updates(data).Error; err != nil {
		r.Log.Error(err.Error())
		return err
	}

	return nil
}

func (r Report) CounterReplace(where *Where, fields map[string]int) error {
	clause := map[string]interface{}{
		"report_date": gorm.Expr("to_date(to_char(current_timestamp at time zone ?, 'YYYY-MM-DD'), 'YYYY-MM-DD')", r.TimeLocation.String()),
	}

	// checking where clause
	if where != nil {
		list := []string{}
		v := reflect.TypeOf(r.TableStruct[table_report])
		for i := 0; i < v.NumField(); i++ {
			y := v.Field(i)
			if y.Type.Kind() == reflect.String {
				list = append(list, ToUnderScore(y.Name))
			}
		}
		for _, item := range list {
			v, found := where.Find(item)
			if found {
				if v == "" {
					clause[item] = "default"
				} else {
					clause[item] = v
				}
			} else {
				clause[item] = "default"
			}
		}
	} else {
		// default value of where clause
		v := reflect.TypeOf(r.TableStruct[table_report])
		for i := 0; i < v.NumField(); i++ {
			y := v.Field(i)
			if y.Type.Kind() == reflect.String {
				clause[ToUnderScore(y.Name)] = "default"
			}
		}
	}

	data := clause
	sm := make(map[string]interface{})
	if err := r.DB.Model(r.TableStruct[table_report]).Where(clause).Last(&sm).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		for k, v := range fields {
			if v > 0 {
				data[k] = v
			}
		}
		err = r.DB.Model(r.TableStruct[table_report]).Create(data).Error
		if err != nil {
			r.Log.Error(err.Error())
			return err
		}
		return nil
	}

	for k, v := range fields {
		data[k] = v
	}

	if err := r.DB.Model(r.TableStruct[table_report]).Where("id = ?", sm["id"]).Updates(data).Error; err != nil {
		r.Log.Error(err.Error())
		return err
	}

	return nil
}
