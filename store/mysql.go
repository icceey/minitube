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
	log.Info("Initialize mysql connection pool...")
	var err error
	dataSourceName := fmt.Sprintf("%v:%v@tcp(%v)/%v?charset=utf8&parseTime=True&loc=Local",
		os.Getenv("MYSQL_USER"), os.Getenv("MYSQL_PASSWORD"), os.Getenv("MYSQL_ADDR"), os.Getenv("MYSQL_DATABASE"))
	db, err = gorm.Open("mysql", dataSourceName)
	if err != nil {
		log.Fatal("MySQL open failed: ", err)
	}

	db.SingularTable(true)
	db.DB().SetConnMaxLifetime(0)
	db.DB().SetMaxIdleConns(3)
	db.DB().SetMaxOpenConns(3)

	log.Info("Checking MySQL service...")
	retry, interval := 5, 10
	for i := 0; i < retry; i++ {
		err = pingMySQL()
		if err == nil {
			break
		}
		errMsg := fmt.Sprintf("Ping MySQL failed %v times, ", i+1)
		if i == retry-1 {
			errMsg += "maybe some errors have occurred."
			log.Fatal(errMsg, "MySQL service access failed: ", err)
		} else {
			errMsg += fmt.Sprintf("maybe mysql container is not ready? Will retry after %v seconds", interval)
			log.Warn(errMsg)
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}

	db.AutoMigrate(&models.User{})

	log.Info("MySQL is OK.")
}

func pingMySQL() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return db.DB().PingContext(ctx)
}

func getUserByUsernameFromMysql(username string) (*models.User, error) {
	return getUserFromMysqlBy("username", username)
}

func getUserByEmailFromMysql(email string) (*models.User, error) {
	return getUserFromMysqlBy("email", email)
}

func getUserByPhoneFromMysql(phone string) (*models.User, error) {
	return  getUserFromMysqlBy("phone_number", phone)
}

func getUserFromMysqlBy(by, value string) (*models.User, error) {
	user := new(models.User)
	err := db.Where(by + " = ?", value).Take(user).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, ErrMySQLUserNotExists
		}
		log.Warnf("Get user %v from Mysql failed: %v", value, err)
		return nil, ErrMySQLFailed
	}
	return user, nil
}

func saveUserToMysql(user *models.User) error {
	if db.NewRecord(user) {
		log.Debugf("%#v", user)
		err := db.Create(user).Error
		if err != nil {
			log.Warnf("Save user %#v to Mysql failed: %v", user, err)
			return err
		}
		return nil
	}
	return nil
}
