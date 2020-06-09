package store

import (
	"errors"
	"fmt"
	"minitube/entities"
	"minitube/utils"
	"time"
)

var log = utils.Sugar

var timeout = 400 * time.Millisecond

// store's error
var (
	ErrStoreFailed        = errors.New("Store Error")
	ErrRedisFailed        = fmt.Errorf("%w From Redis", ErrStoreFailed)
	ErrMySQLFailed        = fmt.Errorf("%w From MySQL", ErrStoreFailed)
	ErrRedisUserNotExists = fmt.Errorf("%w user not exists", ErrRedisFailed)
	ErrMySQLUserNotExists = fmt.Errorf("%w user not exists", ErrMySQLFailed)
)

// GetUserByUsername - get user from store by username.
func GetUserByUsername(username string) (*entities.User, error) {
	user, errRedis := getUserByUsernameFromRedis(username)
	if errRedis == nil && user != nil {
		return user, nil
	}
	user, errMysql := getUserByUsernameFromMysql(username)
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
func SaveUser(user *entities.User) error {
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

// CloseAll - close redis client and
// mysql connection pool and mysql statement.
func CloseAll() {
	redisClient.Close()
	userQueryStmt.Close()
	mysqlPool.Close()
}
