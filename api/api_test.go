package api

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"minitube/models"
	"minitube/store"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

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
	changeProfile    []map[string]string
	changePass       []map[string]string
	tokens           []string
)

type baseResponse struct {
	Code    int
	Message string
}

type tokenResponse struct {
	baseResponse
	Token  string
	Expire string
}

type meResponse struct {
	baseResponse
	User *models.Me
}

type keyResponse struct {
	baseResponse
	Key string
}

type liveResponse struct {
	baseResponse
	Total int
	Users []*models.PublicUser
}

type pubResponse struct {
	baseResponse
	User *models.PublicUser
}

type historyResponse struct {
	baseResponse
	History []*models.History
}

type followResponse struct {
	baseResponse
	Followers []*models.PublicUser
	Followings []*models.PublicUser
}

func TestRegister(t *testing.T) {
	require := require.New(t)

	respRegisterInvalid := baseResponse{Code: http.StatusNotAcceptable, Message: "invalid felid"}
	respRegisterOK := baseResponse{Code: http.StatusOK, Message: "OK"}

	// not has username or password will fail.
	var resp baseResponse
	body := postJSON(t, "/register", nil, "")
	err := json.Unmarshal(body, &resp)
	require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
	require.Equal(respRegisterInvalid, resp, "User {} shouldn't register success.")
	log.Info(resp)

	// Username or password not valid  will fail register.
	for _, user := range invalidRegister {
		var resp baseResponse
		body := postJSON(t, "/register", mapUser(user), "")
		err := json.Unmarshal(body, &resp)
		require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
		require.Equalf(respRegisterInvalid, resp, "User %#v shouldn't register success.", user)
	}

	// This should register success.
	for _, user := range validRegister {
		var resp baseResponse
		body := postJSON(t, "/register", mapUser(user), "")
		err := json.Unmarshal(body, &resp)
		require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
		require.Equalf(respRegisterOK, resp, "User %#v should register success.", user)
	}
}

func TestLogin(t *testing.T) {
	require := require.New(t)

	respLoginIncorrect := tokenResponse{
		baseResponse: baseResponse{
			Code:    http.StatusUnauthorized,
			Message: "incorrect Username or Password",
		},
	}

	// Username or password not valid  will fail login.
	for _, user := range invalidLoginUser {
		var resp tokenResponse
		body := postJSON(t, "/login", mapUser(user), "")
		err := json.Unmarshal(body, &resp)
		require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
		require.Equalf(respLoginIncorrect, resp, "User %#v shouldn't login success.", user)
	}

	// Username and password not match will fail.
	for _, user := range wrongLoginUser {
		var resp tokenResponse
		body := postJSON(t, "/login", mapUser(user), "")
		err := json.Unmarshal(body, &resp)
		require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
		require.Equalf(respLoginIncorrect, resp, "User %#v shouldn't login success.", user)
	}

	// This should login success.
	for i, user := range validLoginUser {
		var resp tokenResponse
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
		var resp tokenResponse
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
		var resp baseResponse
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
		var resp meResponse
		body := get(t, "/user/me", tokens[i])
		err := json.Unmarshal(body, &resp)
		require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
		require.Equal(http.StatusOK, resp.Code, "Get Info should return OK")
		require.Equal(user.Username, resp.User.Username, "Username Not Equal.")
	}

}

func TestGetStreamKey(t *testing.T) {
	require := require.New(t)

	respTokenNotMatch := keyResponse{
		baseResponse: baseResponse {
			Code: http.StatusForbidden, 
			Message: "you don't have permission to access this resource",
		},
	}

	// Get stream key
	for i, user := range validRegister {
		var resp keyResponse
		body := get(t, "/stream/key/"+user.Username, tokens[i])
		err := json.Unmarshal(body, &resp)
		require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
		require.Equal(http.StatusOK, resp.Code, "Get stream key should return OK")
		require.Empty(resp.Message, "message should empty")
		require.NotEmpty(resp.Key, "Key shouldn't empty")
	}

	// When username and token not match, key will not return
	for i, user := range validRegister {
		var resp keyResponse
		token := tokens[0]
		num := len(validRegister)
		token = tokens[(i+3)%num]
		body := get(t, "/stream/key/"+user.Username, token)
		err := json.Unmarshal(body, &resp)
		require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
		require.Equalf(respTokenNotMatch, resp, "User %#v shouldn't get key.", string(body))
	}
}

