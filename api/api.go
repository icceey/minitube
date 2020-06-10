package api

import (
	"errors"
	"io/ioutil"
	"minitube/entities"
	"minitube/middleware"
	"minitube/store"
	"minitube/utils"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Router - gin global router
var Router *gin.Engine

var log = utils.Sugar

func init() {
	Router = gin.New()

	Router.Use(middleware.Ginzap(utils.Logger, time.RFC3339, true))
	Router.Use(middleware.RecoveryWithZap(utils.Logger, true))

	Router.POST("/register", register)
	Router.POST("/login", authMiddleware.LoginHandler)
	Router.POST("/refresh/", authMiddleware.RefreshHandler)
	Router.POST("/logout", authMiddleware.LogoutHandler)

	streamGroup := Router.Group("/stream")
	streamGroup.Use(authMiddleware.MiddlewareFunc())
	streamGroup.GET("/key/:username", getStreamKey)

}

func register(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	log.Debugf("User register <%v> <%v>", username, password)

	_, err := store.GetUserByUsername(username)
	if err == nil {
		c.JSON(200, gin.H{
			"code": 1,
			"message": "Username already exists.",
		})
		return
	}
	if !errors.Is(err, store.ErrRedisUserNotExists) && !errors.Is(err, store.ErrMySQLUserNotExists) {
		c.Error(err)	
		c.JSON(500, gin.H{
			"code": 9,
			"message": "Server Error.",
		})
		return
	}

	passwordEncrypted, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		c.Error(err)
		c.JSON(500, gin.H{
			"code": 9,
			"message": "Server Error.",
		})
		return
	}

	log.Debug("username: ", username, "passwordEncrypted: ", string(passwordEncrypted))
	err = store.SaveUser(&entities.User{Username: username, Password: string(passwordEncrypted)})
	if err != nil {
		c.Error(err)
		c.JSON(500, gin.H{
			"code": 9,
			"message": "Server Error.",
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    0,
		"message": "OK",
	})
}

func getStreamKey(c *gin.Context) {
	username := c.Param("username")
	key := getStreamKeyFromLive(c, username)
	c.JSON(200, gin.H{
		"key": key,
	})
}

func getStreamKeyFromLive(c *gin.Context, username string) string {
	resp, err := http.Get("http://live:8090/control/get?room=" + username)
	if err != nil {
		c.Error(err)
		return ""
	}
	resBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.Error(err)
		return ""
	}
	str := string(resBytes)
	key := str[strings.LastIndexByte(str, ':')+2 : len(str)-2]
	return key
}
