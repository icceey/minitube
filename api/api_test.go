package api

import (
	"encoding/json"
	"io/ioutil"
	"minitube/models"
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
	invalidRegister  []*models.RegisterModel
	invalidLoginUser []*models.LoginModel
	wrongLoginUser   []*models.LoginModel
	validRegister    []*models.RegisterModel
	validLoginUser   []*models.LoginModel
	tokens           []string
)

type response struct {
	Code    int
	Message string

	Token  string
	Expire string

	Key string

	User *models.User
}

var (
	respLoginMissing    = response{Code: http.StatusUnauthorized, Message: "missing Username or Password"}
	respLoginIncorrect  = response{Code: http.StatusUnauthorized, Message: "incorrect Username or Password"}
	respRegisterExists  = response{Code: http.StatusConflict, Message: "username already exists"}
	respRegisterInvalid = response{Code: http.StatusNotAcceptable, Message: "invalid felid"}
	respRegisterOK      = response{Code: http.StatusOK, Message: "OK"}
	respTokenNotMatch   = response{Code: http.StatusForbidden, Message: "you don't have permission to access this resource"}
)

func TestRegister(t *testing.T) {
	require := require.New(t)

	// not has username or password will fail.
	var resp response
	body := postJSON(t, "/register", nil, "")
	err := json.Unmarshal(body, &resp)
	require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
	require.Equal(respRegisterInvalid, resp, "User {} shouldn't register success.")
	log.Info(resp)

	// Username or password not valid  will fail register.
	for _, user := range invalidRegister {
		var resp response
		body := postJSON(t, "/register", mapUser(user), "")
		err := json.Unmarshal(body, &resp)
		require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
		require.Equalf(respRegisterInvalid, resp, "User %#v shouldn't register success.", user)
	}

	// This should register success.
	for _, user := range validRegister {
		var resp response
		body := postJSON(t, "/register", mapUser(user), "")
		err := json.Unmarshal(body, &resp)
		require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
		require.Equalf(respRegisterOK, resp, "User %#v should register success.", user)
	}
}

func TestLogin(t *testing.T) {
	require := require.New(t)

	// Username or password not valid  will fail login.
	for _, user := range invalidLoginUser {
		var resp response
		body := postJSON(t, "/login", mapUser(user), "")
		err := json.Unmarshal(body, &resp)
		require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
		require.Equalf(respLoginIncorrect, resp, "User %#v shouldn't login success.", user)
	}

	// Username and password not match will fail.
	for _, user := range wrongLoginUser {
		var resp response
		body := postJSON(t, "/login", mapUser(user), "")
		err := json.Unmarshal(body, &resp)
		require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
		require.Equalf(respLoginIncorrect, resp, "User %#v shouldn't login success.", user)
	}

	// This should login success.
	for i, user := range validLoginUser {
		var resp response
		body := postJSON(t, "/login", mapUser(user), "")
		err := json.Unmarshal(body, &resp)
		require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
		require.Equal(http.StatusOK, resp.Code, "Login should return OK")
		require.Empty(resp.Message, "message should empty")
		require.NotEmpty(resp.Expire, "Expire shouldn't empty")
		require.NotEmpty(resp.Token, "Token shouldn't empty")
		tokens[i] = resp.Token
	}
}

func TestRefresh(t *testing.T) {
	require := require.New(t)

	// Refresh token
	for i, token := range tokens {
		var resp response
		body := postJSON(t, "/refresh", nil, token)
		err := json.Unmarshal(body, &resp)
		require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
		require.Equal(http.StatusOK, resp.Code, "Refresh should return OK")
		require.Empty(resp.Message, "message should empty")
		require.NotEmpty(resp.Expire, "Expire shouldn't empty")
		require.NotEmpty(resp.Token, "Token shouldn't empty")
		tokens[i] = resp.Token
	}
}

func TestLogout(t *testing.T) {
	require := require.New(t)

	// Logout
	for _, token := range tokens {
		var resp response
		body := postJSON(t, "/logout", nil, token)
		err := json.Unmarshal(body, &resp)
		require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
		require.Equal(http.StatusOK, resp.Code, "Logout should return OK")
	}
}

