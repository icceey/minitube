package models

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
	Email    string `form:"email"     json:"email"     binding:"omitempty,email,max=50"`
	Phone    string `form:"phone"     json:"phone"     binding:"omitempty,e164"`
	LiveName string `form:"live_name" json:"live_name" binding:"omitempty,max=30"`
}

// Map - get ChangeProfileModel in map
func (m *ChangeProfileModel) Map() map[string]interface{} {
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
	if m.LiveName != "" {
		mp["live_name"] = m.LiveName
	} else {
		mp["live_name"] = nil
	}
	return mp
}