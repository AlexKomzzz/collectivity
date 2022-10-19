package handler

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// URL для запроса к API Google для получения данных о пользователе
const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

//const oauthGoogleUrlAPI = "https://www.googleapis.com/drive/v2/files?access_token="

// конфигурация клиента
var googleOauthConfig = &oauth2.Config{
	RedirectURL: "http://localhost:8080/auth/google/callback",
	// ClientID:     viper.GetString("client_ID"), так не записывает!!!!
	//ClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
	Scopes:   []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	Endpoint: google.Endpoint,
}

// перенаправление клиента по ссылке на форму предоставление доступа
func (h *Handler) oauthGoogleLogin(c *gin.Context) {
	// Create oauthState cookie
	oauthState := generateStateOauthCookie(c)

	googleOauthConfig.ClientID = viper.GetString("GOOGLE_OAUTH_CLIENT_ID")
	googleOauthConfig.ClientSecret = os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET")

	// AuthCodeURL receive state that is a token to protect the user from CSRF attacks. You must always provide a non-empty string and
	// validate that it matches the the state query parameter on your redirect callback.
	// создание URL в соответствии с полем конфигурации Endpoint
	url := googleOauthConfig.AuthCodeURL(oauthState)
	logrus.Printf("url = %s\n", url)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// получение данных о пользователе от API Google
func (h *Handler) oauthGoogleCallback(c *gin.Context) {
	// Сравним cookie
	oauthState, err := c.Cookie("oauthstate")
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "Cookie invalid: "+err.Error())
		return
	}

	if c.Request.FormValue("state") != oauthState {
		logrus.Println("invalid oauth google state")
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	data, err := getUserDataFromGoogle(c)
	if err != nil {
		logrus.Println(err.Error())
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	// GetOrCreate User in your db.
	// Redirect or response with a token.
	// More code .....
	logrus.Printf("UserInfo: %s\n", data)
	c.JSON(http.StatusOK, gin.H{
		"user_data": string(data),
	})
}

// создание Cookie
func generateStateOauthCookie(c *gin.Context) string {

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	c.SetCookie("oauthstate", state, 60*60*24, "/", "localhost", true, true)

	return state
}

// запрос к API Google для получение данных о пользователе по access token`у
func getUserDataFromGoogle(c *gin.Context) ([]byte, error) {

	// Use code to get token and get user info from Google.
	// Exchange преобразует 'code' авторизации в токен.
	token, err := googleOauthConfig.Exchange(c, c.Request.FormValue("code"))
	if err != nil {
		return nil, fmt.Errorf("code exchange wrong: %s", err.Error())
	}

	// отправление GET запроса
	response, err := http.Get(oauthGoogleUrlAPI + token.AccessToken)
	// лучше использовать curl -H "Authorization: Bearer access_token" https://www.googleapis.com/drive/v2/files
	if err != nil {
		return nil, fmt.Errorf("failed request to API Google: %s", err.Error())
	}
	defer response.Body.Close()

	// чтение тела ответа на запрос GET
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed read response by API Google: %s", err.Error())
	}
	return contents, nil
}
