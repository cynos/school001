package router

import (
	"github.com/gin-gonic/gin"

	"gitlab.com/cynomous/school001/handler"
)

func RouterModules(rtr *gin.Engine) {
	// view
	rtr.GET("/modules", handler.ModulesPage())
	// data
	data := rtr.Group("/api/modules")
	{
		data.GET("/")
		data.POST("/create")
		data.POST("/update")
		data.POST("/delete")
	}
}
