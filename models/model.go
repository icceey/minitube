package models

import (
	"time"

	"github.com/go-redis/redis/v8"
)

// RegisterModel - register request model
type RegisterModel struct {
	Username string `form:"username" json:"username" binding:"required,alphanum,min=1,max=20"`
	Password string `form:"password" json:"password" binding:"required,hexadecimal,len=64"`
	Email    string `form:"email"    json:"email"    binding:"omitempty,email,max=50"`
	Phone    string `form:"phone"    json:"phone"    binding:"omitempty,e164"`
}

// LoginModel - login request model
type LoginModel struct {
	Username string `form:"username" json:"username" binding:"required_without_all=Email Phone,omitempty,alphanum,min=1,max=20"`
	Password string `form:"password" json:"password" binding:"required,hexadecimal,len=64"`
	Email    string `form:"email"    json:"email"    binding:"required_without_all=Username Phone,omitempty,email,max=50"`
	Phone    string `form:"phone"    json:"phone"    binding:"required_without_all=Username Email,omitempty,e164"`
}

// ChangeProfileModel - user change profile request model
type ChangeProfileModel struct {
	Email     string `form:"email"      json:"email"      binding:"omitempty,email,max=50"`
	Phone     string `form:"phone"      json:"phone"      binding:"omitempty,e164"`
	LiveName  string `form:"live_name"  json:"live_name"  binding:"omitempty,max=30"`
	LiveIntro string `form:"live_intro" json:"live_intro" binding:"omitempty,max=200"`
}

// MapUser - get ChangeProfileModel in map
func (m *ChangeProfileModel) MapUser() map[string]interface{} {
	mp := make(map[string]interface{})
	if m.Email != "" {
		mp["email"] = m.Email
	} else {
		mp["email"] = nil
	}
	if m.Phone != "" {
		mp["phone"] = m.Phone
	} else {
		mp["phone"] = nil
	}
	return mp
}

// MapRoom - get room profile in map
func (m *ChangeProfileModel) MapRoom() map[string]interface{} {
	mp := make(map[string]interface{})
	if m.LiveName != "" {
		mp["name"] = m.LiveName
	} else {
		mp["name"] = nil
	}
	if m.LiveIntro != "" {
		mp["intro"] = m.LiveIntro
	} else {
		mp["intro"] = nil
	}
	return mp
}

// Me - getMe respone model
type Me struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Username  string    `json:"username"`
	Email     *string   `json:"email"`
	Phone     *string   `json:"phone"`
	LiveName  *string   `json:"live_name"`
	LiveIntro *string   `json:"live_intro"`
}

// GetMeFromUser - get Me from User
func GetMeFromUser(user *User) *Me {
	me := &Me{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Username:  user.Username,
		Email:     user.Email,
		Phone:     user.Phone,
		LiveName:  user.Room.Name,
		LiveIntro: user.Room.Intro,
	}
	if user.UpdatedAt.After(user.Room.UpdatedAt) {
		me.UpdatedAt = user.UpdatedAt
	} else {
		me.UpdatedAt = user.Room.UpdatedAt
	}
	return me
}

// ChangePasswordModel - change password request model
type ChangePasswordModel struct {
	OldPassword string `json:"old_password" form:"old_password" binding:"required,hexadecimal,len=64"`
	NewPassword string `json:"new_password" form:"new_password" binding:"required,hexadecimal,len=64"`
}

// PublicUser - public user don't have private info.
type PublicUser struct {
	Username  string     `json:"username"`
	RoomName  *string    `json:"live_name"`
	RoomIntro *string    `json:"live_intro"`
	Living    bool       `json:"living"`
	StartTime *time.Time `json:"start_time"`
	Watching  int        `json:"watching"`
	Follow    int        `json:"follow"`
}

// LivingListModel - living list response model
type LivingListModel struct {
	Total int           `json:"total"`
	Users []*PublicUser `json:"users"`
}

// History - watch history
type History struct {
	Username  string `json:"username"`
	TimeStamp int64  `json:"timestamp"`
}

// ZToHistory - parse redis.Z to History
func ZToHistory(z *redis.Z) *History {
	return &History{
		Username:  z.Member.(string),
		TimeStamp: int64(z.Score),
	}
}
