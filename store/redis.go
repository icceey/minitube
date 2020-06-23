package store

import (
	"context"
	"encoding/json"
	"errors"
	"minitube/models"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

var client *redis.Client

func init() {
	log.Info("Initialize redis client...")
	client = NewRedisClient()

	log.Info("Checking redis service...")
	err := pingRedis()
	if err != nil {
		log.Fatal("Redis service access failed: ", err.Error())
	}

	log.Info("Redis is OK.")
}

// NewRedisClient - new redis client
func NewRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
}

func pingRedis() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return client.Ping(ctx).Err()
}

func getUserByIDFromRedis(id uint) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	result := client.Get(ctx, wrapIDKey(id))
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

func getUserByUsernameFromRedis(username string) (*models.User, error) {
	return getUserFromRedisBy(byUsername, username)
}

func getUserByEmailFromRedis(email string) (*models.User, error) {
	return getUserFromRedisBy(byEmail, email)
}

func getUserByPhoneFromRedis(phone string) (*models.User, error) {
	return getUserFromRedisBy(byPhone, phone)
}

func getUserFromRedisBy(by string, value interface{}) (*models.User, error) {
	id, err := getIDFromRedisBy(by, value)
	if err != nil {
		if err == redis.Nil {
			return nil, ErrRedisUserNotExists
		}
		log.Warnf("Get user from redis failed: %v", err)
		return nil, ErrRedisFailed
	}
	return getUserByIDFromRedis(id)
}

func getIDFromRedisBy(by string, value interface{}) (uint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var key string
	switch by {
	case byID:
		return value.(uint), nil
	case byUsername:
		key = wrapUsernameKey(value.(string))
	case byEmail:
		key = wrapEmailKey(value.(string))
	case byPhone:
		key = wrapPhoneKey(value.(string))
	default:
		return 0, errors.New("Get username by " + by + " not support")
	}
	result := client.Get(ctx, key)
	err := result.Err()
	if err != nil {
		return 0, err
	}
	id, err := result.Int()
	if err != nil {
		return 0, err
	}
	return uint(id), nil
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

	pipe := client.TxPipeline()

	// map userID -> user
	err = pipe.Set(ctx, wrapIDKey(user.ID), userBytes, 0).Err()
	if err != nil {
		log.Warnf("Save user %#v to redis failed: %v", user, err)
		return err
	}

	// map username -> userID
	err = pipe.Set(ctx, wrapUsernameKey(user.Username), user.ID, 0).Err()
	if err != nil {
		log.Warnf("Create index username for user %#v to redis failed: %v", user, err)
		return err
	}

	// map email -> userID
	if user.Email != nil {
		err = pipe.Set(ctx, wrapEmailKey(*user.Email), user.ID, 0).Err()
		if err != nil {
			log.Warnf("Create index email for user %#v to redis failed: %v", user, err)
			return err
		}
	}

	// map phone -> userID
	if user.Phone != nil {
		err = pipe.Set(ctx, wrapPhoneKey(*user.Phone), user.ID, 0).Err()
		if err != nil {
			log.Warnf("Create index phone for user %#v to redis failed: %v", user, err)
			return err
		}
	}

	_, err = pipe.Exec(ctx)
	return err
}

func updateUserProfileToRedis(user *models.User, profile *models.ChangeProfileModel) error {
	// log.Debug("updateUserProfileToRedis")
	err := setProfileRedis(user, profile)
	if err != nil {
		return err
	}

	return saveUserToRedis(user)
}

func setProfileRedis(user *models.User, profile *models.ChangeProfileModel) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout*2)
	defer cancel()

	pipe := client.TxPipeline()

	// log.Debug("del email key")
	if (user.Email == nil && profile.Email != "") || (user.Email != nil && *user.Email != profile.Email) {
		err := pipe.Del(ctx, wrapEmailKey(strconv.Itoa(int(user.ID)))).Err()
		if err != nil {
			return err
		}
		if profile.Email == "" {
			user.Email = nil
		} else {
			user.Email = &profile.Email
		}
	}
	// log.Debug("del phone key")
	if (user.Phone == nil && profile.Phone != "") || (user.Phone != nil && *user.Phone != profile.Phone) {
		err := pipe.Del(ctx, wrapPhoneKey(strconv.Itoa(int(user.ID)))).Err()
		if err != nil {
			return err
		}
		if profile.Phone == "" {
			user.Phone = nil
		} else {
			user.Phone = &profile.Phone
		}
	}

	if profile.LiveName == "" {
		user.Room.Name = nil
	} else {
		user.Room.Name = &profile.LiveName
	}
	if profile.LiveIntro == "" {
		user.Room.Intro = nil
	} else {
		user.Room.Intro = &profile.LiveIntro
	}

	_, err := pipe.Exec(ctx)
	return err
}

func changePasswordToRedis(user *models.User, password string) error {
	user.Password = password
	return saveUserToRedis(user)
}

// GetLivingUsernameList - get who is living
func GetLivingUsernameList(num int64) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout*2)
	defer cancel()

	result, err := client.SRandMemberN(ctx, "living", num).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return []string{}, nil
		}
		log.Warn("GetLivingUsernameList: ", err)
		return []string{}, err
	}

	return result, nil
}

// GetUserIsLiving - whether user is living
func GetUserIsLiving(username string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout*2)
	defer cancel()

	living, err := client.SIsMember(ctx, "living", username).Result()
	if err != nil {
		log.Warn("GetUserIsLiving: ", err)
	}
	return living, err
}

// GetLivingTime - get when user start living 
func GetLivingTime(username string) (*time.Time, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout*2)
	defer cancel()

	result, err := client.Get(ctx, "living:"+username).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		log.Warn("GetLivingTime: ", err)
		return nil, err
	}
	
	t, err := time.Parse(time.RFC3339, result)
	if err != nil {
		log.Warn("GetLivingTime: ", err)
		return nil, err
	}
	return &t, nil
}

// UpdateWatchHistory - update watch history
func UpdateWatchHistory(id uint, username string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err := client.ZAdd(ctx, wrapHistoryKey(id), &redis.Z{
		Score: float64(time.Now().Unix()),
		Member: username,
	}).Err()

	if err != nil {
		log.Warn("UpdateWatchHistory: ", err)
		return err
	}

	return nil
}

// GetWatchHistory - get watch history
func GetWatchHistory(id uint) ([]*models.History, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	result, err := client.ZRevRangeWithScores(ctx, wrapHistoryKey(id), 0, 31).Result()
	if err != nil {
		log.Warn(err)
	}

	s := make([]*models.History, len(result))
	for i := range result {
		s[i] = models.ZToHistory(&result[i])
	}

	return s, err
}

func wrapUserKey(key string) string {
	return "user:" + key
}

func wrapIDKey(id uint) string {
	return wrapUserKey("id:" + strconv.Itoa(int(id)))
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

func wrapHistoryKey(id uint) string {
	return wrapUserKey("history:"+strconv.Itoa(int(id)))
}