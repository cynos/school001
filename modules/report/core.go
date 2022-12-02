package report

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/jasonlvhit/gocron"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

const (
	Increment    = "+"
	Decrement    = "-"
	table_report = "report"
)

type VisibleColumn struct {
	ReportType []string
	SubType    []string
}

var VisibleColumnList VisibleColumn

var PriorityColumnList []string

var mandatoryFields = []string{}

var mandatoryTables = []string{
	table_report,
}

type Where map[string]string

func (c Where) Find(key string) (string, bool) {
	for k, v := range c {
		if key == k {
			return v, true
		}
	}
	return "", false
}

type ReportModel struct {
	ID         uint      `gorm:"primary_key"`
	ReportDate time.Time `gorm:"type:date"`
	ReportType string    `gorm:"type:varchar(255)"`
	SubType    string    `gorm:"type:varchar(255)"`
}

type ReportSummary struct{}

func (ReportSummary) TableName() string {
	return "report_summary"
}

type Report struct {
	DB               *gorm.DB
	Log              *zap.Logger
	TimeLocation     *time.Location
	TimeZone         string
	AutoGenerateCron *gocron.Scheduler
	TableName        []string
	TableStruct      map[string]interface{}
}

func New(database *gorm.DB, logger *zap.Logger, timezone string, autogenerate bool, tables ...interface{}) (*Report, error) {
	report := &Report{
		DB:          database,
		Log:         logger,
		TimeZone:    timezone,
		TableStruct: make(map[string]interface{}),
	}

	// set default table fields
	for _, v := range []string{"ID", "ReportDate", "ReportType", "SubType"} {
		if _, found := FindSlice(mandatoryFields, v); !found {
			mandatoryFields = append(mandatoryFields, v)
		}
	}

	// load time location
	tl, err := time.LoadLocation(timezone)
	if err != nil {
		return report, fmt.Errorf(err.Error())
	}
	report.TimeLocation = tl

	// set temporary db config
	database.Config.NamingStrategy = schema.NamingStrategy{SingularTable: true}
	// rollback db config
	defer func() {
		database.Config.NamingStrategy = schema.NamingStrategy{}
	}()

	for tableindex, v := range tables {
		x := reflect.TypeOf(v)
		y := reflect.ValueOf(v)

		if y.Kind() == reflect.Ptr {
			v = y.Elem().Interface()
			x = reflect.TypeOf(v)
			y = reflect.ValueOf(v)
		}

		invoke_res, err := Invoke(v, "TableName")
		if err == nil && len(invoke_res) > 0 {
			name := invoke_res[0].String()
			report.TableName = append(report.TableName, name)
		} else {
			name := database.Config.NamingStrategy.TableName(x.Name())
			report.TableName = append(report.TableName, name)
		}

		fieldNames := []string{}
		for i := 0; i < x.NumField(); i++ {
			if x.Field(i).Type.Name() == "ReportModel" {
				for ii := 0; ii < x.Field(i).Type.NumField(); ii++ {
					fieldNames = append(fieldNames, x.Field(i).Type.Field(ii).Name)
				}
			} else {
				fieldNames = append(fieldNames, x.Field(i).Name)
			}
		}

		// checking mandatory fields
		if len(mandatoryFields) < 1 {
			return report, fmt.Errorf("mandatory fields cannot be empty")
		}
		for _, v := range mandatoryFields {
			if _, found := FindSlice(fieldNames, v); !found {
				return report, fmt.Errorf("%s field not found in this struct %s", v, x.Name())
			}
		}

		report.TableStruct[report.TableName[tableindex]] = v
	}

	// checking mandatory tables
	if len(mandatoryTables) < 1 {
		return report, fmt.Errorf("mandatory tables cannot be empty")
	}
	for _, v := range mandatoryTables {
		if _, found := FindSlice(report.TableName, v); !found {
			return report, fmt.Errorf("mandatory table must be present in the list [%s]", strings.Join(mandatoryTables, ", "))
		}
	}

	// migrate table
	for _, v := range report.TableStruct {
		database.AutoMigrate(v)
	}

	// create table report summary
	if database.Migrator().HasTable(table_report) {
		trs := ReportSummary{}.TableName()
		database.Exec(fmt.Sprintf(
			`create table %s (
				like %s
				including defaults
				including indexes
				including constraints
			);`, trs, table_report,
		))
		database.Exec(fmt.Sprintf(`create sequence %s_id_seq;`, trs))
		database.Exec(fmt.Sprintf(`alter table %s alter column id set default nextval('%s_id_seq'::regclass);`, trs, trs))
		database.Exec(fmt.Sprintf(`alter sequence %s_id_seq owned by %s.id`, trs, trs))
	}

	// auto generate data of table report everyday
	if autogenerate {
		go autoGenerate(report)
	}

	go report.Summarizer(30)

	return report, nil
}

func autoGenerate(r *Report) {
	cron := gocron.NewScheduler()
	cron.Every(1).Day().At("00:05").Loc(r.TimeLocation).From(gocron.NextTick()).Do(func() {
		now := time.Now().In(r.TimeLocation).Format("2006-01-02")
		sql := fmt.Sprintf("select * from %s where report_date = '%s'", table_report, now)
		exc := r.DB.Exec(sql)
		if exc.RowsAffected == 0 {
			var column, value []string
			v := reflect.TypeOf(r.TableStruct[table_report])
			for i := 0; i < v.NumField(); i++ {
				y := v.Field(i)
				if y.Type.Kind() == reflect.Struct {
					for ii := 0; ii < y.Type.NumField(); ii++ {
						if y.Type.Field(ii).Type.Kind() == reflect.String {
							column = append(column, ToUnderScore(y.Type.Field(ii).Name))
							value = append(value, "'default'")
						}
					}
				} else {
					if y.Type.Kind() == reflect.String {
						column = append(column, ToUnderScore(y.Name))
						value = append(value, "'default'")
					}
				}
			}
			sql = fmt.Sprintf(
				"insert into %s (report_date, %s) values ('%s', %s)",
				table_report,
				strings.Join(column, ", "),
				time.Now().In(r.TimeLocation).Format("2006-01-02"),
				strings.Join(value, ", "),
			)
			r.DB.Exec(sql)
		}
	})
	r.AutoGenerateCron = cron
	<-r.AutoGenerateCron.Start()
}

func SetMandatoryFiedls(field string) {
	if _, found := FindSlice(mandatoryFields, field); !found {
		mandatoryFields = append(mandatoryFields, field)
	}
}
