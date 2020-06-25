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
	"strconv"
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

	Router.LoadHTMLFiles("./out/index.html", "./out/live/[streamer].html", "./out/mine.html",
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
	Router.GET("/mine", func(c *gin.Context) {
		c.HTML(http.StatusOK, "mine.html", nil)
	})
	Router.GET("/live/:username", func(c *gin.Context) {
		if id, ok := getUserID(c); ok {
			go store.UpdateWatchHistory(id, c.Param("username"))
		}
		c.HTML(http.StatusOK, "[streamer].html", nil)
	})
	Router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "404.html", nil)
	})

	Router.POST("/register", register)
	Router.POST("/login", authMiddleware.LoginHandler)
	Router.POST("/refresh", authMiddleware.RefreshHandler)
	Router.POST("/logout", authMiddleware.LogoutHandler)

	Router.GET("/followers/:username", getFollowers)
	Router.GET("/followings/:username", getFollowings)
	Router.GET("/profile/:username", getPublicUser)
	Router.GET("/living/:num", getLivingList)

	userGroup := Router.Group("/user")
	userGroup.Use(authMiddleware.MiddlewareFunc())
	userGroup.GET("/me", getMe)
	userGroup.POST("/profile", updateUserProfile)
	userGroup.POST("/password", changePassword)
	userGroup.POST("/follow/:username", follow)
	userGroup.POST("/unfollow/:username", unFollow)
	userGroup.GET("/history", getHistory)

	streamGroup := Router.Group("/stream")
	streamGroup.Use(authMiddleware.MiddlewareFunc())
	streamGroup.GET("/key/:username", getStreamKey)

}

func getFollowers(c *gin.Context) {
	getFollows(c, true)
}

func getFollowings(c *gin.Context) {
	getFollows(c, false)
}

func getFollows(c *gin.Context, followers bool) {
	username := c.Param("username")
	_, err := store.GetUserByUsername(username)
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

	var usernameList []string
	if followers {
		usernameList, err = store.GetFollowersFromRedis(username)
	} else {
		usernameList, err = store.GetFollowingsFromRedis(username)
	}
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Server Error",
		})
		return
	}

	me, _ := getUsername(c)
	userList := make([]*models.PublicUser, 0)
	for _, username := range usernameList {
		user, err := store.GetUserByUsername(username)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": "Server Error",
			})
			return
		}
		userList = append(userList, store.NewPublicUserFromUser(me, user))
	}

	if followers {
		c.JSON(http.StatusOK, gin.H{
			"code":      http.StatusOK,
			"followers": userList,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code":       http.StatusOK,
			"followings": userList,
		})
	}
}

func follow(c *gin.Context) {
	followOrNot(c, true)
}

func unFollow(c *gin.Context) {
	followOrNot(c, false)
}

func followOrNot(c *gin.Context, follow bool) {
	username, ok := getUsernameWithError(c)
	if !ok {
		return
	}

	dstUsername := c.Param("username")
	if username == dstUsername {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "Can't follow or unfollow yourself.",
		})
		return
	}

	_, err := store.GetUserByUsername(dstUsername)
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

	if follow {
		err = store.FollowUserInRedis(username, dstUsername)
	} else {
		err = store.UnFollowUserInRedis(username, dstUsername)
	}
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Server Error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "OK",
	})
}

func getHistory(c *gin.Context) {
	id, ok := getUserIDWithError(c)
	if !ok {
		return
	}

	history, err := store.GetWatchHistory(id)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Server Error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"history": history,
	})
}

func getPublicUser(c *gin.Context) {
	username := c.Param("username")

	user, err := store.GetUserByUsername(username)
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

	me, _ := getUsername(c)
	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"user": store.NewPublicUserFromUser(me, user),
	})
}

func getLivingList(c *gin.Context) {
	numStr := c.Param("num")

	num, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil || num <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "param not correct.",
		})
		return
	}

	if num > 24 {
		num = 24
	}

	userList, err := store.GetLivingUserList(num)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Server Error.",
		})
		return
	}

	username, _ := getUsername(c)
	model := store.NewLivingListModelFromUserList(username, userList)
	c.JSON(http.StatusOK, gin.H{
		"code":  http.StatusOK,
		"total": model.Total,
		"users": model.Users,
	})
}

func getMe(c *gin.Context) {
	id, ok := getUserIDWithError(c)
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
	if user.Email != "" {
		_, err = store.GetUserByEmail(user.Email)
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{
				"code":    http.StatusConflict,
				"message": "Email has been used",
			})
			return
		}
	}
	if user.Phone != "" {
		_, err = store.GetUserByPhone(user.Phone)
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{
				"code":    http.StatusConflict,
				"message": "Phone has been used",
			})
			return
		}
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
	id, ok := getUserIDWithError(c)
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
	id, ok := getUserIDWithError(c)
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

func getUserIDWithError(c *gin.Context) (uint, bool) {
	claims := middleware.ExtractClaims(c)
	i, exists := claims[authMiddleware.IdentityKey]
	id, ok := i.(float64)
	if ok && exists {
		return uint(id), true
	}
	c.JSON(http.StatusBadRequest, gin.H{
		"code":    http.StatusBadRequest,
		"message": "Bad Token.",
	})
	return 0, false
}

func getUserID(c *gin.Context) (uint, bool) {
	claims, err := authMiddleware.GetClaimsFromJWT(c)
	if err != nil {
		return 0, false
	}

	i, exists := claims[authMiddleware.IdentityKey]
	id, ok := i.(float64)
	return uint(id), ok && exists
}

func getUsernameWithError(c *gin.Context) (string, bool) {
	claims := middleware.ExtractClaims(c)
	i, exists := claims["username"]
	username, ok := i.(string)
	if ok && exists {
		return username, true
	}
	c.JSON(http.StatusBadRequest, gin.H{
		"code":    http.StatusBadRequest,
		"message": "Bad Token.",
	})
	return "", false
}

func getUsername(c *gin.Context) (string, bool) {
	claims, err := authMiddleware.GetClaimsFromJWT(c)
	if err != nil {
		return "", false
	}

	i, exists := claims["username"]
	username, ok := i.(string)
	return username, ok && exists
}