package handler

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"

	core "gitlab.com/cynomous/school001/common"
	table "gitlab.com/cynomous/school001/modules/tables"
	"gitlab.com/cynomous/school001/modules/tools"
)

func ClassMemberPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tmpl = template.Must(template.ParseFiles(
			core.App.Config.BasePath.Join("/pages/class_member.html"),
			core.App.Config.BasePath.Join("/views/_head.html"),
			core.App.Config.BasePath.Join("/views/_header.html"),
			core.App.Config.BasePath.Join("/views/_sidebar.html"),
			core.App.Config.BasePath.Join("/views/_pagenav.html"),
			core.App.Config.BasePath.Join("/views/_modal.html"),
			core.App.Config.BasePath.Join("/views/_jsscript.html"),
		))

		data := c.MustGet("userInfo").(jwt.MapClaims)

		// binding data with template
		err := tmpl.ExecuteTemplate(c.Writer, "class_member", gin.H{
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

// class logic data ================================================================================================

func ClassMemberList() gin.HandlerFunc {
	reverse := func(x []table.ClassMember) []table.ClassMember {
		for i := 0; i < len(x)/2; i++ {
			j := len(x) - i - 1
			x[i], x[j] = x[j], x[i]
		}
		return x
	}

	return func(c *gin.Context) {
		var (
			id           = tools.StringToInt(c.Query("id"))
			limit        = c.DefaultQuery("limit", "15")
			cursor       = c.DefaultQuery("cursor", "0")
			direction    = c.Query("direction")
			search       = c.Query("search")
			classMembers = []table.ClassMember{}
			order        string
			wherePaging  string
			whereSearch  string
			next         uint
			prev         uint
			first        uint
			last         uint
		)

		// get first and last id
		{
			var s string
			if search != "" {
				s = fmt.Sprintf(`and c2."name" like '%%%s%%' or c2.username like '%%%s%%'`, search, search)
			}
			err := core.App.DB.Raw(fmt.Sprintf(`
			select
				(select c1.id from class_members c1 inner join users c2 on c2.username = c1.users_username where c1.class_id = ? %s order by c1.id asc limit 1),
				(select c1.id from class_members c1 inner join users c2 on c2.username = c1.users_username where c1.class_id = ? %s order by c1.id desc limit 1)
			`, s, s), id, id).Row().Scan(&first, &last)
			if err != nil {
				core.App.Logger.Zap.Error(err.Error())
				Response(c, 200, true, "data not found", gin.H{"data": nil, "error": err.Error()})
				return
			}
		}

		switch direction {
		case "next":
			order = "c1.id DESC"
			wherePaging = fmt.Sprintf("AND c1.id < %d", tools.StringToInt(cursor))
		case "prev":
			order = "c1.id ASC"
			wherePaging = fmt.Sprintf("AND c1.id > %d", tools.StringToInt(cursor))
		default:
			order = "c1.id DESC"
			wherePaging = fmt.Sprintf("AND c1.id < %d", tools.StringToInt(cursor))
		}

		if cursor == "0" {
			wherePaging = ""
		}

		if search != "" {
			whereSearch = fmt.Sprintf(`AND c2."name" like '%%%s%%' or c2.username like '%%%s%%'`, search, search)
		}

		raw := fmt.Sprintf(`
		SELECT
			c1.id, c1.class_id, c1.users_username, c1.created_at,
			c2."name", c2.username,
			c3."name", c3."name_id",
			c4."name", c4."name_id"
		FROM class_members c1
		INNER JOIN users c2 ON c2.username = c1.users_username
		INNER JOIN "class" c3 ON c3.id = c1.class_id
		INNER JOIN competence c4 ON c4.name_id = c3.competence_name_id 
		WHERE (c1.class_id = %d %s) %s
		ORDER BY %s
		LIMIT %d;
		`, id, wherePaging, whereSearch, order, tools.StringToInt(limit))

		// err := query.
		// 	Preload("Users", "users.name like ?", fmt.Sprintf("%%%s%%", search)).
		// 	Preload("Class").
		// 	Preload("Class.Competence").
		// 	Where("class_id = ?", id).
		// 	Order(order).
		// 	Find(&classMembers).
		// 	Error

		rows, err := core.App.DB.Raw(raw).Rows()
		if err != nil {
			Response(c, 200, true, "data not found", gin.H{"data": nil, "error": err.Error()})
			return
		}

		for rows.Next() {
			row := table.ClassMember{}
			if err := rows.Scan(
				&row.ID, &row.ClassID, &row.UsersUsername, &row.CreatedAt,
				&row.Users.Name, &row.Users.Username,
				&row.Class.Name, &row.Class.NameID,
				&row.Class.Competence.Name, &row.Class.Competence.NameID,
			); err != nil {
				core.App.Logger.Zap.Error("error scan column", zap.String("details", err.Error()))
				continue
			}
			classMembers = append(classMembers, row)
		}

		if len(classMembers) < 1 {
			Response(c, 200, true, "data not found", gin.H{"data": nil, "error": ""})
			return
		}

		switch direction {
		case "next":
			next = classMembers[len(classMembers)-1].ID
			prev = classMembers[0].ID
		case "prev":
			classMembers = reverse(classMembers)
			next = classMembers[len(classMembers)-1].ID
			prev = classMembers[0].ID
		default:
			next = classMembers[len(classMembers)-1].ID
			prev = classMembers[0].ID
		}

		data := []map[string]interface{}{}
		for _, e := range classMembers {
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

func ClassMemberDelete() gin.HandlerFunc {
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

		query := core.App.DB.Unscoped().Where(fmt.Sprintf("id in(%s)", id)).Delete(&table.ClassMember{})
		err := query.Error
		if err != nil {
			Response(c, 200, false, "Cannot delete data", gin.H{"error": err.Error()})
			return
		}

		Response(c, 200, true, "Delete Succeed", nil)
	}
}

func ClassMemberCreate() gin.HandlerFunc {
	return func(c *gin.Context) {
		inc := c.MustGet("incoming").(*table.Incoming)
		inc.Save = true
		inc.HasResponse = true

		data := make(map[string]interface{})
		if err := json.Unmarshal([]byte(inc.RequestBody), &data); err != nil {
			Response(c, 200, false, "Cannot parsing data", gin.H{"error": err.Error()})
			return
		}

		classMembers := []table.ClassMember{}
		for _, v := range data {
			row := v.(map[string]interface{})
			var (
				username = fmt.Sprint(row["username"])
				classID  = fmt.Sprint(row["classID"])
			)
			if username == "" || classID == "" {
				continue
			}
			classMembers = append(classMembers, table.ClassMember{
				UsersUsername: username,
				ClassID:       uint(tools.StringToInt(classID)),
			})
		}

		if len(classMembers) < 1 {
			Response(c, 200, false, "Canno create empty data", nil)
			return
		}

		if err := core.App.DB.Create(&classMembers).Error; err != nil {
			Response(c, 200, false, "Cannot create data", gin.H{"error": err.Error()})
			return
		}

		Response(c, 200, true, "Success create data", nil)
	}
}

func ClassMemberImport() gin.HandlerFunc {
	return func(c *gin.Context) {
		file, _, err := c.Request.FormFile("file")
		if err != nil {
			Response(c, 400, false, "cannot get file", gin.H{"error": err.Error()})
			return
		}
		defer file.Close()

		reader := csv.NewReader(file)
		reader.FieldsPerRecord = -1

		csvData, err := reader.ReadAll()
		if err != nil {
			Response(c, 400, false, "cannot read file", gin.H{"error": err.Error()})
			return
		}

		rows := []map[string]interface{}{}
		for i, each := range csvData {
			if i == 0 {
				if each[0] != "Username" {
					Response(c, 400, false, "invalid header name for csv file", nil)
					return
				}
			} else {
				row := map[string]interface{}{
					"Username": each[0],
				}
				rows = append(rows, row)
			}
		}

		Response(c, 200, true, "success convert", gin.H{"data": rows})
	}
}
