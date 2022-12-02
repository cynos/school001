package handler

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	core "gitlab.com/cynomous/school001/common"
	table "gitlab.com/cynomous/school001/modules/tables"
	"gitlab.com/cynomous/school001/modules/tools"
)

const (
	ROLE_ADMIN = "admin"
	ROLE_GURU  = "guru"
	ROLE_SISWA = "siswa"
)

func AuthenticateUser(username, password string) (bool, table.Users) {
	var users = table.Users{}

	err := core.App.DB.Where("username = ?", username).Last(&users).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, users
	}

	match := tools.CheckPasswordHash(password, users.PasswordHash)
	if !match {
		return false, users
	}

	core.App.DB.Model(&users).Updates(table.Users{
		LastLogin: time.Now(),
	})

	return true, users
}

func GenerateUsers(name, email, username, password, role string) (table.Users, error) {
	var users = table.Users{}

	// remove whitespace username
	username = strings.TrimSpace(username)

	err := core.App.DB.Where("username = ?", username).Last(&users).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return users, fmt.Errorf("username already exist")
	}

	hash, err := tools.HashPassword(password)
	if err != nil {
		return users, fmt.Errorf("cannot hash password | detail : %s", err.Error())
	}

	users.Name = name
	users.Email = email
	users.Username = username
	users.PasswordHash = hash
	users.Role = role

	err = core.App.DB.Create(&users).Error
	if err != nil {
		return users, err
	}

	return users, nil
}

func UpdateUsers(id, name, email, username, password, role string) (table.Users, error) {
	if err := core.App.DB.Where("username = ? and id <> ?", username, id).Last(&table.Users{}).Error; err == nil {
		return table.Users{}, fmt.Errorf("email already used")
	}

	var user = table.Users{
		Name:     name,
		Email:    email,
		Username: username,
		Role:     role,
	}

	if password != "" {
		pass, err := tools.HashPassword(password)
		if err != nil {
			return table.Users{}, fmt.Errorf("cannot hash password | detail : %s", err.Error())
		}
		user.PasswordHash = pass
	}

	err := core.App.DB.Model(&table.Users{}).Where("id = ?", id).Updates(user).Error
	if err != nil {
		return user, err
	}

	return user, nil
}

func CheckAllowedPath(role string, path string) bool {
	listAllowedPath := map[string][]string{
		ROLE_ADMIN: {"/users", "/class", "/skl", "/modules"},
		ROLE_GURU:  {"/modules"},
		ROLE_SISWA: {"/surat-kelulusan"},
	}

	list, exist := listAllowedPath[role]
	if !exist {
		return false
	}

	var found bool
	for _, v := range list {
		if strings.HasPrefix(path, v) {
			found = true
			break
		}
	}
	return found
}
