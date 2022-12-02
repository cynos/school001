package router

import (
	"github.com/gin-gonic/gin"

	"gitlab.com/cynomous/school001/handler"
)

func RouterReport(rtr *gin.Engine) {
	rtr.GET("/report", handler.ReportPage())
	rtr.GET("/report/data", handler.ReportPageData())
}
