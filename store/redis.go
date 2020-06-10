package store

import (
	"context"
	"minitube/entities"
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


func getUserByUsernameFromRedis(username string) (*entities.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	result := redisClient.HGetAll(ctx, getUserKey(username))
	mp, err := result.Result()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrRedisUserNotExists
		}
		log.Warnf("Get user from redis failed: %v", err)
		return nil, ErrRedisFailed
	}
	user := entities.NewUserFromMap(mp)
	if user == nil {
		return nil, ErrRedisUserNotExists
	}
	return user, nil
}


func saveUserToRedis(user *entities.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	// log.Debugf("Save %#v to Redis", user)
	result := redisClient.HSet(ctx, getUserKey(user.Username), 
		"username", user.Username,
		"password", user.Password,
	)
	err := result.Err()
	if err != nil {
		log.Warnf("Save user %#v to redis failed: %v", user, err)
		return err
	}
	return nil
}


func getUserKey(username string) string {
	return "user:" + username
}