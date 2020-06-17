package store

import (
	"errors"
	"fmt"
	"minitube/models"
	"minitube/utils"
	"time"
)

var log = utils.Sugar

var timeout = 600 * time.Millisecond

// store's error
var (
	ErrStoreFailed        = errors.New("Store Error")
	ErrRedisFailed        = fmt.Errorf("%w From Redis", ErrStoreFailed)
	ErrMySQLFailed        = fmt.Errorf("%w From MySQL", ErrStoreFailed)
	ErrRedisUserNotExists = fmt.Errorf("%w user not exists", ErrRedisFailed)
	ErrMySQLUserNotExists = fmt.Errorf("%w user not exists", ErrMySQLFailed)
)

var (
	byID       = "id"
	byUsername = "username"
	byEmail    = "email"
	byPhone    = "phone"
)

// GetUserByID - get user from store by id.
func GetUserByID(id uint) (*models.User, error) {
	return getUserBy(byID, id)
}

// GetUserByUsername - get user from store by username.
func GetUserByUsername(username string) (*models.User, error) {
	return getUserBy(byUsername, username)
}

// GetUserByEmail - get user from store by email.
func GetUserByEmail(email string) (*models.User, error) {
	return getUserBy(byEmail, email)
}

// GetUserByPhone - get user from store by phone number.
func GetUserByPhone(phone string) (*models.User, error) {
	return getUserBy(byPhone, phone)
}

func getUserBy(by string, value interface{}) (*models.User, error) {
	var user *models.User
	var errRedis, errMysql error
	switch by {
	case byID:
		user, errRedis = getUserByIDFromRedis(value.(uint))
	case byUsername:
		user, errRedis = getUserByUsernameFromRedis(value.(string))
	case byEmail:
		user, errRedis = getUserByEmailFromRedis(value.(string))
	case byPhone:
		user, errRedis = getUserByPhoneFromRedis(value.(string))
	default:
		return nil, errors.New("Get user by " + by + " not support")
	}
	if errRedis == nil {
		return user, nil
	}
	switch by {
	case byID:
		user, errMysql = getUserByIDFromMysql(value.(uint))
	case byUsername:
		user, errMysql = getUserByUsernameFromMysql(value.(string))
	case byEmail:
		user, errMysql = getUserByEmailFromMysql(value.(string))
	case byPhone:
		user, errMysql = getUserByPhoneFromMysql(value.(string))
	default:
		return nil, errors.New("Get user by " + by + " not support")
	}
	if errMysql == nil {
		if errors.Is(errRedis, ErrRedisUserNotExists) {
			err := saveUserToRedis(user)
			if err != nil {
				log.Warnf("User %#v found in mysql, but store to redis failed: ", err)
			}
		}
		return user, nil
	}
	return nil, errMysql
}

// SaveUser - store user to mysql and redis
func SaveUser(user *models.User) error {
	err := saveUserToMysql(user)
	if err != nil {
		return err
	}
	err = saveUserToRedis(user)
	if err != nil {
		return err
	}
	return nil
}

// UpdateUserProfile - update user profile
func UpdateUserProfile(id uint, profile *models.ChangeProfileModel) error {
	user, err := GetUserByID(id)
	if err != nil {
		return err
	}

	err = updateUserProfileToMysql(user, profile)
	if err != nil {
		return err
	}

	return updateUserProfileToRedis(user, profile)
}


// ChangePassword - user change password to store
func ChangePassword(user *models.User, password string) error {
	err := changePasswordToMysql(user, password)
	if err != nil {
		return err
	}
	return changePasswordToRedis(user, password)
}

// CloseAll - close redis client and mysql connection.
func CloseAll() {
	client.Close()
	db.Close()
}
