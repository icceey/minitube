package entities

import "minitube/utils"

// User - minitube user
type User struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

// NewUserFromMap - return a user from map
func NewUserFromMap(mp map[string]string) *User {
	utils.Sugar.Debugf("NewUserFromMap: <%v> <%v>", mp["username"], mp["password"])
	if len(mp) == 0 {
		return nil
	}
	return &User{
		Username: mp["username"],
		Password: mp["password"],
	}
}