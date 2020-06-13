package entities

// User - minitube user
type User struct {
	Username string `form:"username" json:"username" binding:"required,alphanum,min=1,max=20"`
	Password string `form:"password" json:"password" binding:"required,hexadecimal,len=64"`
	Email    string `form:"email"    json:"email"    binding:"omitempty,email,max=50"`
}

// NewUser - return a user by username and password
func NewUser(username, password string) *User {
	return &User{Username: username, Password: password}
}

// NewUserFromMap - return a user from map
func NewUserFromMap(mp map[string]string) *User {
	// utils.Sugar.Debugf("NewUserFromMap: <%v> <%v>", mp["username"], mp["password"])
	if len(mp) == 0 {
		return nil
	}
	return NewUser(mp["username"], mp["password"])
}
