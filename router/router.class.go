package router

import (
	"github.com/gin-gonic/gin"

	"gitlab.com/cynomous/school001/handler"
)

func RouterClass(rtr *gin.Engine) {
	// class views
	rtr.GET("/class", handler.ClassPage()) // main views for both of class and competence
	rtr.GET("/class/update", handler.ClassUpdatePage())
	// class member views
	rtr.GET("/class/member", handler.ClassMemberPage())
	// competence views
	rtr.GET("/class/competence/update", handler.CompetenceUpdatePage())
	// class data
	classData := rtr.Group("/api/class")
	{
		classData.GET("/", handler.ClassList())
		classData.POST("/create", handler.ClassCreate())
		classData.GET("/update", handler.ClassUpdate())
		classData.POST("/update", handler.ClassUpdate())
		classData.POST("/delete", handler.ClassDelete())
	}
	// class member data
	classMemberData := rtr.Group("/api/class/member")
	{
		classMemberData.GET("/", handler.ClassMemberList())
		classMemberData.POST("/create", handler.ClassMemberCreate())
		classMemberData.POST("/delete", handler.ClassMemberDelete())
		classMemberData.POST("/import", handler.ClassMemberImport())
	}
	// competence data
	competenceData := rtr.Group("/api/competence")
	{
		competenceData.GET("/", handler.CompetenceList())
		competenceData.POST("/create", handler.CompetenceCreate())
		competenceData.GET("/update", handler.ComptenceUpdate())
		competenceData.POST("/update", handler.ComptenceUpdate())
		competenceData.POST("/delete", handler.CompetenceDelete())
	}
}
