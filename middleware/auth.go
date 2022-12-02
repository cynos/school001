package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"

	core "gitlab.com/cynomous/school001/common"
)

func JWTAuthorization(c *gin.Context) {
	if c.Request.URL.Path == "/login" || c.Request.URL.Path == "/register" || strings.HasPrefix(c.Request.URL.Path, "/static") {
		switch c.Request.URL.Path {
		case "/login":
			if _, err := core.App.SecureCookie["auth"].GetValue(nil, c.Request); err == nil {
				c.Abort()
				c.Redirect(http.StatusFound, "/")
				return
			}

			// if _, err := handler.GetCookie(c, "token"); err == nil {
			// 	c.Abort()
			// 	c.Redirect(http.StatusFound, "/")
			// 	return
			// }
		}
		c.Next()
		return
	}

	tokenString, err := core.App.SecureCookie["auth"].GetValue(nil, c.Request)
	// tokenString, err := handler.GetCookie(c, "token")
	if err != nil {
		core.App.SecureCookie["auth"].Delete(c.Writer)
		// handler.ClearCookie(c, "token")
		c.Abort()
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// prolongation expires cookie
	core.ReplenishExpiredCookie(
		"auth",
		(int(time.Minute)*core.App.Config.LoginExpires)/int(time.Second),
	).SetValue(c.Writer, tokenString)
	// handler.SetCookie(c, "token", tokenString, time.Now().Add(time.Duration(core.App.Config.LoginExpires)*time.Minute))

	token, err := jwt.Parse(string(tokenString), func(token *jwt.Token) (interface{}, error) {
		if method, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("signing method invalid")
		} else if method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("signing method invalid")
		}
		return []byte(core.App.Config.JWT.SigKey), nil
	})
	if err != nil {
		core.App.SecureCookie["auth"].Delete(c.Writer)
		// handler.ClearCookie(c, "token")
		c.Abort()
		c.Redirect(http.StatusFound, "/login")
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		core.App.SecureCookie["auth"].Delete(c.Writer)
		// handler.ClearCookie(c, "token")
		c.Abort()
		c.Redirect(http.StatusFound, "/login")
		return
	}

	c.Set("userInfo", claims)
}