func TestGetMe(t *testing.T) {
	require := require.New(t)

	// Get my info.
	for i, user := range validRegister {
		var resp response
		body := get(t, "/user/me", tokens[i])
		err := json.Unmarshal(body, &resp)
		require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
		require.Equal(http.StatusOK, resp.Code, "Get Info should return OK")
		require.Equal(user.Username, resp.User.Username, "Username Not Equal.")
		require.Empty(resp.User.Password, "Password should empty.")
	}

}

func TestGetStreamKey(t *testing.T) {
	require := require.New(t)

	// Get stream key
	for i, user := range validRegister {
		var resp response
		body := get(t, "/stream/key/"+user.Username, tokens[i])
		err := json.Unmarshal(body, &resp)
		require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
		require.Equal(http.StatusOK, resp.Code, "Get stream key should return OK")
		require.Empty(resp.Message, "message should empty")
		require.NotEmpty(resp.Key, "Key shouldn't empty")
	}

	// When username and token not match, key will not return
	for i, user := range validRegister {
		var resp response
		token := tokens[0]
		num := len(validRegister)
		token = tokens[(i+3)%num]
		body := get(t, "/stream/key/"+user.Username, token)
		err := json.Unmarshal(body, &resp)
		require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
		require.Equalf(respTokenNotMatch, resp, "User %#v shouldn't get key.", string(body))
	}
}

func postJSON(t *testing.T, uri string, mp map[string]string, token string) []byte {
	rec := httptest.NewRecorder()

	var req *http.Request
	if mp != nil {
		jsonBytes, err := json.Marshal(mp)
		require.NoErrorf(t, err, "Json Marshal error <%v>", mp)
		req = httptest.NewRequest("POST", uri, strings.NewReader(string(jsonBytes)))
	} else {
		req = httptest.NewRequest("POST", uri, nil)
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "MiniTube "+token)
	}

	Router.ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()
	log.Info("===========>", resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	require.NoErrorf(t, err, "Request %v shouldn't has error.", uri)
	return body
}

