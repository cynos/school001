package log

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jasonlvhit/gocron"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	Zap      *zap.Logger
	File     *os.File
	AutoFile bool // if true, automatic create and change the file output in every day
	Settings struct {
		Location *time.Location // used to naming files based on time intervals
		Dir      string
		Prefix   string
		Level    zapcore.Level
	}
	Encoder struct {
		File    zapcore.Encoder
		Console zapcore.Encoder
	}
}

// NewLogger create new logger
func NewLogger(Dir, Prefix string, Location *time.Location, Debug, AutoFile bool) *Logger {
	core := Logger{}

	var newzap zapcore.EncoderConfig

	if Debug {
		newzap = zap.NewDevelopmentEncoderConfig()
		newzap.EncodeTime = zapcore.ISO8601TimeEncoder
		core.Settings.Level = zap.DebugLevel
	} else {
		newzap = zap.NewProductionEncoderConfig()
		newzap.EncodeTime = zapcore.ISO8601TimeEncoder
		core.Settings.Level = zap.InfoLevel
	}

	core.Encoder.File = zapcore.NewJSONEncoder(newzap)
	core.Encoder.Console = zapcore.NewConsoleEncoder(newzap)
	core.Settings.Dir = Dir
	core.Settings.Prefix = Prefix
	core.Settings.Location = Location
	core.AutoFile = AutoFile
	core.File = core.getfile()

	core.Zap = core.newzap(
		zapcore.NewCore(core.Encoder.File, zapcore.AddSync(core.File), core.Settings.Level),
		zapcore.NewCore(core.Encoder.Console, zapcore.AddSync(os.Stdout), core.Settings.Level),
	)

	// auto generate file per day if in config is set true
	go core.autofile()

	return &core
}

func (c *Logger) Write(args ...interface{}) {
	log.SetOutput(c.File)
	log.Writer().Write([]byte(fmt.Sprint(args...)))
}

func (c *Logger) Print(args ...interface{}) {
	fmt.Println(args...)
}

func (c *Logger) newzap(cores ...zapcore.Core) *zap.Logger {
	return zap.New(zapcore.NewTee(cores...))
}

func (c *Logger) getfile() *os.File {
	// close previous file if already exist/called
	c.File.Close()

	file, err := os.OpenFile(fmt.Sprintf("%s/%s%s.log", c.Settings.Dir, c.Settings.Prefix, time.Now().In(c.Settings.Location).Format("2006_01_02")), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprint("error occured : ", err))
	}

	return file
}

func (c *Logger) autofile() {
	if c.AutoFile {
		task := func(c *Logger) {
			c.File = c.getfile()
			c.Zap = c.newzap(
				zapcore.NewCore(c.Encoder.File, zapcore.AddSync(c.File), c.Settings.Level),
				zapcore.NewCore(c.Encoder.Console, zapcore.AddSync(os.Stdout), c.Settings.Level),
			)
		}

		cron := gocron.NewScheduler()
		cron.Every(1).Day().At("00:01").Loc(c.Settings.Location).Do(task, c)
		<-cron.Start()
	}
}
