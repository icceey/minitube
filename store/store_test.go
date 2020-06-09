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
	_, err := getUserByUsernameFromMysql("1")
	if err != nil {
		t.Fatal("Get user 1 by username failed.", err)
	}
	// User 2 isn't exists, should return `ErrMySQLUserNotExists` err.
	user, err := getUserByUsernameFromMysql("2")
	if err == nil {
		t.Fatalf("Get user 2 shouldn't success %#v", user)
	}
	if err != ErrMySQLUserNotExists {
		t.Fatal("Get user failed: ", err)
	}
}