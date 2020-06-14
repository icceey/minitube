package api

import (
	"errors"
	jwt "minitube/middleware"
	"minitube/models"
	"minitube/store"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var authMiddleware, err = jwt.New(&jwt.GinJWTMiddleware{
	Realm:         "MiniTube",
	Key:           []byte(os.Getenv("JWT_SECRET_KEY")),
	Timeout:       time.Hour,
	MaxRefresh:    24 * time.Hour,
	IdentityKey:   "username",
	TokenHeadName: "MiniTube",

	PayloadFunc: func(data interface{}) jwt.MapClaims {
		if v, ok := data.(*models.User); ok {
			return jwt.MapClaims{
				"username": v.Username,
			}
		}
		return jwt.MapClaims{}
	},

	IdentityHandler: func(c *gin.Context) interface{} {
		claims := jwt.ExtractClaims(c)
		return &models.User{
			Username: claims["username"].(string),
		}
	},

	Authenticator: func(c *gin.Context) (interface{}, error) {
		loginUser := new(models.LoginModel)
		if err := c.ShouldBind(loginUser); err != nil {
			log.Debug(err)
			return nil, jwt.ErrFailedAuthentication
		}

		log.Debugf("User %#v is logining in.", loginUser)
		var user *models.User
		var err error
		if username := loginUser.Username; username != "" {
			user, err = store.GetUserByUsername(username)
		} else if email := loginUser.Email; email != "" {
			user, err = store.GetUserByEmail(email)
		} else if phone := loginUser.Phone; phone != "" {
			user, err = store.GetUserByPhone(phone)
		} else {
			err = errors.New("Login validator has some error")
		}
		if err != nil {
			if errors.Is(err, store.ErrMySQLUserNotExists) {
				return nil, jwt.ErrFailedAuthentication
			}
			c.Error(err)
			return nil, err
		}

		// log.Debugf("User %#v need auth to %#v", loginUser, user)
		password := loginUser.Password
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err == nil {
			log.Debugf("User %#v auth success", user)
			return user, nil
		}
		if errors.Is(err, bcrypt.ErrHashTooShort) {
			c.Error(err)
		}
		return nil, jwt.ErrFailedAuthentication
	},
	Authorizator: func(data interface{}, c *gin.Context) bool {
		if user, ok := data.(*models.User); ok {
			if c.FullPath() == "/stream/key/:username" {
				if user.Username != c.Param("username") {
					return false
				}
			}
			return true
		}
		return false
	},
	Unauthorized: func(c *gin.Context, code int, message string) {
		c.JSON(code, gin.H{
			"code":    code,
			"message": message,
		})
	},

	TokenLookup: "header: Authorization, query: token, cookie: token",

	SendCookie:     true,
	SecureCookie:   false,
	CookieHTTPOnly: true,
	CookieName:     "token",
})
