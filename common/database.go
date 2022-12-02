package common

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DBConfig struct {
	Host            string `required:"true"`
	Port            string `required:"true"`
	Name            string `required:"true"`
	User            string `required:"true"`
	Password        string `required:"true"`
	ApplicationName string `required:"true"`
	ConnectTimeout  int    `default:"30"`
	MaxOpenConn     int    `default:"50"`
	MaxIdleConn     int    `default:"10"`
}

func (c *DBConfig) dbinfo() string {
	return fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s connect_timeout=%d application_name=%s", c.Host, c.Port, c.User, c.Name, c.Password, c.ConnectTimeout, c.ApplicationName)
}

func DBConnect(config DBConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(config.dbinfo()), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(config.MaxIdleConn)
	sqlDB.SetMaxOpenConns(config.MaxOpenConn)

	return db, nil
}
