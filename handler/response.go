package handler

import "github.com/gin-gonic/gin"

func Response(c *gin.Context, httpCode int, event bool, message string, ext interface{}) {
	var response = gin.H{
		"event":   event,
		"message": message,
	}

	if ext != nil {
		switch data := ext.(type) {
		case map[string]interface{}:
			for key, val := range data {
				response[key] = val
			}
		case gin.H:
			for key, val := range data {
				response[key] = val
			}
		case string:
			response["ext"] = data
		}
	}

	c.JSON(httpCode, response)
	c.Set("response", response)
}

func ResponseSettings(c *gin.Context, httpstatus int, result bool, message string, ext interface{}) {
	var response = gin.H{
		"result":  result,
		"message": message,
	}

	if ext != nil {
		switch data := ext.(type) {
		case map[string]interface{}:
			for key, val := range data {
				response[key] = val
			}
		case gin.H:
			for key, val := range data {
				response[key] = val
			}
		}
	}

	c.JSON(httpstatus, response)
}
