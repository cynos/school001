package common

import (
	"flag"
	"fmt"
	"time"

	"github.com/chmike/securecookie"
	"github.com/jinzhu/configor"

	l "gitlab.com/cynomous/school001/modules/log"
	table "gitlab.com/cynomous/school001/modules/tables"
)

type Config struct {
	Debug        bool
	ServiceName  string
	LogPath      string `required:"true"`
	LogName      string
	Port         string `required:"true"`
	LoginExpires int
	TimeZone     string
	TimeLocation *time.Location
	DBConfig     DBConfig
	JWT          JWTConfig
}

type JWTConfig struct {
	TokenExpires int
	SigKey       string
}

func SetupConfig() error {
	var (
		config     Config
		configFile string
		debug      bool
	)

	// variable flag
	flag.StringVar(&configFile, "config", "", "set config filename (default: empty)")
	flag.BoolVar(&debug, "debug", false, "set debug mode true/false (default: false)")
	flag.Parse()

	// set debug & api mode
	config.Debug = debug

	err := configor.New(&configor.Config{
		ErrorOnUnmatchedKeys: true,
		Debug:                debug,
		Verbose:              debug,
		AutoReload:           true,
		AutoReloadInterval:   time.Minute * 5,
	}).Load(&config, configFile)
	if err != nil {
		return err
	}

	// set time location
	config.TimeLocation, err = time.LoadLocation(config.TimeZone)
	if err != nil {
		return err
	}

	// setup database
	App.DB, err = DBConnect(config.DBConfig)
	if err != nil {
		return err
	}

	// migration/create table structure
	err = App.DB.AutoMigrate(
		table.ApiCall{},
		table.Incoming{},
		table.Settings{},
		table.Users{},
		table.Competence{},
		table.Class{},
		table.ClassMember{},
		table.SKL{},
		table.SKLDetails{},
	)
	if err != nil {
		return err
	}

	// setup logging
	App.Logger = l.NewLogger(
		config.LogPath, fmt.Sprintf("%s_", config.LogName),
		config.TimeLocation,
		config.Debug,
		true,
	)

	// setup cookie
	// App.SecureCookie = securecookie.New(
	// 	securecookie.GenerateRandomKey(32),
	// 	securecookie.GenerateRandomKey(16),
	// )

	sc := map[string]securecookie.Params{
		"auth": {
			Path:     "/",
			MaxAge:   (int(time.Minute) * config.LoginExpires) / int(time.Second),
			HTTPOnly: true,
			Secure:   false,
			SameSite: securecookie.Lax,
		},
	}
	err = setupCookies(sc)
	if err != nil {
		return err
	}

	// assign config to global var
	App.Config = &config

	return nil
}
