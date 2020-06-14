package store

import (
	"context"
	"encoding/json"
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
	result := redisClient.Get(ctx, getUserKey(username))
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

func saveUserToRedis(user *models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	userBytes, err := json.Marshal(user)
	if err != nil {
		log.Warnf("Marshal user %#v error: %v", user, err)
		return err
	}
	err = redisClient.Set(ctx, getUserKey(user.Username), userBytes, 0).Err()
	if err != nil {
		log.Warnf("Save user %#v to redis failed: %v", user, err)
		return err
	}
	return nil
}

func getUserKey(username string) string {
	return "user:" + username
}
