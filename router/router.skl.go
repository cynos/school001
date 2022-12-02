package router

import (
	"github.com/gin-gonic/gin"

	"gitlab.com/cynomous/school001/handler"
)

func RouterSKL(rtr *gin.Engine) {
	// view
	rtr.GET("/skl", handler.SKLListPage())
	rtr.GET("/surat-kelulusan", handler.SKLResultPage())
	// data
	data := rtr.Group("/api/skl")
	{
		data.GET("/", handler.SKLList())
		data.POST("/set-email", handler.SKLSetEmail())
		data.POST("/set-manual-graduated", handler.SKLSetManualGraduated())
		data.POST("/details", handler.SKLDetails())
		data.POST("/create", handler.SKLCreate())
		data.POST("/delete", handler.SKLDelete())
		data.GET("/print", handler.SKLPrint())
	}
}
