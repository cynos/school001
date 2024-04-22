package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"

	core "gitlab.com/cynomous/school001/common"
	"gitlab.com/cynomous/school001/modules/report"
	table "gitlab.com/cynomous/school001/modules/tables"
)

func LoginPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tmpl = template.Must(template.ParseFiles(
			core.App.Config.BasePath.Join("/pages/login.html"),
			core.App.Config.BasePath.Join("/views/_jsscript.html"),
		))
		err := tmpl.ExecuteTemplate(c.Writer, "login", nil)
		if err != nil {
			core.App.Logger.Zap.Error("cannot call page", zap.Error(err))
			http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		}
		// counter hit
		core.App.Report.Counter(report.Increment, &report.Where{"report_type": c.Request.URL.Path}, 1, "hit")
		c.Header("REQUIRES-AUTH", "1")
	}
}

func LoginProcess() gin.HandlerFunc {
	return func(c *gin.Context) {
		inc := c.MustGet("incoming").(*table.Incoming)
		inc.Save = true
		inc.Track = TrackLogin
		inc.HasResponse = true

		if c.Request.Method != "POST" {
			Response(c, 400, false, "Unsupported http method", nil)
			return
		}

		username, password, ok := c.Request.BasicAuth()
		if !ok {
			Response(c, 400, false, "Invalid username or password", nil)
			return
		}

		valid, users := AuthenticateUser(username, password)
		if !valid {
			Response(c, 400, false, "Invalid username or password", nil)
			return
		}

		// check token if not yet expired
		lastToken, err := jwt.Parse(users.Token, func(token *jwt.Token) (interface{}, error) {
			if method, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("signing method invalid")
			} else if method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("signing method invalid")
			}
			return []byte(core.App.Config.JWT.SigKey), nil
		})
		// return saved token
		if err == nil && lastToken.Valid {
			// set default cookies to 1 hour
			err = core.App.SecureCookie["auth"].SetValue(c.Writer, []byte(users.Token))
			// err = SetCookie(c, "token", users.Token, time.Now().Add(time.Duration(core.App.Config.LoginExpires)*time.Minute))
			if err != nil {
				core.App.Logger.Zap.Error("cannot set cookie", zap.String("detail", err.Error()))
				Response(c, 400, false, "Internal service error", nil)
				return
			}
			Response(c, 200, true, "Login Success", nil)
			return
		}

		var expires *jwt.NumericDate
		if rememberMe, _ := strconv.ParseBool(c.PostForm("rememberMe")); !rememberMe {
			// default expiration jwt = 1 week
			expires = jwt.NewNumericDate(time.Now().Add(time.Duration(core.App.Config.JWT.TokenExpires) * time.Minute))
		} else {
			// set one month expired for jwt
			expires = jwt.NewNumericDate(time.Now().Add(time.Duration(43800) * time.Minute))
		}

		claims := struct {
			jwt.RegisteredClaims
			Name     string `json:"Name"`
			Username string `json:"Username"`
			Email    string `json:"Email"`
			Role     string `json:"Role"`
		}{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    core.App.Config.ServiceName,
				ExpiresAt: expires,
			},
			Username: users.Username,
			Email:    users.Email,
			Name:     users.Name,
			Role:     users.Role,
		}

		token := jwt.NewWithClaims(
			jwt.SigningMethodHS256,
			claims,
		)

		signedToken, err := token.SignedString([]byte(core.App.Config.JWT.SigKey))
		if err != nil {
			core.App.Logger.Zap.Error("cannot signed token", zap.String("detail", err.Error()))
			Response(c, 400, false, err.Error(), nil)
			return
		}

		core.App.DB.Model(&users).Updates(table.Users{Token: signedToken})

		// set cookie auth
		err = core.App.SecureCookie["auth"].SetValue(c.Writer, []byte(signedToken))
		// err = SetCookie(c, "token", signedToken, time.Now().Add(time.Duration(core.App.Config.LoginExpires)*time.Minute))
		if err != nil {
			core.App.Logger.Zap.Error("cannot set cookie", zap.String("detail", err.Error()))
			Response(c, 400, false, "Internal service error", nil)
			return
		}

		Response(c, 200, true, "Login Success", nil)
	}
}

func LogoutProcess() gin.HandlerFunc {
	return func(c *gin.Context) {
		inc := c.MustGet("incoming").(*table.Incoming)
		inc.Save = true
		inc.Track = TrackLogout
		core.App.SecureCookie["auth"].Delete(c.Writer)
		// ClearCookie(c, "token")
		if x := c.Query("redir"); x == "surat-kelulusan" {
			c.Redirect(http.StatusFound, "/login?redir="+x)
			return
		}
		c.Redirect(http.StatusFound, "/login")
	}
}
