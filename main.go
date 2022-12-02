package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	core "gitlab.com/cynomous/school001/common"
	"gitlab.com/cynomous/school001/modules/report"
	table "gitlab.com/cynomous/school001/modules/tables"
	"gitlab.com/cynomous/school001/router"
)

var g errgroup.Group

func main() {
	if err := core.SetupConfig(); err != nil {
		log.Fatalf("cannot setup config : %v", err)
	}

	// close db connection
	genericDB, _ := core.App.DB.DB()
	defer genericDB.Close()

	// init report
	rpt, err := report.New(core.App.DB, core.App.Logger.Zap, core.App.Config.TimeZone, false, table.Report{})
	if err != nil {
		log.Fatal(err)
	}
	core.App.Report = rpt

	// setup gin mode
	if core.App.Config.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	server01 := &http.Server{
		Addr:    core.App.Config.Port,
		Handler: router.Core(),
	}

	g.Go(func() error {
		err := server01.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			core.App.Logger.Zap.Fatal("cannot run webserver", zap.Error(err))
		}

		core.App.Logger.Zap.Info(fmt.Sprint("server running at", server01.Addr))

		quit := make(chan os.Signal, 2)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		<-quit

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server01.Shutdown(ctx); err != nil {
			core.App.Logger.Zap.Error("can't shutdown server", zap.Error(err))
		}

		core.App.Logger.Zap.Info("Server exiting")

		return err
	})

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}
