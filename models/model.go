package models

// RegisterModel - register model
type RegisterModel struct {
	Username    string `form:"username" json:"username" binding:"required,alphanum,min=1,max=20"`
	Password    string `form:"password" json:"password" binding:"required,hexadecimal,len=64"`
	Email       string `form:"email"    json:"email"    binding:"omitempty,email,max=50"`
	PhoneNumber string `form:"phone"    json:"phone"    binding:"omitempty,e164"`
}

// LoginModel - login model
type LoginModel struct {
	Username    string `form:"username" json:"username" binding:"required_without_all=Email PhoneNumber,omitempty,alphanum,min=1,max=20"`
	Password    string `form:"password" json:"password" binding:"required,hexadecimal,len=64"`
	Email       string `form:"email"    json:"email"    binding:"required_without_all=Username PhoneNumber,omitempty,email,max=50"`
	PhoneNumber string `form:"phone"    json:"phone"    binding:"required_without_all=Username Email,omitempty,e164"`
}