func postForm(t *testing.T, uri string, form url.Values, token string) []byte {
	rec := httptest.NewRecorder()

	var req *http.Request
	if form != nil {
		req = httptest.NewRequest("POST", uri, strings.NewReader(form.Encode()))
	} else {
		req = httptest.NewRequest("POST", uri, nil)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if token != "" {
		req.Header.Set("Authorization", "MiniTube "+token)
	}

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
	if token != "" {
		req.Header.Set("Authorization", "MiniTube "+token)
	}

	Router.ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	require.NoErrorf(t, err, "Request %v shouldn't has error.", uri)
	return body
}

func mapUser(u interface{}) map[string]string {
	if user, ok := u.(*models.LoginModel); ok {
		return map[string]string{
			"username": user.Username,
			"password": user.Password,
			"email":    user.Email,
			"phone":    user.PhoneNumber,
		}
	}
	if user, ok := u.(*models.RegisterModel); ok {
		return map[string]string{
			"username": user.Username,
			"password": user.Password,
			"email":    user.Email,
			"phone":    user.PhoneNumber,
		}
	}
	return nil
}

func valUser(u interface{}) url.Values {
	if user, ok := u.(*models.LoginModel); ok {
		return url.Values{
			"username": []string{user.Username},
			"password": []string{user.Password},
			"email":    []string{user.Email},
			"phone":    []string{user.PhoneNumber},
		}
	}
	if user, ok := u.(*models.RegisterModel); ok {
		return url.Values{
			"username": []string{user.Username},
			"password": []string{user.Password},
			"email":    []string{user.Email},
			"phone":    []string{user.PhoneNumber},
		}
	}
	return nil
}

func createUserForTest() {
	invalidRegister = []*models.RegisterModel{
		{},
		{Username: "1"},
		{Password: "1"},
		{Email: "a@b.com"},
		{PhoneNumber: "13668686868"},
		{Username: "1@1", Password: "c96d84ba3b4a823c4fee088a7369a5c02f50ef40f9ca54bdec34843eba157132"},
		{Username: "1234567890123456789 ", Password: "c96d84ba3b4a823c4fee088a7369a5c02f50ef40f9ca54bdec34843eba157132"},
		{Username: "123456789012345678900", Password: "c96d84ba3b4a823c4fee088a7369a5c02f50ef40f9ca54bdec34843eba157132"},
		{Username: "12345678901234567890", Password: "c96d84ba3b4a823c4fee088a7369a5c02f50ef40f9ca54bdec34843eba15713"},
		{Username: "12345678901234567890", Password: "c96d84ba3b4a823c4fee088a7369a5c02f50ef40f9ca54bdec34843eba1571321"},
		{Username: "12345678901234567890", Password: "c96d84ba3b4a823c4fee088a7369a5c02f50ef40f9ca54bdec34843eba157@3"},
	}
	invalidLoginUser = []*models.LoginModel{
		{},
		{Username: "1"},
		{Password: "1"},
		{Email: "a@b.com"},
		{PhoneNumber: "+8613668686868"},
		{Username: "1@1", Password: "c96d84ba3b4a823c4fee088a7369a5c02f50ef40f9ca54bdec34843eba157132"},
		{Username: "1234567890123456789 ", Password: "c96d84ba3b4a823c4fee088a7369a5c02f50ef40f9ca54bdec34843eba157132"},
		{Username: "123456789012345678900", Password: "c96d84ba3b4a823c4fee088a7369a5c02f50ef40f9ca54bdec34843eba157132"},
		{Username: "12345678901234567890", Password: "c96d84ba3b4a823c4fee088a7369a5c02f50ef40f9ca54bdec34843eba15713"},
		{Username: "12345678901234567890", Password: "c96d84ba3b4a823c4fee088a7369a5c02f50ef40f9ca54bdec34843eba1571321"},
		{Username: "12345678901234567890", Password: "c96d84ba3b4a823c4fee088a7369a5c02f50ef40f9ca54bdec34843eba157@3"},
	}
	wrongLoginUser = []*models.LoginModel{
		{Username: "111", Password: "c96d84ba3b4a823c4fee088a7369a5c02f50ef40f9ca54bdec34843eba157132"},
		{Username: "112", Password: "c96d84ba3b4a823c4fee088a7369a5c02f50ef40f9ca54bdec34843eba157132"},
		{Username: "121", Password: "c96d84ba3b4a823c4fee088a7369a5c02f50ef40f9ca54bdec34843eba157131"},
		{Username: "122", Password: "fca26135ea43ad0ba904e62c85793768e4c9a136d2660bb2b45952ad445f5921"},
	}
	validLoginUser = []*models.LoginModel{
		{Username: "121", Password: "c96d84ba3b4a823c4fee088a7369a5c02f50ef40f9ca54bdec34843eba157132"},
		{Username: "122", Password: "fca26135ea43ad0ba904e62c85793768e4c9a136d2660bb2b15952ad445f5921"},
		{Email: "123@minitube.com", Password: "fca26135ea43ad0ba904e62c85793768e4c9af40f9ca54bdec34843eba157132"},
		{PhoneNumber: "+8612468686868", Password: "fca26135ea43ad0ba9f40f9ca54bdec34843eba157132bb2b15952ad445f5921"},
		{Username: "125", Password: "fca261f40f9ca54bdec34843eba1571324c9a136d2660bb2b15952ad445f5921"},
		{Email: "125@minitube.com", Password: "fca261f40f9ca54bdec34843eba1571324c9a136d2660bb2b15952ad445f5921"},
		{PhoneNumber: "+8612568686868", Password: "fca261f40f9ca54bdec34843eba1571324c9a136d2660bb2b15952ad445f5921"},
	}
	validRegister = []*models.RegisterModel{
		{Username: "121", Password: "c96d84ba3b4a823c4fee088a7369a5c02f50ef40f9ca54bdec34843eba157132"},
		{Username: "122", Password: "fca26135ea43ad0ba904e62c85793768e4c9a136d2660bb2b15952ad445f5921"},
		{Username: "123", Email: "123@minitube.com", Password: "fca26135ea43ad0ba904e62c85793768e4c9af40f9ca54bdec34843eba157132"},
		{Username: "124", PhoneNumber: "+8612468686868", Password: "fca26135ea43ad0ba9f40f9ca54bdec34843eba157132bb2b15952ad445f5921"},
		{Username: "125", Email: "125@minitube.com", PhoneNumber: "+8612568686868", Password: "fca261f40f9ca54bdec34843eba1571324c9a136d2660bb2b15952ad445f5921"},
	}
	tokens = make([]string, len(validLoginUser))
}

func TestMain(m *testing.M) {
	createUserForTest()
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
