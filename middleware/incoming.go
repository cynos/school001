package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	core "gitlab.com/cynomous/school001/common"
	"gitlab.com/cynomous/school001/modules/geoip"
	table "gitlab.com/cynomous/school001/modules/tables"
)

func Incoming(c *gin.Context) {
	var t = time.Now()
	var incdata = struct {
	}{}

	// get request body first before it will be consumed by other
	reqbody, _ := ioutil.ReadAll(c.Request.Body)
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(reqbody)) // rebuild request body

	// parse incoming data
	c.ShouldBind(&incdata)

	inc := table.Incoming{
		// general
		XForward:     c.Request.Header.Get("X-Forwarded-For"),
		XRealIP:      c.Request.Header.Get("X-Real-Ip"),
		Path:         c.Request.URL.Path,
		Method:       c.Request.Method,
		RequestQuery: c.Request.URL.RawQuery,
		RequestBody:  string(reqbody),
		UserAgent:    c.Request.UserAgent(),
		Headers:      fmt.Sprintf("%s", c.Request.Header),
		Callback:     nil,
	}

	// set inc ip
	if ip, err := geoip.GetIP(c.Request); err == nil {
		inc.IP = ip
	}

	// // set inc country code
	// if ipinfo, err := geoip.Search(inc.IP); err == nil {
	// 	inc.CountryCode = ipinfo.CountryCode
	// }

	c.Set("incoming", &inc)
	c.Next()

	inc.Latency = time.Since(t).String()
	inc.StatusCode = c.Writer.Status()

	if inc.HasResponse {
		switch r := c.MustGet("response").(type) {
		case gin.H:
			response, _ := json.Marshal(r)
			inc.ResponseBody = string(response)
		}
	}

	if inc.Save {
		if err := core.App.DB.Create(&inc).Error; err != nil {
			core.App.Logger.Zap.Error("cannot create incoming", zap.Error(err))
		}

		if inc.Callback != nil {
			inc.Callback(&inc)
		}
	}
}
