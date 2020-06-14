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
	byUsername = "username"
	byEmail    = "email"
	byPhone    = "phone"
)

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

func getUserBy(by, value string) (*models.User, error) {
	var user *models.User
	var errRedis, errMysql error
	switch by {
	case byUsername:
		user, errRedis = getUserByUsernameFromRedis(value)
	case byEmail:
		user, errRedis = getUserByEmailFromRedis(value)
	case byPhone:
		user, errRedis = getUserByPhoneFromRedis(value)
	default:
		return nil, errors.New("Get user by " + by + " not support")
	}
	if errRedis == nil {
		return user, nil
	}
	switch by {
	case byUsername:
		user, errMysql = getUserByUsernameFromMysql(value)
	case byEmail:
		user, errMysql = getUserByEmailFromMysql(value)
	case byPhone:
		user, errMysql = getUserByPhoneFromMysql(value)
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

// CloseAll - close redis client and mysql connection.
func CloseAll() {
	redisClient.Close()
	db.Close()
}
