package handler

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	core "gitlab.com/cynomous/school001/common"
	table "gitlab.com/cynomous/school001/modules/tables"
)

func SettingsHandler() gin.HandlerFunc {
	settingsEventGet := func(c *gin.Context) {
		inc := c.MustGet("incoming").(*table.Incoming)

		params := struct {
			Name string `json:"name"`
		}{}
		err := json.Unmarshal([]byte(inc.RequestBody), &params)
		if err != nil {
			core.App.Logger.Zap.Warn(err.Error())
			ResponseSettings(c, 400, false, "cannot parse request body", nil)
			return
		}

		setting := table.Settings{}
		err = core.App.DB.Where("name = ?", params.Name).Last(&setting).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ResponseSettings(c, 400, false, "setting not found", nil)
				return
			}
			core.App.Logger.Zap.Warn(err.Error())
			ResponseSettings(c, 400, false, fmt.Sprintf("internal system error, detail:%s", err.Error()), nil)
			return
		}

		ResponseSettings(c, 200, true, "setting successfully executed", gin.H{
			"data": setting,
		})
	}

	settingsEventSet := func(c *gin.Context) {
		inc := c.MustGet("incoming").(*table.Incoming)

		params := struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		}{}
		err := json.Unmarshal([]byte(inc.RequestBody), &params)
		if err != nil {
			core.App.Logger.Zap.Warn(err.Error())
			ResponseSettings(c, 400, false, "cannot parse request body", nil)
			return
		}

		setting := table.Settings{}
		err = core.App.DB.Where("name = ?", params.Name).Last(&setting).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = core.App.DB.Model(setting).Create(map[string]interface{}{
					"name":  params.Name,
					"value": params.Value,
				}).Error
				if err != nil {
					ResponseSettings(c, 400, false, fmt.Sprintf("internal system error, detail:%s", err.Error()), nil)
					return
				}
				ResponseSettings(c, 200, true, "setting successfully executed", nil)
				return
			}
			core.App.Logger.Zap.Warn(err.Error())
			ResponseSettings(c, 400, false, fmt.Sprintf("internal system error, detail:%s", err.Error()), nil)
			return
		}

		err = core.App.DB.Model(&setting).Update("value", params.Value).Error
		if err != nil {
			core.App.Logger.Zap.Warn(err.Error())
			ResponseSettings(c, 400, false, fmt.Sprintf("internal system error, detail:%s", err.Error()), nil)
			return
		}

		ResponseSettings(c, 200, true, "setting successfully executed", nil)
	}

	return func(c *gin.Context) {
		event := c.Param("event")
		switch event {
		case "set":
			settingsEventSet(c)
		case "get":
			settingsEventGet(c)
		default:
			ResponseSettings(c, 400, false, "invalid parameter", nil)
		}
	}
}
