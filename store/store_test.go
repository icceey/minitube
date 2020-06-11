package store

import (
	"minitube/entities"
	"testing"

	"github.com/stretchr/testify/require"
)

// This test needs redis and mysql service.
// Please run test with `test.sh`.

var (
	user1 = entities.NewUser("1", "1")
	user2 = entities.NewUser("2", "2")
	user3 = entities.NewUser("3", "3")
	user4 = entities.NewUser("4", "4")
	user5 = entities.NewUser("5", "5")
)

func TestRedisConnection(t *testing.T) {
	err := pingRedis()
	require.NoError(t, err, "Redis should connected.")
}

func TestMySQLConnection(t *testing.T) {
	err := pingMySQL()
	require.NoError(t, err, "MySQL should connected.")
}

func TestMySQLCreateTable(t *testing.T) {
	// It should create table `user`.
	createTableIfNot()

	// Try again, it should ignore.
	createTableIfNot()
}

func TestMySQLInsertUser(t *testing.T) {
	// Add a user to table `user`.
	err := saveUserToMysql(user1)
	require.NoError(t, err, "User 1 should be inserted to mysql.")
}

func TestMySQLGetUserByUsername(t *testing.T) {
	require := require.New(t)

	// User 1 has inserted.
	user, err := getUserByUsernameFromMysql("1")
	require.NoError(err, "Get user1 from mysql should success.")
	require.Equal(user1, user, "User 1 should equal to user1.")

	// User 2 isn't exists, should return `ErrMySQLUserNotExists` err.
	user, err = getUserByUsernameFromMysql("2")
	require.Errorf(err, "Get user2 from mysql shouldn't success %#v.", user)
	require.EqualError(err, ErrMySQLUserNotExists.Error(), "Get user2 has an unexpected error.")
}

func TestRedisSaveUser(t *testing.T) {
	err := saveUserToRedis(user3)
	require.NoError(t, err, "Save user3 to redis should success.")
}

func TestRedisGetUser(t *testing.T) {
	require := require.New(t)

	// User 3 should in redis.
	user, err := getUserByUsernameFromRedis("3")
	require.NoError(err, "Get user3 from redis should success.")
	require.Equal(user3, user, "User 3 should equal to user3.")

	// User 4 shouldn't in redis, and this should return `ErrRedisUserNotExists` err.
	user, err = getUserByUsernameFromRedis("4")
	require.Error(err, "Get user4 from redis shouldn't success.")
	require.EqualError(err, ErrRedisUserNotExists.Error(), "Get user4 has an unexpected error.")
}

func TestGetUserOnlyInRedis(t *testing.T) {
	require := require.New(t)

	// User 3 only in redis.
	user, err := GetUserByUsername("3")
	require.NoError(err, "Get user3 from redis should success.")
	require.Equal(user3, user, "User 3 should equal to user3.")

	// MySQL don't has this user.
	user, err = getUserByUsernameFromMysql("3")
	require.Errorf(err, "Get user3 from mysql shouldn't success %#v.", user)
	require.EqualError(err, ErrMySQLUserNotExists.Error(), "Get user3 has an unexpected error.")
}

func TestGetUserOnlyInMysql(t *testing.T) {
	require := require.New(t)

	// User 1 not in redis.
	user, err := getUserByUsernameFromRedis("1")
	require.Error(err, "Get user1 from redis shouldn't success.")
	require.EqualError(err, ErrRedisUserNotExists.Error(), "Get user1 has an unexpected error.")

	// User 1 in mysql, so get should success.
	user, err = GetUserByUsername("1")
	require.NoError(err, "Get user1 should success.")
	require.Equal(user1, user, "User 3 should equal to user3.")

	// And then user 1 should store in redis.
	user, err = getUserByUsernameFromRedis("1")
	require.NoError(err, "Get user1 from redis should success.")
	require.Equal(user1, user, "User 1 should equal to user1.")
}

func TestStoreUser(t *testing.T) {
	require := require.New(t)

	err := SaveUser(user5)
	require.NoError(err, "Save user5 should success.")

	// get user 5 from redis.
	user, err := getUserByUsernameFromRedis("5")
	require.NoError(err, "Get user5 from redis should success.")
	require.Equal(user5, user, "User 5 should equal to user3.")

	// get user 5 from mysql.
	user, err = getUserByUsernameFromMysql("5")
	require.NoError(err, "Get user5 from mysql should success.")
	require.Equal(user5, user, "User 5 should equal to user5.")
}