func TestUpdateUserProfile(t *testing.T) {
	require := require.New(t)

	for i := range validRegister {
		var resp baseResponse
		body := postJSON(t, "/user/profile", changeProfile[i], tokens[i])
		err := json.Unmarshal(body, &resp)
		require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
		require.Equal(http.StatusOK, resp.Code, "Get stream key should return OK")
		require.Equal("OK", resp.Message, "message should be OK")
	}

	for i, user := range validRegister {
		var resp meResponse
		body := get(t, "/user/me", tokens[i])
		err := json.Unmarshal(body, &resp)
		require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
		require.Equal(http.StatusOK, resp.Code, "Get Info should return OK")
		require.Equal(user.Username, resp.User.Username, "Username Not Equal.")
		if v, ok := changeProfile[i]["email"]; ok {
			if v == "" {
				require.Nil(resp.User.Email, "Email has deleted, should nil.")
			} else {
				require.Equal(v, *resp.User.Email, "Email has changed, should equal.")
			}
		}
		if v, ok := changeProfile[i]["phone"]; ok {
			if v == "" {
				require.Nil(resp.User.Phone, "Phone has deleted, should nil.")
			} else {
				require.Equal(v, *resp.User.Phone, "Phone has changed, should equal.")
			}
		}
		if v, ok := changeProfile[i]["live_name"]; ok {
			if v == "" {
				require.Nil(resp.User.LiveName, "LiveName has deleted, should nil.")
			} else {
				require.Equal(v, *resp.User.LiveName, "LiveName has changed, should equal.")
			}
		}
		if v, ok := changeProfile[i]["live_intro"]; ok {
			if v == "" {
				require.Nil(resp.User.LiveIntro, "LiveIntro has deleted, should nil.")
			} else {
				require.Equal(v, *resp.User.LiveIntro, "LiveIntro has changed, should equal.")
			}
		}
	}

}

func TestChangePassword(t *testing.T) {
	require := require.New(t)

	for i := range changePass {
		var resp baseResponse
		body := postJSON(t, "/user/password", changePass[i], tokens[i])
		err := json.Unmarshal(body, &resp)
		require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
		require.Equal(http.StatusOK, resp.Code, "Get stream key should return OK")
		require.Equal("OK", resp.Message, "message should be OK")
	}
}

func TestLivingList(t *testing.T) {
	require := require.New(t)

	var resp liveResponse
	body := get(t, "/living/3", "")
	err := json.Unmarshal(body, &resp)
	require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
	require.Equal(0, resp.Total, "No user are living")
	require.Empty(resp.Users, "users should empty")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	client := store.NewRedisClient()
	defer client.Close()

	for i := 121; i < 124; i++ {
		client.SAdd(ctx, "living", strconv.Itoa(i))
		client.Set(ctx, "living:"+strconv.Itoa(i), time.Now().Format(time.RFC3339), 0)
	}
	body = get(t, "/living/4", "")
	err = json.Unmarshal(body, &resp)
	require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
	require.Equal(3, resp.Total, "No user are living")
	require.Len(resp.Users, 3, "3 users living")

	body = get(t, "/living/2", "")
	err = json.Unmarshal(body, &resp)
	require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
	require.Equal(2, resp.Total, "No user are living")
	require.Len(resp.Users, 2, "3 users living but should only return 2")

	client.SRem(ctx, "living", 122)
	client.Del(ctx, "living:122")

	body = get(t, "/living/3", "")
	err = json.Unmarshal(body, &resp)
	require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
	require.Equal(2, resp.Total, "No user are living")
	require.Len(resp.Users, 2, "2 users living")

}

func TestGetPublicUser(t *testing.T) {
	require := require.New(t)

	var resp pubResponse
	body := get(t, "/profile/121", "")
	err := json.Unmarshal(body, &resp)
	require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
	require.NotNil(resp.User, "User shouldn't nil")

	body = get(t, "/profile/122", "")
	err = json.Unmarshal(body, &resp)
	require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
	require.NotNil(resp.User, "User shouldn't nil")
}

func TestGetHistory(t *testing.T) {
	require := require.New(t)

	var resp historyResponse
	body := get(t, "/user/history", tokens[0])
	err := json.Unmarshal(body, &resp)
	require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
	require.Empty(resp.History, "History should empty")

	store.UpdateWatchHistory(31, "121")
	time.Sleep(time.Second)
	store.UpdateWatchHistory(31, "122")
	time.Sleep(time.Second)
	store.UpdateWatchHistory(31, "123")
	time.Sleep(time.Second)
	store.UpdateWatchHistory(31, "121")

	body = get(t, "/user/history", tokens[0])
	err = json.Unmarshal(body, &resp)
	require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
	require.Len(resp.History, 3, "history has 3 items")

}

