package store

import (
	"minitube/models"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

// This test needs redis and mysql service.
// Please run test with `test.sh`.

var users []*models.User

func TestRedisConnection(t *testing.T) {
	err := pingRedis()
	require.NoError(t, err, "Redis should connected.")
}

func TestMySQLConnection(t *testing.T) {
	err := pingMySQL()
	require.NoError(t, err, "MySQL should connected.")
}

func TestMySQLInsertUser(t *testing.T) {
	// Add user 0-9 to mysql.
	for i := 0; i < 10; i++ {
		err := saveUserToMysql(users[i])
		require.NoErrorf(t, err, "User[%v] should be inserted to mysql.", i)
	}

}

func TestMySQLGetUserByUsername(t *testing.T) {

	// User 0-9 has inserted.
	checkUserInMySQL(t, 0, 10)

	// User 10-19 isn't exists, should return `ErrMySQLUserNotExists` err.
	checkUserNotInMySQL(t, 10, 20)

}

func TestRedisSaveUser(t *testing.T) {
	// Save user 10-19 to redis.
	for i := 10; i < 20; i++ {
		err := saveUserToRedis(users[i])
		require.NoErrorf(t, err, "Save user[%v] to redis should success.", i)
	}

}

func TestRedisGetUser(t *testing.T) {

	// User 10-19 should in redis.
	checkUserInRedis(t, 10, 20)

	// User 0-9 shouldn't in redis, and this should return `ErrRedisUserNotExists` err.
	checkUserNotInRedis(t, 0, 10)

}

func TestGetUserOnlyInRedis(t *testing.T) {

	// User 10-19 only in redis.
	checkUserInRedis(t, 10, 20)

	// User 10-19 isn't exists, should return `ErrMySQLUserNotExists` err.
	checkUserNotInMySQL(t, 10, 20)
}

func TestGetUserOnlyInMysql(t *testing.T) {
	require := require.New(t)

	// User 0-9 not in redis.
	checkUserNotInRedis(t, 0, 10)

	// User 0-9 in mysql, so get should success.
	for i := 0; i < 10; i++ {
		user, err := GetUserByUsername(strconv.Itoa(i))
		require.NoErrorf(err, "Get user[%v] from mysql should success.", i)
		user.CreatedAt = users[i].CreatedAt
		user.UpdatedAt = users[i].UpdatedAt
		require.Equalf(users[i], user, "User %v should equal to user[%v].", i, i)
	}

	// And then User 0-9 should store in redis.
	checkUserInRedis(t, 0, 10)
}

func TestStoreUser(t *testing.T) {
	for i := 20; i < 30; i++ {
		err := SaveUser(users[i])
		require.NoError(t, err, "Save user[%v] should success.", i)
	}
	checkUserInRedis(t, 20, 30)
	checkUserInMySQL(t, 20, 30)
}

func TestUpdateUserProfileToMysql(t *testing.T) {
	require := require.New(t)
	for i := 0; i < 10; i++ {
		profile := new(models.ChangeProfileModel)
		profile.Email = "0" + strconv.Itoa(i) + "@minitube.com"
		profile.Phone = "+11370000000" + strconv.Itoa(i)
		profile.LiveName = strconv.Itoa(i) + "'s living room"
		err := updateUserProfileToMysql(users[i].Username, profile)
		require.NoError(err, "update shouldn't error")
		users[i].Email = &profile.Email
		users[i].Phone = &profile.Phone
		users[i].LiveName = &profile.LiveName
	}
	checkUserInMySQL(t, 0, 10)
	t.Log("Test clear profile")
	for i := 0; i < 10; i++ {
		profile := new(models.ChangeProfileModel)
		err := updateUserProfileToMysql(users[i].Username, profile)
		require.NoError(err, "update shouldn't error")
		users[i].Email = nil
		users[i].Phone = nil
		users[i].LiveName = nil
	}
	checkUserInMySQL(t, 0, 10)
}

func TestUpdateUserProfileToRedis(t *testing.T) {
	require := require.New(t)
	for i := 0; i < 10; i++ {
		profile := new(models.ChangeProfileModel)
		profile.Email = "0" + strconv.Itoa(i) + "@minitube.com"
		profile.Phone = "+11370000000" + strconv.Itoa(i)
		profile.LiveName = strconv.Itoa(i) + "'s living room"
		err := updateUserProfileToRedis(users[i].Username, profile)
		require.NoError(err, "update shouldn't error")
		users[i].Email = &profile.Email
		users[i].Phone = &profile.Phone
		users[i].LiveName = &profile.LiveName
	}
	checkUserInRedis(t, 0, 10)
	t.Log("Test clear profile")
	for i := 0; i < 10; i++ {
		profile := new(models.ChangeProfileModel)
		err := updateUserProfileToRedis(users[i].Username, profile)
		require.NoError(err, "update shouldn't error")
		users[i].Email = nil
		users[i].Phone = nil
		users[i].LiveName = nil
	}
	checkUserInRedis(t, 0, 10)
}

func TestUpdateUserProfile(t *testing.T) {
	require := require.New(t)
	for i := 0; i < 10; i++ {
		profile := new(models.ChangeProfileModel)
		profile.Email = "0" + strconv.Itoa(i) + "@minitube.com"
		profile.Phone = "+11370000000" + strconv.Itoa(i)
		profile.LiveName = strconv.Itoa(i) + "'s living room"
		err := UpdateUserProfile(users[i].Username, profile)
		require.NoError(err, "update shouldn't error")
		users[i].Email = &profile.Email
		users[i].Phone = &profile.Phone
		users[i].LiveName = &profile.LiveName
	}
	checkUserInRedis(t, 0, 10)
	checkUserInMySQL(t, 0, 10)
	t.Log("Test clear profile")
	for i := 0; i < 10; i++ {
		profile := new(models.ChangeProfileModel)
		err := UpdateUserProfile(users[i].Username, profile)
		require.NoError(err, "update shouldn't error")
		users[i].Email = nil
		users[i].Phone = nil
		users[i].LiveName = nil
	}
	checkUserInRedis(t, 0, 10)
	checkUserInMySQL(t, 0, 10)
}

func createUserForTest() {
	users = make([]*models.User, 0, 50)
	phone := int64(13688866600)
	for i := 0; i < 5; i++ {
		for j := 0; j < 10; j++ {
			id := i*10 + j
			user := models.NewUserFromMap(map[string]string{
				"username": strconv.Itoa(id),
				"password": strconv.Itoa(id),
			})
			if j%5 == 1 || j%5 == 3 {
				email := strconv.Itoa(id) + "@minitube.com"
				user.Email = &email
			}
			if j%5 == 2 || j%5 == 3 {
				phone := strconv.FormatInt(phone, 10)
				user.Phone = &phone
			}
			users = append(users, user)
			phone++
		}
	}
}

func checkUserInRedis(t *testing.T, from, to int) {
	require := require.New(t)
	for i := from; i < to; i++ {
		user, err := getUserByUsernameFromRedis(strconv.Itoa(i))
		require.NoErrorf(err, "Get user[%v] from redis should success.", i)
		user.CreatedAt = users[i].CreatedAt
		user.UpdatedAt = users[i].UpdatedAt
		require.Equalf(users[i], user, "User %v should equal to user[%v].", i, i)
	}
}

func checkUserNotInRedis(t *testing.T, from, to int) {
	require := require.New(t)
	for i := from; i < to; i++ {
		user, err := getUserByUsernameFromRedis(strconv.Itoa(i))
		require.Errorf(err, "Get user[%v]from redis shouldn't success. %#v", i, user)
		require.EqualErrorf(err, ErrRedisUserNotExists.Error(), "Get user[%v] has an unexpected error.", i)
	}
}

func checkUserNotInMySQL(t *testing.T, from, to int) {
	require := require.New(t)
	for i := from; i < to; i++ {
		user, err := getUserByUsernameFromMysql(strconv.Itoa(i))
		require.Errorf(err, "Get user[%v] from mysql shouldn't success %#v.", i, user)
		require.EqualErrorf(err, ErrMySQLUserNotExists.Error(), "Get user[%v] has an unexpected error.", i)
	}
}

func checkUserInMySQL(t *testing.T, from, to int) {
	require := require.New(t)
	for i := from; i < to; i++ {
		user, err := getUserByUsernameFromMysql(strconv.Itoa(i))
		require.NoErrorf(err, "Get user[%v] from mysql should success.", i)
		user.CreatedAt = users[i].CreatedAt
		user.UpdatedAt = users[i].UpdatedAt
		require.Equalf(users[i], user, "User %v should equal to user[%v].", i, i)
	}
}

func TestMain(m *testing.M) {
	createUserForTest()
	os.Exit(m.Run())
}
