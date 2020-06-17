package models

import "github.com/jinzhu/gorm"

// Room - live room
type Room struct {
	gorm.Model
	UserID uint
	Name   *string `gorm:"type:varchar(30)"`
	Intro  *string `gorm:"type:varchar(200)"`
}
