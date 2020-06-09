package http

import (
	"errors"
	"minitube/entities"
	jwt "minitube/middleware"
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
		if v, ok := data.(*entities.User); ok {
			return jwt.MapClaims{
				"username": v.Username,
			}
		}
		return jwt.MapClaims{}
	},

	IdentityHandler: func(c *gin.Context) interface{} {
		claims := jwt.ExtractClaims(c)
		return &entities.User{
			Username: claims["username"].(string),
		}
	},

	Authenticator: func(c *gin.Context) (interface{}, error) {
		var loginUser entities.User
		if err := c.ShouldBind(&loginUser); err != nil {
			return nil, jwt.ErrMissingLoginValues
		}
		username := loginUser.Username
		password := loginUser.Password

		log.Debugf("User %#v is logining in.", loginUser)
		user, err := store.GetUserByUsername(username)
		if err != nil {
			if errors.Is(err, store.ErrMySQLUserNotExists) {
				return nil, jwt.ErrFailedAuthentication
			}
			c.Error(err)
			return nil, err
		}

		log.Debugf("User %#v need auth to %#v", loginUser, user)
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err == nil {
			return &entities.User{
				Username: username,
			}, nil
		}
		if errors.Is(err, bcrypt.ErrHashTooShort) {
			c.Error(err)
		}
		return nil, jwt.ErrFailedAuthentication
	},

	Unauthorized: func(c *gin.Context, code int, message string) {
		c.JSON(code, gin.H{
			"code":    code,
			"message": message,
		})
	},
})
