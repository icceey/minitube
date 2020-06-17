package api

import (
	"errors"
	"io/ioutil"
	"minitube/middleware"
	"minitube/models"
	"minitube/store"
	"minitube/utils"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Router - gin global router
var Router *gin.Engine

var log = utils.Sugar

func init() {
	if debug, ok := os.LookupEnv("DEBUG"); ok && debug == "true" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	Router = gin.New()

	Router.Use(middleware.Ginzap(utils.Logger, time.RFC3339, true))
	Router.Use(middleware.RecoveryWithZap(utils.Logger, true))

	Router.LoadHTMLFiles("./out/index.html", "./out/watch.html",
		"./out/login.html", "./out/register.html", "./out/404.html")
	Router.Static("/_next/static", "./out/_next/static")
	Router.StaticFile("/favicon.ico", "./out/favicon.ico")

	Router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	Router.GET("/index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	Router.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})
	Router.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", nil)
	})
	Router.GET("/live/:username", func(c *gin.Context) {
		c.HTML(http.StatusOK, "watch.html", nil)
	})
	Router.NoRoute(func (c *gin.Context) {
		c.HTML(http.StatusNotFound, "404.html", nil)
	})


	Router.POST("/register", register)
	Router.POST("/login", authMiddleware.LoginHandler)
	Router.POST("/refresh", authMiddleware.RefreshHandler)
	Router.POST("/logout", authMiddleware.LogoutHandler)

	userGroup := Router.Group("/user")
	userGroup.Use(authMiddleware.MiddlewareFunc())
	userGroup.GET("/me", getMe)
	userGroup.POST("/profile", updateUserProfile)
	userGroup.POST("/password", changePassword)

	streamGroup := Router.Group("/stream")
	streamGroup.Use(authMiddleware.MiddlewareFunc())
	streamGroup.GET("/key/:username", getStreamKey)

}

func getMe(c *gin.Context) {
	id, ok := getUserID(c)
	if !ok {
		return
	}

	user, err := store.GetUserByID(id)
	if err != nil {
		if errors.Is(err, store.ErrRedisUserNotExists) || errors.Is(err, store.ErrMySQLUserNotExists) {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": "User not exists.",
			})
			return
		}
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Server Error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"user": models.GetMeFromUser(user),
	})
}

func register(c *gin.Context) {
	user := new(models.RegisterModel)
	if err := c.ShouldBind(user); err != nil {
		log.Debug(err)
		c.JSON(http.StatusNotAcceptable, gin.H{
			"code":    http.StatusNotAcceptable,
			"message": "invalid felid",
		})
		return
	}

	log.Debugf("User register <%#v>", user)
	_, err = store.GetUserByUsername(user.Username)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"code":    http.StatusConflict,
			"message": "username already exists",
		})
		return
	}
	if !errors.Is(err, store.ErrRedisUserNotExists) && !errors.Is(err, store.ErrMySQLUserNotExists) {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Server Error",
		})
		return
	}

	passwordEncrypted, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Server Error.",
		})
		return
	}

	user.Password = string(passwordEncrypted)
	log.Debugf("User register <%#v>", user)
	err = store.SaveUser(models.NewUserFromRegister(user))
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Server Error.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "OK",
	})
}

func getStreamKey(c *gin.Context) {
	username := c.Param("username")
	key := getStreamKeyFromLive(c, username)
	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"key":  key,
	})
}

func getStreamKeyFromLive(c *gin.Context, username string) string {
	url := "http://" + os.Getenv("LIVE_ADDR") + "/control/get?room=" + username
	resp, err := http.Get(url)
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

func updateUserProfile(c *gin.Context) {
	id, ok := getUserID(c)
	if !ok {
		return
	}

	profile := new(models.ChangeProfileModel)
	if err := c.ShouldBind(profile); err != nil {
		log.Debug(err)
		c.JSON(http.StatusNotAcceptable, gin.H{
			"code":    http.StatusNotAcceptable,
			"message": "invalid felid",
		})
		return
	}

	err := store.UpdateUserProfile(id, profile)
	if err != nil {
		if errors.Is(err, store.ErrRedisUserNotExists) || errors.Is(err, store.ErrMySQLUserNotExists) {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": "Username not exists",
			})
			return
		}
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Server Error.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "OK",
	})
}

func changePassword(c *gin.Context) {
	id, ok := getUserID(c)
	if !ok {
		return
	}

	pass := new(models.ChangePasswordModel)
	if err := c.ShouldBind(pass); err != nil {
		log.Debug(err)
		c.JSON(http.StatusNotAcceptable, gin.H{
			"code":    http.StatusNotAcceptable,
			"message": "invalid felid",
		})
		return
	}

	user, err := store.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "User not exists",
		})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(pass.OldPassword))
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "Password is wrong",
		})
		return 
	}

	passwordEncrypted, err := bcrypt.GenerateFromPassword([]byte(pass.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Server Error.",
		})
		return
	}

	err = store.ChangePassword(user, string(passwordEncrypted))
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Server Error.",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "OK",
	})
}

func getUserID(c *gin.Context) (uint, bool) {
	claims := middleware.ExtractClaims(c)

	i, exists := claims[authMiddleware.IdentityKey]
	id, ok := i.(float64)
	if !exists || !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "Bad Token.",
		})
		return 0, false
	}
	return uint(id), true
}
