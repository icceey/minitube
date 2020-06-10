package store

import (
	"minitube/entities"
	"testing"
)

// This test needs redis and mysql service.
// Please run test with `test.sh`.

func TestRedisConnection(t *testing.T) {
	if err := pingRedis(); err != nil {
		t.Fatal(err.Error())
	}
}

func TestMySQLConnection(t *testing.T) {
	if err := pingMySQL(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLCreateTable(t *testing.T) {
	// It should create table `user`.
	createTableIfNot()
	// Try again, it should ignore.
	createTableIfNot()
}

func TestMySQLInsertUser(t *testing.T) {
	// Add a user to table `user`.
	err := saveUserToMysql(&entities.User{Username: "1", Password: "1"})
	if err != nil {
		t.Fatal("Insert user failed.", err)
	}
}

func TestMySQLGetUserByUsername(t *testing.T) {
	// User 1 has inserted.
	user, err := getUserByUsernameFromMysql("1")
	if err != nil {
		t.Fatal("Get user 1 by username failed.", err)
	}
	if user.Username != "1" || user.Password != "1" {
		t.Fatal("Get wrong user from MySQL.", user)
	}
	// User 2 isn't exists, should return `ErrMySQLUserNotExists` err.
	user, err = getUserByUsernameFromMysql("2")
	if err == nil {
		t.Fatalf("Get user 2 shouldn't success %#v", user)
	}
	if err != ErrMySQLUserNotExists {
		t.Fatal("Get user failed: ", err)
	}
}

func TestRedisSaveUser(t *testing.T) {
	user := &entities.User{
		Username: "3",
		Password: "3",
	}
	err := saveUserToRedis(user)
	if err != nil {
		t.Fatal("Save user 3 to redis failed.", err)
	}
}

func TestRedisGetUser(t *testing.T) {
	// User 3 should in redis.
	user, err := getUserByUsernameFromRedis("3")
	if err != nil {
		t.Fatal("Get user 3 from redis failed.", err)
	}
	if user.Username != "3" || user.Password != "3" {
		t.Fatal("Get wrong user 3 from redis.", user)
	}
	// User 4 shouldn't in redis, and this should return `ErrRedisUserNotExists` err.
	user, err = getUserByUsernameFromRedis("4")
	if err == nil {
		t.Fatal("Get user 4 shouldn't success.", user)
	}
	if err != ErrRedisUserNotExists {
		t.Fatal("Get user 4 failed.", err)
	}
}


func TestGetUserOnlyInRedis(t *testing.T) {
	// User 3 only in redis.
	user, err := GetUserByUsername("3")
	if err != nil {
		t.Fatal("Get user 3 from redis failed.", err)
	}
	if user.Username != "3" || user.Password != "3" {
		t.Fatal("Get wrong user 3 from redis.", user)
	}
	// MySQL don't has this user.
	user, err = getUserByUsernameFromMysql("3")
	if err == nil {
		t.Fatalf("Get user 3 shouldn't success %#v", user)
	}
	if err != ErrMySQLUserNotExists {
		t.Fatal("Get user failed: ", err)
	}
}


func TestGetUserOnlyInMysql(t *testing.T) {
	// User 1 not in redis.
	user, err := getUserByUsernameFromRedis("1")
	if err == nil {
		t.Fatal("Get user 1 shouldn't success.", user)
	}
	if err != ErrRedisUserNotExists {
		t.Fatal("Get user 1 failed.", err)
	}
	// User 1 in mysql, so get should success.
	user, err = GetUserByUsername("1")
	if err != nil {
		t.Fatal("Get user 1 by username failed.", err)
	}
	if user.Username != "1" || user.Password != "1" {
		t.Fatal("Get wrong user from MySQL.", user)
	}
	// And then user 1 should store in redis.
	user, err = getUserByUsernameFromRedis("1")
	if err != nil {
		t.Fatal("Get user 1 from mysql but not store in redis.", err)
	}
	if user.Username != "1" || user.Password != "1" {
		t.Fatal("Get wrong user 1 from redis.", user)
	}
}

func TestStoreUser(t *testing.T) {
	user := &entities.User{
		Username: "5",
		Password: "5",
	}
	err := SaveUser(user)
	if err != nil {
		t.Fatal("Save user failed.", err)
	}
	// get user 5 from redis.
	userGet, err := getUserByUsernameFromRedis("5")
	if err != nil {
		t.Fatal("Get user 5 from redis failed.", err)
	}
	if userGet.Username != "5" || userGet.Password != "5" {
		t.Fatal("Get wrong user 5 from redis.", user)
	}
	// get user 5 from mysql.
	userGet, err = getUserByUsernameFromMysql("5")
	if err != nil {
		t.Fatal("Get user 5 from mysql failed.", err)
	}
	if userGet.Username != "5" || userGet.Password != "5" {
		t.Fatal("Get wrong user 5 from mysql.", user)
	}
}