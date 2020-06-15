package models

import "time"

// User - minitube user
type User struct {
	// Not use gorm.Model, in order to use json tag.
	ID        uint       `json:"id" gorm:"primary_key"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" sql:"index"`

	Username string  `json:"username"  gorm:"type:varchar(20);unique_index;not null"`
	Password string  `json:"password"  gorm:"type:char(64);not null"`
	Email    *string `json:"email"     gorm:"type:varchar(50);unique_index"`
	Phone    *string `json:"phone"     gorm:"type:varchar(18);unique_index"`
	LiveName *string `json:"live_name" gorm:"type:varchar(30)"`
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
		user.Phone = &v
	}
	if v, ok := mp["live_name"]; ok {
		user.LiveName = &v
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
