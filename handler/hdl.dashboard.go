package handler

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"

	core "gitlab.com/cynomous/school001/common"
)

func DashboardPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tmpl = template.Must(template.ParseFiles(
			core.App.Config.BasePath.Join("/pages/dashboard.html"),
			core.App.Config.BasePath.Join("/views/_head.html"),
			core.App.Config.BasePath.Join("/views/_header.html"),
			core.App.Config.BasePath.Join("/views/_sidebar.html"),
		))

		data := c.MustGet("userInfo").(jwt.MapClaims)

		// binding data with template
		err := tmpl.ExecuteTemplate(c.Writer, "dashboard", gin.H{
			"Funcs": Funcs{
				IsActiveNavLink: IsActiveNavLink(),
			},
			"path":     c.Request.URL.Path,
			"name":     data["Name"],
			"username": data["Username"],
			"allowed":  CheckAllowedPath(fmt.Sprint(data["Role"]), c.Request.URL.Path),
		})
		if err != nil {
			core.App.Logger.Zap.Error("cannot call page", zap.Error(err))
			http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		}
	}
}
