package models

import (
	"github.com/jinzhu/gorm"
)

// User - minitube user
type User struct {
	gorm.Model
	Username string  `gorm:"type:varchar(20);unique_index;not null"`
	Password string  `gorm:"type:char(64);not null"`
	Email    *string `gorm:"type:varchar(50);unique_index"`
	Phone    *string `gorm:"type:varchar(20);unique_index"`
}

// NewUser - return a user by username and password
func NewUser(username, password string) *User {
	return &User{
		Username: username,
		Password: password,
	}
}

// NewUserFromMap - return a user from map
func NewUserFromMap(mp map[string]string) *User {
	// utils.Sugar.Debugf("NewUserFromMap: <%v> <%v>", mp["username"], mp["password"])
	if len(mp) == 0 {
		return nil
	}
	user := NewUser(mp["username"], mp["password"])
	if v, ok := mp["email"]; ok {
		user.Email = &v
	}
	if v, ok := mp["phone"]; ok {
		user.Email = &v
	}
	return user
}

// NewUserFromRegister - new user from register model
func NewUserFromRegister(reg *RegisterModel) *User {
	user := NewUser(reg.Username, reg.Password)
	if reg.Email != "" {
		user.Email = &reg.Email
	}
	if reg.Phone != "" {
		user.Phone = &reg.Phone
	}
	return user
}
