package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	core "gitlab.com/cynomous/school001/common"
	"gitlab.com/cynomous/school001/handler"
	"gitlab.com/cynomous/school001/middleware"
)

func Core() http.Handler {
	// router setup
	rtr := gin.New()

	// middleware
	rtr.Use(gin.Recovery())
	rtr.Use(middleware.Cors)
	rtr.Use(middleware.Incoming)
	rtr.Use(middleware.JWTAuthorization)

	// route static
	rtr.Static("/static", core.App.Config.BasePath.Join("/static"))
	// index
	rtr.GET("/", handler.DashboardPage())
	// signin
	rtr.GET("/login", handler.LoginPage())
	rtr.POST("/login", handler.LoginProcess())
	// signup
	rtr.GET("/register", handler.RegisterPage())
	rtr.POST("/register", handler.RegisterProcess())
	// signout
	rtr.GET("/logout", handler.LogoutProcess())

	RouterUsers(rtr)
	RouterClass(rtr)
	RouterSKL(rtr)
	RouterModules(rtr)
	RouterReport(rtr)

	return rtr
}
