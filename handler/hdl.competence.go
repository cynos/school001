package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"

	core "gitlab.com/cynomous/school001/common"
	table "gitlab.com/cynomous/school001/modules/tables"
	"gitlab.com/cynomous/school001/modules/tools"
)

func CompetenceUpdatePage() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tmpl = template.Must(template.ParseFiles(
			core.App.Config.BasePath.Join("/pages/competence_update.html"),
			core.App.Config.BasePath.Join("/views/_head.html"),
			core.App.Config.BasePath.Join("/views/_header.html"),
			core.App.Config.BasePath.Join("/views/_sidebar.html"),
			core.App.Config.BasePath.Join("/views/_jsscript.html"),
			core.App.Config.BasePath.Join("/views/_modal.html"),
		))

		data := c.MustGet("userInfo").(jwt.MapClaims)

		// binding data with template
		err := tmpl.ExecuteTemplate(c.Writer, "competence_update", gin.H{
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

// competence logic data ================================================================================================

func CompetenceCreate() gin.HandlerFunc {
	return func(c *gin.Context) {
		inc := c.MustGet("incoming").(*table.Incoming)
		inc.Save = true
		inc.HasResponse = true

		data := make(map[string]interface{})
		if err := json.Unmarshal([]byte(inc.RequestBody), &data); err != nil {
			Response(c, 200, false, "Cannot parsing data", gin.H{"error": err.Error()})
			return
		}

		competenceData := []table.Competence{}
		for _, v := range data {
			row := v.(map[string]interface{})
			var competenceName = fmt.Sprint(row["competenceName"])
			var competenceNameID = fmt.Sprint(row["competenceNameID"])
			if competenceName == "" || competenceNameID == "" {
				continue
			}
			competenceData = append(competenceData, table.Competence{
				Name:   competenceName,
				NameID: competenceNameID,
			})
		}

		if len(competenceData) < 1 {
			Response(c, 200, false, "Canno create empty data", nil)
			return
		}

		if err := core.App.DB.Create(&competenceData).Error; err != nil {
			Response(c, 200, false, "Cannot create data", gin.H{"error": err.Error()})
			return
		}

		Response(c, 200, true, "Success create data", nil)
	}
}

func CompetenceList() gin.HandlerFunc {
	reverse := func(x []table.Competence) []table.Competence {
		for i := 0; i < len(x)/2; i++ {
			j := len(x) - i - 1
			x[i], x[j] = x[j], x[i]
		}
		return x
	}

	return func(c *gin.Context) {
		var (
			limit       = c.DefaultQuery("limit", "15")
			cursor      = c.DefaultQuery("cursor", "0")
			direction   = c.Query("direction")
			search      = c.Query("search")
			competences = []table.Competence{}
			order       string
			where       string
			next        uint
			prev        uint
			first       uint
			last        uint
		)

		// get first and last id
		{
			tableName := table.Competence{}.TableName()
			var s string
			if search != "" {
				s = fmt.Sprintf("where name like '%%%s%%' or name_id like '%%%s%%'", search, search)
			}
			err := core.App.DB.Raw(fmt.Sprintf(`
			select
				(select id from %s %s order by id asc limit 1),
				(select id from %s %s order by id desc limit 1)
			`, tableName, s, tableName, s)).Row().Scan(&first, &last)
			if err != nil {
				core.App.Logger.Zap.Error(err.Error())
				Response(c, 200, true, "data not found", gin.H{"data": nil, "error": err.Error()})
				return
			}
		}

		switch direction {
		case "next":
			order = "id desc"
			where = "id < ?"
		case "prev":
			order = "id asc"
			where = "id > ?"
		default:
			order = "id desc"
			where = "id < ?"
		}

		query := core.App.DB.Order(order).Limit(tools.StringToInt(limit))
		if cursor != "0" {
			query.Where(where, tools.StringToInt(cursor))
		}
		if search != "" {
			query.Where("name like ? or name_id like ?", fmt.Sprintf("%%%s%%", search), fmt.Sprintf("%%%s%%", search))
		}
		err := query.Find(&competences).Error
		if err != nil || errors.Is(err, gorm.ErrRecordNotFound) {
			Response(c, 200, true, "data not found", gin.H{"data": nil, "error": err.Error()})
			return
		}

		if len(competences) < 1 {
			Response(c, 200, true, "data not found", gin.H{"data": nil, "error": ""})
			return
		}

		switch direction {
		case "next":
			next = competences[len(competences)-1].ID
			prev = competences[0].ID
		case "prev":
			competences = reverse(competences)
			next = competences[len(competences)-1].ID
			prev = competences[0].ID
		default:
			next = competences[len(competences)-1].ID
			prev = competences[0].ID
		}

		data := []map[string]interface{}{}
		for _, e := range competences {
			row := make(map[string]interface{})
			tools.ForEachStruct(e, func(key string, val interface{}) {
				if (tools.Array{"CreatedAt", "UpdatedAt"}).Includes(key) {
					row[key] = val.(time.Time).Format("2006-01-02 15:04:05")
				} else {
					row[key] = val
				}
			})
			data = append(data, row)
		}

		Response(c, 200, true, "data found", gin.H{
			"data":  data,
			"next":  next,
			"prev":  prev,
			"first": first,
			"last":  last,
			"error": nil,
		})
	}
}

func CompetenceDelete() gin.HandlerFunc {
	return func(c *gin.Context) {
		inc := c.MustGet("incoming").(*table.Incoming)
		inc.Save = true
		inc.HasResponse = true

		id := c.PostForm("ids")
		idArr := strings.Split(id, ",")

		if len(idArr) < 1 {
			Response(c, 200, false, "empty id", nil)
			return
		}

		query := core.App.DB.Unscoped().Where(fmt.Sprintf("id in(%s)", id)).Delete(&table.Competence{})
		// stmt := query.Session(&gorm.Session{DryRun: true}).Statement
		// x := stmt.SQL.String()
		// y := stmt.Vars
		// fmt.Println(core.App.DB.Dialector.Explain(x, y))
		err := query.Error
		if err != nil {
			Response(c, 200, false, "Cannot delete data", gin.H{"error": err.Error()})
			return
		}

		Response(c, 200, true, "Delete Succeed", nil)
	}
}

func ComptenceUpdate() gin.HandlerFunc {
	getData := func(c *gin.Context) {
		var (
			id         = c.Query("id")
			competence = table.Competence{}
		)

		err := core.App.DB.Where("id = ?", tools.StringToInt(id)).Last(&competence).Error
		if err != nil || errors.Is(err, gorm.ErrRecordNotFound) {
			Response(c, 200, false, "record not found", nil)
			return
		}

		Response(c, 200, true, "record found", gin.H{"data": competence})
	}

	saveData := func(c *gin.Context) {
		var (
			id     = c.PostForm("id")
			name   = c.PostForm("name")
			nameID = c.PostForm("name_id")
		)

		if name == "" || nameID == "" {
			Response(c, 200, false, "Invalid parameters", nil)
			return
		}

		x := table.Competence{
			Name:   name,
			NameID: nameID,
		}
		err := core.App.DB.Model(&table.Competence{}).Where("id = ?", id).Updates(x).Error
		if err != nil {
			core.App.Logger.Zap.Warn("cannot update data", zap.String("detail", err.Error()))
			Response(c, 200, false, "Failed update data", gin.H{"error": err.Error()})
			return
		}

		data := map[string]interface{}{}
		tools.ForEachStruct(x, func(key string, val interface{}) {
			if (tools.Array{"CreatedAt", "UpdatedAt"}).Includes(key) {
				data[key] = val.(time.Time).Format("2006-01-02 15:04:05")
			} else {
				data[key] = val
			}
		})

		Response(c, 200, true, "Success update data", gin.H{"data": data})
	}

	return func(c *gin.Context) {
		switch c.Request.Method {
		case "GET":
			getData(c)
		case "POST":
			saveData(c)
		default:
			Response(c, 200, false, "Method not supported", nil)
		}
	}
}
