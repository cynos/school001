package handler

import (
	"encoding/csv"
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

func UsersPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tmpl = template.Must(template.ParseFiles(
			"pages/users.html",
			"views/_head.html",
			"views/_header.html",
			"views/_sidebar.html",
			"views/_pagenav.html",
			"views/_modal.html",
			"views/_jsscript.html",
		))

		data := c.MustGet("userInfo").(jwt.MapClaims)

		// binding data with template
		err := tmpl.ExecuteTemplate(c.Writer, "users", gin.H{
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

func UsersUpdatePage() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tmpl = template.Must(template.ParseFiles(
			"pages/users_update.html",
			"views/_head.html",
			"views/_header.html",
			"views/_sidebar.html",
			"views/_jsscript.html",
			"views/_modal.html",
		))

		data := c.MustGet("userInfo").(jwt.MapClaims)

		// binding data with template
		err := tmpl.ExecuteTemplate(c.Writer, "users_update", gin.H{
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

// logic data ================================================================================================

func UsersList() gin.HandlerFunc {
	reverse := func(x []table.Users) []table.Users {
		for i := 0; i < len(x)/2; i++ {
			j := len(x) - i - 1
			x[i], x[j] = x[j], x[i]
		}
		return x
	}

	return func(c *gin.Context) {
		var (
			limit     = c.DefaultQuery("limit", "15")
			cursor    = c.DefaultQuery("cursor", "0")
			direction = c.Query("direction")
			search    = c.Query("search")
			users     = []table.Users{}
			order     string
			where     string
			next      uint
			prev      uint
			first     uint
			last      uint
		)

		// get first and last id
		{
			tableName := table.Users{}.TableName()
			var s string
			if search != "" {
				s = fmt.Sprintf("where name like '%%%s%%' or email like '%%%s%%'", search, search)
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
			query.Where("name like ? or email like ?", fmt.Sprintf("%%%s%%", search), fmt.Sprintf("%%%s%%", search))
		}
		err := query.Find(&users).Error
		if err != nil || errors.Is(err, gorm.ErrRecordNotFound) {
			Response(c, 200, true, "data not found", gin.H{"data": nil, "error": err.Error()})
			return
		}

		// stmt := query.Session(&gorm.Session{DryRun: true}).Find(&users).Statement
		// fmt.Println(core.App.DB.Explain(stmt.SQL.String(), stmt.Vars...))

		if len(users) < 1 {
			Response(c, 200, true, "data not found", gin.H{"data": nil, "error": ""})
			return
		}

		switch direction {
		case "next":
			next = users[len(users)-1].ID
			prev = users[0].ID
		case "prev":
			users = reverse(users)
			next = users[len(users)-1].ID
			prev = users[0].ID
		default:
			next = users[len(users)-1].ID
			prev = users[0].ID
		}

		data := []map[string]interface{}{}
		for _, e := range users {
			row := make(map[string]interface{})
			tools.ForEachStruct(e, func(key string, val interface{}) {
				excluded := tools.Array{"Token", "PasswordHash"}
				if !excluded.Includes(key) {
					row[key] = val
				}
				if key == "LastLogin" {
					row[key] = val.(time.Time).Format("2006-01-02 15:04:05")
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

func UsersDelete() gin.HandlerFunc {
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

		query := core.App.DB.Unscoped().Where(fmt.Sprintf("id in(%s)", id)).Delete(&table.Users{})
		stmt := query.Session(&gorm.Session{DryRun: true}).Statement
		x := stmt.SQL.String()
		y := stmt.Vars
		fmt.Println(core.App.DB.Dialector.Explain(x, y))
		err := query.Error
		if err != nil {
			Response(c, 200, false, "Cannot delete data", gin.H{"error": err.Error()})
			return
		}

		Response(c, 200, true, "Delete Succeed", nil)
	}
}

func UsersUpdate() gin.HandlerFunc {
	getData := func(c *gin.Context) {
		var (
			id   = c.Query("id")
			user = table.Users{}
		)

		err := core.App.DB.Where("id = ?", tools.StringToInt(id)).Last(&user).Error
		if err != nil || errors.Is(err, gorm.ErrRecordNotFound) {
			Response(c, 200, false, "record not found", nil)
			return
		}

		Response(c, 200, true, "record found", gin.H{"data": user})
	}

	saveData := func(c *gin.Context) {
		var (
			id    = c.PostForm("id")
			name  = c.PostForm("name")
			email = c.PostForm("email")
			role  = c.PostForm("role")
		)

		if name == "" || email == "" || role == "" {
			Response(c, 200, false, "Invalid parameters", nil)
			return
		}

		username, password, ok := c.Request.BasicAuth()
		if !ok {
			Response(c, 200, false, "Invalid username or password", nil)
			return
		}

		user, err := UpdateUsers(id, name, email, username, password, role)
		if err != nil {
			core.App.Logger.Zap.Warn("cannot update data", zap.String("detail", err.Error()))
			Response(c, 200, false, "Failed update data", gin.H{"error": err.Error()})
			return
		}

		data := map[string]interface{}{}
		tools.ForEachStruct(user, func(key string, val interface{}) {
			excluded := tools.Array{"Token", "PasswordHash", "LastLogin"}
			if !excluded.Includes(key) {
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

func UsersCreate() gin.HandlerFunc {
	return func(c *gin.Context) {
		inc := c.MustGet("incoming").(*table.Incoming)
		inc.Save = true
		inc.HasResponse = true

		data := make(map[string]interface{})
		if err := json.Unmarshal([]byte(inc.RequestBody), &data); err != nil {
			Response(c, 200, false, "Cannot parsing data", gin.H{"error": err.Error()})
			return
		}

		if len(data) < 1 {
			Response(c, 200, false, "Cannot create empty data", gin.H{"error": nil})
			return
		}

		countFail := 0
		countSuccess := 0

		for _, v := range data {
			row := v.(map[string]interface{})
			var (
				inputName     = fmt.Sprint(row["name"])
				inputEmail    = fmt.Sprint(row["email"])
				inputUsername = fmt.Sprint(row["username"])
				inputPassword = fmt.Sprint(row["password"])
				inputRole     = fmt.Sprint(row["role"])
			)
			if inputName == "" || inputEmail == "" || inputUsername == "" || inputPassword == "" || inputRole == "" {
				countFail++
				continue
			}

			_, err := GenerateUsers(inputName, inputEmail, inputUsername, inputPassword, inputRole)
			if err != nil {
				core.App.Logger.Zap.Error("cannot generate users", zap.String("details", err.Error()))
				countFail++
				continue
			}

			countSuccess++
		}

		if countFail > 0 && countSuccess > 0 {
			Response(c, 200, true, "Success create data, with some errors", gin.H{"error": 2})
		} else if countFail > 0 && countSuccess == 0 {
			Response(c, 200, true, "Cannot create data", gin.H{"error": 1})
		} else {
			Response(c, 200, true, "Success create data", gin.H{"error": 0})
		}
	}
}

func UsersImport() gin.HandlerFunc {
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

		headerList := tools.Array{"Name", "Email", "Username", "Password", "Role"}
		rows := []map[string]interface{}{}
		for i, each := range csvData {
			if i == 0 {
				if len(headerList) != len(each) {
					Response(c, 400, false, "invalid header column count for csv file", nil)
					return
				}
				for i := 0; i < len(headerList); i++ {
					if !headerList.Includes(each[i]) {
						Response(c, 400, false, "invalid header name for csv file", nil)
						return
					}
				}
			} else {
				row := make(map[string]interface{})
				for i := 0; i < len(headerList); i++ {
					row[fmt.Sprint(headerList[i])] = each[i]
				}
				rows = append(rows, row)
			}
		}

		Response(c, 200, true, "success convert", gin.H{"data": rows})
	}
}