func TestFollow(t *testing.T) {
	require := require.New(t)

	check := func (username string, followersNumber int, followingsNumber int) {
		var resp followResponse
		body := get(t, "/followers/"+username, "")
		err := json.Unmarshal(body, &resp)
		t.Log(string(body))
		require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
		require.Lenf(resp.Followers, followersNumber, "%v has %v followers.", username, followersNumber)

		resp = followResponse{}
		body = get(t, "/followings/"+username, "")
		err = json.Unmarshal(body, &resp)
		t.Log(string(body))
		require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
		require.Lenf(resp.Followings, followingsNumber, "%v has %v followings.", username, followingsNumber)
	}

	check("121", 0, 0)
	check("122", 0, 0)

	var resp baseResponse
	body := postForm(t, "/user/follow/122", nil, tokens[0])
	err := json.Unmarshal(body, &resp)
	t.Log(string(body))
	require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
	require.Equal(200, resp.Code, "Should return 200")

	check("121", 0, 1)
	check("122", 1, 0)

	body = postForm(t, "/user/unfollow/122", nil, tokens[0])
	err = json.Unmarshal(body, &resp)
	t.Log(string(body))
	require.NoErrorf(err, "Json Unmarshal Error <%v>", string(body))
	require.Equal(200, resp.Code, "Should return 200")

	check("121", 0, 0)
	check("122", 0, 0)
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
			"phone":    user.Phone,
		}
	}
	if user, ok := u.(*models.RegisterModel); ok {
		return map[string]string{
			"username": user.Username,
			"password": user.Password,
			"email":    user.Email,
			"phone":    user.Phone,
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
			"phone":    []string{user.Phone},
		}
	}
	if user, ok := u.(*models.RegisterModel); ok {
		return url.Values{
			"username": []string{user.Username},
			"password": []string{user.Password},
			"email":    []string{user.Email},
			"phone":    []string{user.Phone},
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
		{Phone: "13668686868"},
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
		{Phone: "+8613668686868"},
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
		{Phone: "+8612468686868", Password: "fca26135ea43ad0ba9f40f9ca54bdec34843eba157132bb2b15952ad445f5921"},
		{Username: "125", Password: "fca261f40f9ca54bdec34843eba1571324c9a136d2660bb2b15952ad445f5921"},
		{Email: "125@minitube.com", Password: "fca261f40f9ca54bdec34843eba1571324c9a136d2660bb2b15952ad445f5921"},
		{Phone: "+8612568686868", Password: "fca261f40f9ca54bdec34843eba1571324c9a136d2660bb2b15952ad445f5921"},
	}
	validRegister = []*models.RegisterModel{
		{Username: "121", Password: "c96d84ba3b4a823c4fee088a7369a5c02f50ef40f9ca54bdec34843eba157132"},
		{Username: "122", Password: "fca26135ea43ad0ba904e62c85793768e4c9a136d2660bb2b15952ad445f5921"},
		{Username: "123", Email: "123@minitube.com", Password: "fca26135ea43ad0ba904e62c85793768e4c9af40f9ca54bdec34843eba157132"},
		{Username: "124", Phone: "+8612468686868", Password: "fca26135ea43ad0ba9f40f9ca54bdec34843eba157132bb2b15952ad445f5921"},
		{Username: "125", Email: "125@minitube.com", Phone: "+8612568686868", Password: "fca261f40f9ca54bdec34843eba1571324c9a136d2660bb2b15952ad445f5921"},
	}
	changeProfile = []map[string]string{
		{"email": "121@minitube.com"},
		{"phone": "+1123456789"},
		{"live_name": "123's living room"},
		{"email": "124@minitube.com", "phone": ""},
		{"phone": "", "email": "", "live_name": "125's living room", "live_intro": "welcome to 125's room"},
	}
	changePass = []map[string]string{
		{"old_password": validRegister[0].Password, "new_password": "fca26135ea43ad0ba904e62c85793768e4c9a136d2660bb2b15952ad445f5921"},
		{"old_password": validRegister[1].Password, "new_password": "c96d84ba3b4a823c4fee088a7369a5c02f50ef40f9ca54bdec34843eba157132"},
	}
	tokens = make([]string, len(validLoginUser))
}

func TestMain(m *testing.M) {
	createUserForTest()
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
