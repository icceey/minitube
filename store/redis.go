package store

import (
	"context"
	"encoding/json"
	"errors"
	"minitube/models"
	"os"

	"github.com/go-redis/redis/v8"
)

var redisClient *redis.Client

func init() {
	log.Info("Initialize redis client...")
	redisClient = newRedisClient()

	log.Info("Checking redis service...")
	err := pingRedis()
	if err != nil {
		log.Fatal("Redis service access failed: ", err.Error())
	}

	log.Info("Redis is OK.")
}

func newRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
}

func pingRedis() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return redisClient.Ping(ctx).Err()
}

func getUserByUsernameFromRedis(username string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	result := redisClient.Get(ctx, wrapUsernameKey(username))
	jsonStr, err := result.Result()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrRedisUserNotExists
		}
		log.Warnf("Get user from redis failed: %v", err)
		return nil, ErrRedisFailed
	}
	user := new(models.User)
	err = json.Unmarshal([]byte(jsonStr), user)
	if err != nil {
		log.Warnf("Unmarshal <%v> to user err: %v", jsonStr, err)
		return nil, err
	}
	return user, nil
}

func getUserByEmailFromRedis(email string) (*models.User, error) {
	return getUserFromRedisBy(byEmail, email)
}

func getUserByPhoneFromRedis(phone string) (*models.User, error) {
	return getUserFromRedisBy(byPhone, phone)
}

func getUserFromRedisBy(by, value string) (*models.User, error) {
	username, err := getUsernameFromRedisBy(by, value)
	if err != nil {
		if err == redis.Nil {
			return nil, ErrRedisUserNotExists
		}
		log.Warnf("Get user from redis failed: %v", err)
		return nil, ErrRedisFailed
	}
	return getUserByUsernameFromRedis(username)
}

func getUsernameFromRedisBy(by, value string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var key string
	switch by {
	case byUsername:
		return value, nil
	case byEmail:
		key = wrapEmailKey(value)
	case byPhone:
		key = wrapPhoneKey(value)
	default:
		return "", errors.New("Get username by " + by + " not support")
	}
	result := redisClient.Get(ctx, key)
	err := result.Err()
	if err != nil {
		log.Warnf("Get username by %v %v error: %v", by, value, err)
		return "", err
	}
	return result.String(), nil
}

func saveUserToRedis(user *models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout*2)
	defer cancel()
	userBytes, err := json.Marshal(user)
	log.Debug("Save user redis ", string(userBytes))
	if err != nil {
		log.Warnf("Marshal user %#v error: %v", user, err)
		return err
	}
	err = redisClient.Set(ctx, wrapUsernameKey(user.Username), userBytes, 0).Err()
	if err != nil {
		log.Warnf("Save user %#v to redis failed: %v", user, err)
		return err
	}
	if user.Email != nil {
		err = redisClient.Set(ctx, wrapEmailKey(*user.Email), user.Username, 0).Err()
		if err != nil {
			log.Warnf("Create index email for user %#v to redis failed: %v", user, err)
			return err
		}
	}
	if user.Phone != nil {
		err = redisClient.Set(ctx, wrapPhoneKey(*user.Phone), user.Username, 0).Err()
		if err != nil {
			log.Warnf("Create index phone for user %#v to redis failed: %v", user, err)
			return err
		}
	}
	return nil
}

func updateUserProfileToRedis(username string, profile *models.ChangeProfileModel) error {
	log.Debug("updateUserProfileToRedis")
	user, err := GetUserByUsername(username)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout*2)
	defer cancel()

	log.Debug("del email key")
	err = redisClient.Del(ctx, wrapEmailKey(username)).Err()
	if err != nil {
		return err
	}
	if profile.Email == "" {
		user.Email = nil
	} else {
		user.Email = &profile.Email
	}
	log.Debug("del phone key")
	err = redisClient.Del(ctx, wrapPhoneKey(username)).Err()
	if err != nil {
		return err
	}
	if profile.Phone == "" {
		user.Phone = nil
	} else {
		user.Phone = &profile.Phone
	}
	
	if profile.LiveName == "" {
		user.LiveName = nil
	} else {
		user.LiveName = &profile.LiveName
	}
	return saveUserToRedis(user)
}

func wrapUserKey(key string) string {
	return "user:" + key
}

func wrapUsernameKey(username string) string {
	return wrapUserKey("username:" + username)
}

func wrapEmailKey(email string) string {
	return wrapUserKey("email:" + email)
}

func wrapPhoneKey(phone string) string {
	return wrapUserKey("phone:" + phone)
}
