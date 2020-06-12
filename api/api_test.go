package api

import (
	"encoding/json"
	"io/ioutil"
	"minitube/entities"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// This test needs redis and mysql service.
// Please run test with `test.sh`.

var (
	userNil = []*entities.User{
		entities.NewUser("", ""),
		entities.NewUser("11", ""),
		entities.NewUser("", "11"),
	}
	userInvalid = []*entities.User{
		entities.NewUser("11", "c96d84ba3b4a823c4fee088a7369a5c02f50ef40f9ca54bdec34843eba15713"),
		entities.NewUser("11", "c96d84ba3b4a823c4fee088a7369a5c02f50ef40f9ca54bdec34843eba1571321"),
		entities.NewUser("11", "c96d84ba3b4a823c4fee088a7369a5c02f50ef40f9ca54bdec34843eba15713g"),
		entities.NewUser("11", "c96d84ba3b4a823c4fee088a7369a5c02f50ef40f9ca54bdec34843eba15713@"),
		entities.NewUser("@", "c96d84ba3b4a823c4fee088a7369a5c02f50ef40f9ca54bdec34843eba157132"),
		entities.NewUser("-", "c96d84ba3b4a823c4fee088a7369a5c02f50ef40f9ca54bdec34843eba157132"),
		entities.NewUser("_", "c96d84ba3b4a823c4fee088a7369a5c02f50ef40f9ca54bdec34843eba157132"),
		entities.NewUser("1234567890123456789 ", "c96d84ba3b4a823c4fee088a7369a5c02f50ef40f9ca54bdec34843eba157132"),
		entities.NewUser("123456789012345678900", "c96d84ba3b4a823c4fee088a7369a5c02f50ef40f9ca54bdec34843eba157132"),
	}
	userWrong = []*entities.User{
		entities.NewUser("11", "c96d84ba3b4a823c4fee088a7369a5c02f50ef40f9ca54bdec34843eba157132"),
		entities.NewUser("12", "c96d84ba3b4a823c4fee088a7369a5c02f50ef40f9ca54bdec34843eba157132"),
		entities.NewUser("21", "c96d84ba3b4a823c4fee088a7369a5c02f50ef40f9ca54bdec34843eba157131"),
		entities.NewUser("22", "fca26135ea43ad0ba904e62c85793768e4c9a136d2660bb2b45952ad445f5921"),
	}
	userOK = []*entities.User{
		entities.NewUser("21", "c96d84ba3b4a823c4fee088a7369a5c02f50ef40f9ca54bdec34843eba157132"),
		entities.NewUser("22", "fca26135ea43ad0ba904e62c85793768e4c9a136d2660bb2b15952ad445f5921"),
	}
)

type response struct {
	Code    int
	Message string
	Token   string
	Expire  string
	Key     string
}

var (
	respLoginMissing    = response{Code: http.StatusUnauthorized, Message: "missing Username or Password"}
	respLoginIncorrect  = response{Code: http.StatusUnauthorized, Message: "incorrect Username or Password"}
	respRegisterExists  = response{Code: http.StatusConflict, Message: "username already exists"}
	respRegisterInvalid = response{Code: http.StatusNotAcceptable, Message: "invalid username or password"}
	respRegisterOK      = response{Code: http.StatusOK, Message: "OK"}
	respTokenNotMatch   = response{Code: http.StatusForbidden, Message: "you don't have permission to access this resource"}
)

func TestRegister(t *testing.T) {
	require := require.New(t)

	// not has username or password will fail.
	var resp response
	body := postJSON(t, "/register", map[string]string{})
	err := json.Unmarshal(body, &resp)
	require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
	require.Equal(respRegisterInvalid, resp, "User {} shouldn't register success.")
	log.Info(resp)

	// Don't have username or password will register fail.
	for _, user := range userNil {
		var resp response
		body := postJSON(t, "/register", mapUser(user))
		err := json.Unmarshal(body, &resp)
		require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
		require.Equalf(respRegisterInvalid, resp, "User %#v shouldn't register success.", user)
		log.Info(resp)
	}

	// Username should only have english letters and number
	// Password will hash by frontend, so it should be hex number string with length 64.
	// Username or password not valid  will fail register.
	for _, user := range userInvalid {
		var resp response
		body := postJSON(t, "/register", mapUser(user))
		err := json.Unmarshal(body, &resp)
		require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
		require.Equalf(respRegisterInvalid, resp, "User %#v shouldn't register success.", user)
		log.Info(resp)
	}

	// This should register success.
	for _, user := range userOK {
		var resp response
		body := postJSON(t, "/register", mapUser(user))
		err := json.Unmarshal(body, &resp)
		require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
		require.Equalf(respRegisterOK, resp, "User %#v should register success.", user)
		log.Info(resp)
	}

}

func TestLogin(t *testing.T) {
	require := require.New(t)

	// Don't have username or password will login fail.
	for _, user := range userNil {
		var resp response
		body := postJSON(t, "/login", mapUser(user))
		err := json.Unmarshal(body, &resp)
		require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
		require.Equalf(respLoginMissing, resp, "User %#v shouldn't login success.", user)
	}

	// Username should only have english letters and number
	// Password will hash by frontend, so it should be hex number string with length 64.
	// Username or password not valid  will fail login.
	for _, user := range userInvalid {
		var resp response
		body := postJSON(t, "/login", mapUser(user))
		err := json.Unmarshal(body, &resp)
		require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
		require.Equalf(respLoginIncorrect, resp, "User %#v shouldn't login success.", user)
	}

	// Username and password not match will fail.
	for _, user := range userWrong {
		var resp response
		body := postJSON(t, "/login", mapUser(user))
		err := json.Unmarshal(body, &resp)
		require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
		require.Equalf(respLoginIncorrect, resp, "User %#v shouldn't login success.", user)
	}

	// This should login success.
	for _, user := range userOK {
		var resp response
		body := postJSON(t, "/login", mapUser(user))
		err := json.Unmarshal(body, &resp)
		require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
		require.Equal(http.StatusOK, resp.Code, "Login should return OK")
		require.Empty(resp.Message, "message should empty")
		require.Len(resp.Expire, 25, "expire time string should length 25")
		require.Len(resp.Token, 156, "token length should be 156")
	}
}

func TestGetStreamKey(t *testing.T) {
	require := require.New(t)

	// Login get token
	tokens := make([]string, len(userOK))
	for i, user := range userOK {
		var resp response
		body := postJSON(t, "/login", mapUser(user))
		err := json.Unmarshal(body, &resp)
		require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
		require.Equal(http.StatusOK, resp.Code, "Login should return OK")
		require.Empty(resp.Message, "message should empty")
		require.Len(resp.Expire, 25, "expire time string should length 25")
		require.Len(resp.Token, 156, "token length should be 156")
		tokens[i] = resp.Token
	}

	// Get stream key
	for i, user := range userOK {
		var resp response
		body := get(t, "/stream/key/"+user.Username, tokens[i])
		err := json.Unmarshal(body, &resp)
		require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
		require.Equal(http.StatusOK, resp.Code, "Login should return OK")
		require.Empty(resp.Message, "message should empty")
		require.NotEmpty(resp.Key, "Key shouldn't empty")
	}

	// When username and token not match, key will not return
	for i, user := range userOK {
		var resp response
		token := tokens[0]
		if i < len(userOK)-1 {
			token = tokens[i+1]
		}
		body := get(t, "/stream/key/"+user.Username, token)
		err := json.Unmarshal(body, &resp)
		require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
		require.Equalf(respTokenNotMatch, resp, "User %#v shouldn't get key.", string(body))
	}
}

func postJSON(t *testing.T, uri string, mp map[string]string) []byte {
	jsonBytes, err := json.Marshal(mp)
	require.NoErrorf(t, err, "Json Marshal error <%v>", mp)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", uri, strings.NewReader(string(jsonBytes)))
	req.Header.Set("Content-Type", "application/json")

	Router.ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	require.NoErrorf(t, err, "Request %v shouldn't has error.", uri)
	return body
}

func postForm(t *testing.T, uri string, form url.Values) []byte {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", uri, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	Router.ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	require.NoErrorf(t, err, "Request %v shouldn't has error.", uri)
	return body
}

func get(t *testing.T, uri string, token string) []byte {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", uri, nil)
	req.Header.Set("Authorization", "MiniTube "+token)

	Router.ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	require.NoErrorf(t, err, "Request %v shouldn't has error.", uri)
	return body
}

func mapUser(user *entities.User) map[string]string {
	return map[string]string{
		"username": user.Username,
		"password": user.Password,
	}
}

func valUser(user *entities.User) url.Values {
	return url.Values{
		"username": []string{user.Username},
		"password": []string{user.Password},
	}
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
