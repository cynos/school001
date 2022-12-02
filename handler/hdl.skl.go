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

func SKLListPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tmpl = template.Must(template.ParseFiles(
			"pages/skl.html",
			"views/_head.html",
			"views/_header.html",
			"views/_sidebar.html",
			"views/_pagenav.html",
			"views/_modal.html",
			"views/_jsscript.html",
		))

		data := c.MustGet("userInfo").(jwt.MapClaims)

		// binding data with template
		err := tmpl.ExecuteTemplate(c.Writer, "skl", gin.H{
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

func SKLResultPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tmpl = template.Must(template.ParseFiles(
			"pages/skl-result.html",
			"views/_head.html",
			"views/_sidebar.html",
			"views/_jsscript.html",
			"views/_modal.html",
		))

		data := c.MustGet("userInfo").(jwt.MapClaims)

		skl := table.SKL{}
		err := core.App.DB.
			Preload("Users").
			Preload("Class").
			Preload("Class.Competence").
			Find(&skl, "users_username = ?", data["Username"]).Error
		if err != nil {
			core.App.Logger.Zap.Error("cannot find data", zap.Error(err))
		}

		// binding data with template
		err = tmpl.ExecuteTemplate(c.Writer, "skl-result", gin.H{
			"Funcs": Funcs{
				IsActiveNavLink: IsActiveNavLink(),
			},
			"path":     c.Request.URL.Path,
			"name":     data["Name"],
			"username": data["Username"],
			"skl":      skl,
			"allowed":  CheckAllowedPath(fmt.Sprint(data["Role"]), c.Request.URL.Path),
		})
		if err != nil {
			core.App.Logger.Zap.Error("cannot call page", zap.Error(err))
			http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		}
	}
}

// class logic data ================================================================================================

func SKLList() gin.HandlerFunc {
	reverse := func(x []table.SKL) []table.SKL {
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
			sklList     = []table.SKL{}
			order       string
			wherePaging string
			next        uint
			prev        uint
			first       uint
			last        uint
		)

		// get first and last id
		{
			var s string
			if search != "" {
				s = fmt.Sprintf(`where c2."name" like '%%%s%%' or c2.username like '%%%s%%'`, search, search)
			}
			err := core.App.DB.Raw(fmt.Sprintf(`
			select
				(select c1.id from skl c1 inner join users c2 on c2.username = c1.users_username %s order by c1.id asc limit 1),
				(select c1.id from skl c1 inner join users c2 on c2.username = c1.users_username %s order by c1.id desc limit 1)
			`, s, s)).Row().Scan(&first, &last)
			if err != nil {
				core.App.Logger.Zap.Error(err.Error())
				Response(c, 200, true, "data not found", gin.H{"data": nil, "error": err.Error()})
				return
			}
		}

		switch direction {
		case "next":
			order = "id DESC"
			wherePaging = fmt.Sprintf("id < %d", tools.StringToInt(cursor))
		case "prev":
			order = "id ASC"
			wherePaging = fmt.Sprintf("id > %d", tools.StringToInt(cursor))
		default:
			order = "id DESC"
			wherePaging = fmt.Sprintf("id < %d", tools.StringToInt(cursor))
		}

		if cursor == "0" {
			wherePaging = ""
		}

		// get users where
		username := []string{}
		rows, rowsErr := core.App.DB.Select("username").Table("users").Where("name like ? or username like ?", fmt.Sprintf("%%%s%%", search), fmt.Sprintf("%%%s%%", search)).Rows()
		if rowsErr != nil {
			Response(c, 200, true, "data not found", gin.H{"data": nil, "error": ""})
			return
		}
		for rows.Next() {
			var u string
			rows.Scan(&u)
			username = append(username, u)
		}

		q := core.App.DB.
			Preload("Users").
			Preload("Class").
			Preload("Class.Competence").
			Order(order).Limit(tools.StringToInt(limit)).Where("users_username IN ?", username).Where(wherePaging).Find(&sklList)
		if q.Error != nil {
			core.App.Logger.Zap.Error("error preload", zap.Error(q.Error))
		}

		if len(sklList) < 1 {
			Response(c, 200, true, "data not found", gin.H{"data": nil, "error": ""})
			return
		}

		switch direction {
		case "next":
			next = sklList[len(sklList)-1].ID
			prev = sklList[0].ID
		case "prev":
			sklList = reverse(sklList)
			next = sklList[len(sklList)-1].ID
			prev = sklList[0].ID
		default:
			next = sklList[len(sklList)-1].ID
			prev = sklList[0].ID
		}

		data := []map[string]interface{}{}
		for _, e := range sklList {
			row := make(map[string]interface{})
			tools.ForEachStruct(e, func(key string, val interface{}) {
				if (tools.Array{"CreatedAt"}).Includes(key) {
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

func SKLDelete() gin.HandlerFunc {
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

		query := core.App.DB.Unscoped().Where(fmt.Sprintf("id in(%s)", id)).Delete(&table.SKL{})
		err := query.Error
		if err != nil {
			Response(c, 200, false, "Cannot delete data", gin.H{"error": err.Error()})
			return
		}

		Response(c, 200, true, "Delete Succeed", nil)
	}
}

func SKLCreate() gin.HandlerFunc {
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
		err := core.App.DB.
			Preload("Class", "name_id like ?", fmt.Sprintf("%%%s%%", data["graduation"])).
			Find(&classMembers).
			Error
		if err != nil {
			Response(c, 200, false, "Cannot set graduation", gin.H{"error": err.Error()})
			return
		}

		if len(classMembers) < 1 {
			Response(c, 200, false, "Cannot set empty data", nil)
			return
		}

		skl := []table.SKL{}
		for _, v := range classMembers {
			if v.Class.ID == 0 {
				continue
			}
			skl = append(skl, table.SKL{
				Graduated:     true,
				ClassID:       v.Class.ID,
				UsersUsername: v.UsersUsername,
			})
		}

		if err := core.App.DB.Create(&skl).Error; err != nil {
			Response(c, 200, false, "Cannot set data", gin.H{"error": err.Error()})
			return
		}

		Response(c, 200, true, "Success set data", nil)
	}
}

func SKLSetEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		inc := c.MustGet("incoming").(*table.Incoming)
		inc.Save = true
		inc.HasResponse = true

		username := c.Query("username")
		email := c.Query("email")

		if email == "" || username == "" {
			Response(c, 200, false, "invalid paramaters", nil)
			return
		}

		users := table.Users{}
		query := core.App.DB.Last(&users, "username = ?", username)
		err := query.Error
		if err != nil {
			Response(c, 200, false, "Cannot get users data", gin.H{"error": err.Error()})
			return
		}

		users.Email = email
		core.App.DB.Save(&users)

		Response(c, 200, true, "Set Succeed", nil)
	}
}

func SKLSetManualGraduated() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.PostForm("ids")
		idArr := strings.Split(id, ",")

		if len(idArr) < 1 {
			Response(c, 200, false, "empty id", nil)
			return
		}

		action := c.PostForm("action")
		if action == "" {
			Response(c, 200, false, "invalid parameters", nil)
			return
		}

		err := core.App.DB.Model(&table.SKL{}).Where(fmt.Sprintf("id in(%s)", id)).Update("graduated", tools.StringToBool(action)).Error
		if err != nil {
			Response(c, 200, false, "Cannot update skl data", gin.H{"error": err.Error()})
			return
		}

		Response(c, 200, true, "Set Succeed", nil)
	}
}

