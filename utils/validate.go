package utils


// CheckUsername - check username is valid.
func CheckUsername(username string) bool {
	if len(username) == 0 || len(username) > 20 {
		return false
	}
	for _, c := range username {
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			continue
		}
		return false
	}
	return true
}

// CheckPassword - check password is valid.
func CheckPassword(password string) bool {
	if len(password) != 64 {
		return false
	}
	for _, c := range password {
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') {
			continue
		}
		return false
	}
	return true
}