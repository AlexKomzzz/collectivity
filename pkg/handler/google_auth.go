package handler

// https://console.cloud.google.com/apis/credentials?project=my-project-id-365820

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	app "github.com/AlexKomzzz/collectivity"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// URL для запроса к API Google для получения данных о пользователе
const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

// const oauthGoogleUrlAPI = "https://www.googleapis.com/drive/v2/files?access_token="

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

	c.Redirect(http.StatusTemporaryRedirect, url)
}

// получение данных о пользователе от API Google
func (h *Handler) oauthGoogleCallback(c *gin.Context) {
	// Сравним cookie
	oauthState, err := c.Cookie("oauthstate")
	if err != nil {
		logrus.Println("ошибка при получении куков в oauthGoogleCallback", err)
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"error": "Непредвиденная ошибка, пожалуйста, повторите.",
		})
		// newErrorResponse(c, http.StatusBadRequest, "Cookie invalid: "+err.Error())
		return
	}

	if c.Request.FormValue("state") != oauthState {
		logrus.Println("invalid oauth google state")
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"error": "Непредвиденная ошибка, пожалуйста, повторите.",
		})
		return
	}

	// token, err := getAccessTokenFromGoogle(c)
	// if err != nil {
	// 	logrus.Println(err.Error())
	// 	c.Redirect(http.StatusTemporaryRedirect, "/")
	// 	return
	// }

	data, err := getUserDataFromGoogle(c)
	if err != nil {
		logrus.Println(err)
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{
			"error": "Непредвиденная ошибка, пожалуйста, повторите.",
		})
		return
	}

	var userData = &app.UserGoogle{}

	json.Unmarshal(data, userData)

	// создание пользователя в БД (или обновление, если уже создан)
	idUser, err := h.service.CreateUserAPI("google", userData.Id, userData.FirstName, userData.LastName, userData.Email)
	if err != nil {
		logrus.Println(err)
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{
			"error": "Непредвиденная ошибка, пожалуйста, повторите.",
		})
		return
	}

	// получение JWT по id пользователя
	token, err := h.service.GenerateJWT_API(idUser)
	if err != nil {
		logrus.Println(err)
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{
			"error": "Непредвиденная ошибка, пожалуйста, повторите.",
		})
		return
	}

	logrus.Printf("UserInfoGoogle: %s\n", userData)
	logrus.Printf("JWT: %s\n", token)

	// передача JWT токена пользователю
	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

// создание Cookie
func generateStateOauthCookie(c *gin.Context) string {

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	c.SetCookie("oauthstate", state, 60*60*24, "/", viper.GetString("domain"), true, true)

	return state
}

// запрос для полученя access token
func getAccessTokenFromGoogle(c *gin.Context) (string, error) {

	// Use code to get token and get user info from Google.
	// Exchange преобразует 'code' авторизации в токен.
	token, err := googleOauthConfig.Exchange(c, c.Request.FormValue("code"))
	if err != nil {
		logrus.Println("ошибка получения токена")
		return "", fmt.Errorf("code exchange wrong: %s", err.Error())
	}
	return token.AccessToken, nil
}

// запрос к API Google для получение данных о пользователе по access token`у
func getUserDataFromGoogle(c *gin.Context) ([]byte, error) {

	// получение токена
	accessToken, err := getAccessTokenFromGoogle(c)
	if err != nil {
		return nil, err
	}

	// создание GET запроса с заголовком токена доступа для получения данных о пользователе
	req, err := http.NewRequest("GET", oauthGoogleUrlAPI, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("not create GET request: %s", err.Error())
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	// logrus.Printf("request Head = %s\n", req.Header.Get("Authorization"))
	// logrus.Printf("request = %v\n", req)

	// отправление GET запроса
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("not create DO request: %s", err.Error())
	}

	// отправление GET запроса
	// response, err := http.Get(oauthGoogleUrlAPI + accessToken)
	// // лучше использовать curl -H "Authorization: Bearer access_token" https://www.googleapis.com/drive/v2/files
	// if err != nil {
	// 	return nil, fmt.Errorf("failed request to API Google: %s", err.Error())
	// }
	defer response.Body.Close()

	// чтение тела ответа на запрос GET
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed read response by API Google: %s", err.Error())
	}
	return contents, nil
}