func SKLDetails() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.PostForm("username")
		if username == "" {
			Response(c, 200, false, "invalid parameters", nil)
			return
		}

		details := table.SKLDetails{}
		err := core.App.DB.Where("nis = ?", username).Last(&details).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			Response(c, 200, false, "record not found", nil)
			return
		}

		Response(c, 200, true, "record found", gin.H{"data": details})
	}
}

func SKLPrint() gin.HandlerFunc {
	return func(c *gin.Context) {
		data := c.MustGet("userInfo").(jwt.MapClaims)

		username := fmt.Sprint(data["Username"])
		if username == "" {
			Response(c, 200, false, "invalid parameters", nil)
			return
		}

		details := table.SKLDetails{}
		err := core.App.DB.Where("nis = ?", username).Last(&details).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			Response(c, 200, false, "record not found", nil)
			return
		}

		// f, err := os.Open("./pages/skl-template.html")
		// if f != nil {
		// 	defer f.Close()
		// }
		// if err != nil {
		// 	core.App.Logger.Zap.Error("error open file template", zap.Error(err))
		// 	Response(c, 200, false, "internal service error", gin.H{"error": err.Error()})
		// 	return
		// }

		tmpl, err := template.ParseFiles("./pages/skl-template.html")
		if err != nil {
			core.App.Logger.Zap.Error("error cannot load file by template", zap.Error(err))
			Response(c, 200, false, "internal service error", gin.H{"error": err.Error()})
			return
		}

		err = tmpl.Execute(c.Writer, details)
		if err != nil {
			core.App.Logger.Zap.Error("error cannot parse file by template", zap.Error(err))
			Response(c, 200, false, "internal service error", gin.H{"error": err.Error()})
			return
		}

		// var tmplBytes bytes.Buffer
		// err = tmpl.Execute(&tmplBytes, details)
		// if err != nil {
		// 	core.App.Logger.Zap.Error("error cannot parse file by template", zap.Error(err))
		// 	Response(c, 200, false, "internal service error", gin.H{"error": err.Error()})
		// 	return
		// }

		// file, err := os.CreateTemp("tmp", "skl.*.html")
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// defer os.Remove(file.Name())

		// file.Write(tmplBytes.Bytes())

		// pdfg, err := wkhtmltopdf.NewPDFGenerator()
		// if err != nil {
		// 	core.App.Logger.Zap.Error("error create newPDFGenerator", zap.Error(err))
		// 	Response(c, 200, false, "internal service error", gin.H{"error": err.Error()})
		// 	return
		// }

		// pdfg.AddPage(wkhtmltopdf.NewPage(file.Name()))
		// pdfg.Orientation.Set(wkhtmltopdf.OrientationPortrait)
		// pdfg.Dpi.Set(300)

		// err = pdfg.Create()
		// if err != nil {
		// 	core.App.Logger.Zap.Error("error create pdf file", zap.Error(err))
		// 	Response(c, 200, false, "internal service error", gin.H{"error": err.Error()})
		// 	return
		// }

		// err = pdfg.WriteFile("./output.pdf")
		// if err != nil {
		// 	core.App.Logger.Zap.Error("error write pdf file", zap.Error(err))
		// 	Response(c, 200, false, "internal service error", gin.H{"error": err.Error()})
		// 	return
		// }

		// w := c.Writer
		// w.Header().Set("Content-Disposition", "attachment; filename=Surat Keterangan Lulus.pdf")
		// w.Header().Set("Content-Type", "application/pdf")
		// w.WriteHeader(http.StatusOK)
		// w.Write(pdfg.Bytes())
	}
}
