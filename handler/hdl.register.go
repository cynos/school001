package handler

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	core "gitlab.com/cynomous/school001/common"
	"gitlab.com/cynomous/school001/modules/report"
	table "gitlab.com/cynomous/school001/modules/tables"
)

func RegisterPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tmpl = template.Must(template.ParseFiles(
			"pages/register.html",
			"views/_jsscript.html",
		))
		err := tmpl.ExecuteTemplate(c.Writer, "register", nil)
		if err != nil {
			core.App.Logger.Zap.Error("cannot call page", zap.Error(err))
			http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		}
		// counter hit
		core.App.Report.Counter(report.Increment, &report.Where{"report_type": c.Request.URL.Path}, 1, "hit")
	}
}

func RegisterProcess() gin.HandlerFunc {
	return func(c *gin.Context) {
		inc := c.MustGet("incoming").(*table.Incoming)
		inc.Save = true
		inc.Track = TrackRegister
		inc.HasResponse = true

		if c.Request.Method != "POST" {
			Response(c, 400, false, "Unsupported http method", nil)
			return
		}

		username, password, ok := c.Request.BasicAuth()
		if !ok {
			Response(c, 400, false, "Invalid parameters, invalid usersname or password", nil)
			return
		}

		var (
			name  = c.PostForm("name")
			email = c.PostForm("email")
		)

		if username == "" || password == "" || email == "" || name == "" {
			Response(c, 400, false, "Invalid parameters", nil)
			return
		}

		_, err := GenerateUsers(name, email, username, password, ROLE_ADMIN)
		if err != nil {
			core.App.Logger.Zap.Warn("cannot create account", zap.String("detail", err.Error()))
			Response(c, 400, false, "Internal service error", nil)
			return
		}

		Response(c, 200, true, "Successfully created account", nil)
	}
}
