package router

import (
	"github.com/gin-gonic/gin"

	"gitlab.com/cynomous/school001/handler"
)

func RouterUsers(rtr *gin.Engine) {
	// view
	rtr.GET("/users", handler.UsersPage())
	rtr.GET("/users/update", handler.UsersUpdatePage())
	// data
	data := rtr.Group("/api/users")
	{
		data.GET("/", handler.UsersList())          // get user list data
		data.GET("/update", handler.UsersUpdate())  // get update data
		data.POST("/update", handler.UsersUpdate()) // save update data
		data.POST("/create", handler.UsersCreate()) // create data
		data.POST("/delete", handler.UsersDelete()) // delete user data
		data.POST("/import", handler.UsersImport()) // delete user data
	}
}
