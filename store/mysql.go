package store

import (
	"context"
	"database/sql"
	"fmt"
	"minitube/entities"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql" // mysql driver
)


var mysqlPool *sql.DB
var userQueryStmt *sql.Stmt


func init() {
	log.Info("Initialize mysql connection pool...")
	var err error
	dataSourceName := fmt.Sprintf("%v:%v@tcp(%v)/%v", os.Getenv("MYSQL_USER"), 
		os.Getenv("MYSQL_PASSWORD"), os.Getenv("MYSQL_ADDR"), os.Getenv("MYSQL_DATABASE"))
	mysqlPool, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Fatal("MySQL open failed: ", err)
	}

	mysqlPool.SetConnMaxLifetime(0)
	mysqlPool.SetMaxIdleConns(3)
	mysqlPool.SetMaxOpenConns(3)

	log.Info("Checking MySQL service...")
	for i := 0; i < 5; i++ {
		time.Sleep(5 * time.Second)
		err = pingMySQL()
		if err == nil {
			break
		}
		log.Warnf("Ping MySQL failed %v times, may be mysql container is not ready.", i+1)
	}
	if err != nil {
		log.Fatal("MySQL service access failed: ", err)
	}

	createTableIfNot()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	userQueryStmt, err = mysqlPool.PrepareContext(ctx, "SELECT username, password FROM user WHERE username = ?")
	if err != nil {
		log.Fatal("MySQL user query statement prepare failed: ", err)
	}


	log.Info("MySQL is OK.")
}


func pingMySQL() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return mysqlPool.PingContext(ctx)
}


func createTableIfNot() {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_, err := mysqlPool.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS user (
		id INT AUTO_INCREMENT PRIMARY KEY,
		username VARCHAR(20) UNIQUE,
		password CHAR(64) NOT NULL
	)`)
	if err != nil {
		log.Fatal(err.Error())
	}
}


func getUserByUsernameFromMysql(username string) (*entities.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	row := userQueryStmt.QueryRowContext(ctx, username)
	user := new(entities.User)
	err := row.Scan(&user.Username, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrMySQLUserNotExists
		}
		log.Warnf("Get user %v from Mysql failed: %v", username, err)
		return nil, ErrMySQLFailed
	}
	return user, nil
}


func saveUserToMysql(user *entities.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_, err := mysqlPool.ExecContext(ctx, "INSERT user (username,password) VALUES(?,?)", user.Username, user.Password)
	if err != nil {
		log.Warnf("Save user %#v to Mysql failed: %v", user, err)
		return err
	}
	return nil
}