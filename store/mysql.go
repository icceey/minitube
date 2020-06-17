package store

import (
	"context"
	"fmt"
	"minitube/models"
	"os"
	"time"

	"github.com/jinzhu/gorm"

	_ "github.com/go-sql-driver/mysql" // mysql driver
)

var db *gorm.DB

func init() {
	var err error
	dataSourceName := fmt.Sprintf("%v:%v@tcp(%v)/%v?charset=utf8&parseTime=True&loc=Local",
		os.Getenv("MYSQL_USER"), os.Getenv("MYSQL_PASSWORD"), os.Getenv("MYSQL_ADDR"), os.Getenv("MYSQL_DATABASE"))

	log.Info("Connect to MySQL service...")
	retry, interval := 5, 10
	for i := 0; i < retry; i++ {
		db, err = gorm.Open("mysql", dataSourceName)
		if err == nil {
			break
		}
		errMsg := fmt.Sprintf("Connect to MySQL failed %v times, ", i+1)
		if i == retry-1 {
			errMsg += "maybe some errors have occurred."
			log.Fatal(errMsg, "MySQL service access failed: ", err)
		} else {
			errMsg += fmt.Sprintf("maybe mysql container is not ready? Will retry after %v seconds", interval)
			log.Warn(errMsg)
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}

	db.SingularTable(true)
	db.DB().SetConnMaxLifetime(0)
	db.DB().SetMaxIdleConns(3)
	db.DB().SetMaxOpenConns(3)

	db.AutoMigrate(&models.User{})
	db.AutoMigrate(&models.Room{})

	if debug := os.Getenv("DEBUG"); debug == "true" {
		db = db.Debug()
	}

	log.Info("MySQL is OK.")
}

func pingMySQL() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return db.DB().PingContext(ctx)
}

func getUserByUsernameFromMysql(username string) (*models.User, error) {
	return getUserFromMysqlBy(byUsername, username)
}

func getUserByEmailFromMysql(email string) (*models.User, error) {
	return getUserFromMysqlBy(byEmail, email)
}

func getUserByPhoneFromMysql(phone string) (*models.User, error) {
	return getUserFromMysqlBy(byPhone, phone)
}

func getUserByIDFromMysql(id uint) (*models.User, error) {
	return getUserFromMysqlBy(byID, id)
}

func getUserFromMysqlBy(by string, value interface{}) (*models.User, error) {
	tx := db.Begin()
	user := new(models.User)
	err := tx.Where(by+" = ?", value).Take(user).Error
	if err != nil {
		tx.Rollback()
		if gorm.IsRecordNotFoundError(err) {
			return nil, ErrMySQLUserNotExists
		}
		log.Warnf("Get user %v from Mysql failed: %v", value, err)
		return nil, ErrMySQLFailed
	}
	room := new(models.Room)
	err = tx.Model(user).Related(room).Error
	if err != nil {
		tx.Rollback()
		if gorm.IsRecordNotFoundError(err) {
			return user, nil
		}
		log.Warnf("Get user %v's room from Mysql failed: %v", value, err)
		return nil, ErrMySQLFailed
	}
	user.Room = *room
	return user, tx.Commit().Error
}

// func getPasswordFromMysql(username string) (string, error) {
// 	result := db.Model(models.User{}).Select("password").Where("username = ?", username)
// 	err := result.Error
// 	if err != nil {
// 		if gorm.IsRecordNotFoundError(err) {
// 			return "", ErrMySQLUserNotExists
// 		}
// 		log.Warnf("Get user %v's password from Mysql failed: %v", username, err)
// 		return "", ErrMySQLFailed
// 	}
// 	var pass string
// 	err = result.Row().Scan(&pass)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return "", ErrMySQLUserNotExists
// 		}
// 		log.Warnf("Get user %v's password from Mysql failed: %v", username, err)
// 		return "", ErrMySQLFailed
// 	}
// 	return pass, nil
// }

func saveUserToMysql(user *models.User) error {
	if db.NewRecord(user) {
		// log.Debugf("%#v", user)
		err := db.Create(user).Error
		if err != nil {
			log.Warnf("Save user %#v to Mysql failed: %v", user, err)
			return err
		}
		return nil
	}
	return nil
}

func updateUserProfileToMysql(user *models.User, profile *models.ChangeProfileModel) error {
	tx := db.Begin()
	err := tx.Model(user).Updates(profile.MapUser()).Error
	if err != nil {
		tx.Rollback()
		log.Warnf("Update user<%v> profile to %#v Mysql failed: %v", user.ID, profile, err)
		return err
	}

	user.Room.UserID = user.ID
	if tx.NewRecord(&user.Room) {
		err := tx.Create(&user.Room).Error
		if err != nil {
			tx.Rollback()
			log.Warnf("Create user %#v's room to Mysql failed: %v", user, err)
			return err
		}
	}
	
	err = tx.Model(&user.Room).Updates(profile.MapRoom()).Error
	if err != nil {
		tx.Rollback()
		log.Warnf("Update user<%v> profile to %#v Mysql failed: %v", user.ID, profile, err)
		return err
	}

	return tx.Commit().Error
}
