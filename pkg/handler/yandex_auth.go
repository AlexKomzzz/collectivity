package handler

// https://oauth.yandex.ru/

// инфа https://yandex.ru/dev/id/doc/dg/oauth/reference/auto-code-client.html#auto-code-client__get-code
// https://oauth.yandex.ru/client/fe3918e6e2ee4e68b1391c55e117ebc5

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	app "github.com/AlexKomzzz/collectivity"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

// URL для запроса к API Яндекс ID для получения данных о пользователе
const oauthYandexUrlAPI = "https://login.yandex.ru/info?"

// структура ответа Яндекс.OAuth с токеном и данными пользователя
type responceYandex struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"` // время жизни токена в сек
	RefreshToken string `json:"refresh_token"`
	//Scope        []string `json:"scope"`
}

// конфигурация клиента
var yandexOauthConfig = &oauth2.Config{
	RedirectURL: "http://localhost:8080/auth/yandex/callback",
	// ClientID:     viper.GetString("YANDEX_OAUTH_CLIENT_ID"), так не записывает!!!!
	//ClientSecret: os.Getenv("YANDEX_OAUTH_CLIENT_SECRET"),
	Scopes: []string{"login:default_phone", "login:info", "login:email"},
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://oauth.yandex.com/authorize",
		TokenURL: "https://oauth.yandex.com/token",
	},
}

// перенаправление клиента по ссылке на форму предоставление доступа
func (h *Handler) oauthYandexLogin(c *gin.Context) {
	// Create oauthState cookie
	oauthState := generateStateOauthCookie(c)

	yandexOauthConfig.ClientID = viper.GetString("YANDEX_OAUTH_CLIENT_ID")
	yandexOauthConfig.ClientSecret = os.Getenv("YANDEX_OAUTH_CLIENT_SECRET")

	// AuthCodeURL receive state that is a token to protect the user from CSRF attacks. You must always provide a non-empty string and
	// validate that it matches the the state query parameter on your redirect callback.
	// создание URL в соответствии с полем конфигурации Endpoint
	url := yandexOauthConfig.AuthCodeURL(oauthState)
	// logrus.Printf("url = %s\n", url)

	c.Redirect(http.StatusTemporaryRedirect, url)
}

// получение данных о пользователе от Яндекс.OAuth
// Когда пользователь разрешает доступ к своим данным, Яндекс.OAuth перенаправляет его в приложение.
// Код подтверждения включается в URL перенаправления.
// генерация JWT и выдача пользователю
func (h *Handler) oauthYandexCallback(c *gin.Context) {
	// Сравним cookie
	oauthState, err := c.Cookie("oauthstate")
	if err != nil {
		logrus.Println(err)
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"error": "Непредвиденная ошибка, пожалуйста, повторите.",
		})
		return
	}

	// сравним поле state  для защиты от CSRF-атак
	if c.Request.FormValue("state") != oauthState {
		logrus.Println("invalid oauth yandex state")
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"error": "Непредвиденная ошибка, пожалуйста, повторите.",
		})
		return
	}

	// token, err := getAccessTokenFromYandex(c)
	// if err != nil {
	// 	logrus.Println(err.Error())
	// 	c.Redirect(http.StatusTemporaryRedirect, "/")
	// 	return
	// }

	// получение данных пользователя от API Яндекс ID
	dataAPIyandex, err := getUserDataFromYandex(c)
	if err != nil {
		logrus.Println(err)
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{
			"error": "Непредвиденная ошибка, пожалуйста, повторите.",
		})
		return
	}

	var userDataAPI = &app.UserYandex{}

	json.Unmarshal(dataAPIyandex, userDataAPI)
	// logrus.Println("dataUserYandex: ", string(dataAPIyandex))
	logrus.Println("newUserYandex: ", userDataAPI)

	idUserAPI, err := h.service.ConvertID(userDataAPI.Id)
	if err != nil {
		logrus.Println("Handler/oauthYandexCallback()/ConvertID()/ ошибка при конвертации idUser из строки в число: ", err)
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{
			"error": "Непредвиденная ошибка, пожалуйста, повторите.",
		})
		return
	}

	// ГЕНЕРАЦИЯ JWT_API
	tokenAPI, err := h.service.GenerateJWTbyID(idUserAPI)
	if err != nil {
		logrus.Println("Handler/oauthYandexCallback()/GenerateJWT_API(): ", err)
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{
			"error": "Непредвиденная ошибка, пожалуйста, повторите.",
		})
		return
	}

	user := &app.User{
		FirstName: userDataAPI.FirstName,
		LastName:  userDataAPI.LastName,
		Email:     userDataAPI.Email,
	}

	// декодировка структуры пользователя в слайз байт для кэширования в redis
	dataCash, err := json.Marshal(user)
	if err != nil {
		logrus.Println("Handler/oauthYandexCallback()/Marshal()/ ошибка при декодировке данных пользователя: ", err)
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{
			"error": "Непредвиденная ошибка, пожалуйста, повторите.",
		})
		return
	}

	// запись данных о пользователе в кэш
	// ключ в кэше - ID
	err = h.service.SetUserCash(idUserAPI, dataCash)
	if err != nil {
		logrus.Println("Handler/oauthYandexCallback()/SetUserCash()/ ошибка при записи данных в кэш: ", err)
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{
			"error": "Непредвиденная ошибка, пожалуйста, повторите.",
		})
		return
	}

	// отправка формы для получения отчества пользователя
	c.HTML(http.StatusOK, "middle_names.html", gin.H{
		"token": tokenAPI,
	})

	/*
		// создание пользователя в БД (или обновление, если уже создан)
		idUser, err := h.service.CreateUserAPI("yandex", userData.Id, userData.FirstName, userData.LastName, userData.Email)
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
			logrus.Println(err.Error())
			c.HTML(http.StatusInternalServerError, "login.html", gin.H{
				"error": "Непредвиденная ошибка, пожалуйста, повторите.",
			})
			return
		}

		logrus.Printf("UserInfoYandex: %s\n", userData)
		logrus.Printf("JWT: %s\n", token)

		// передача JWT токена пользователю
		c.JSON(http.StatusOK, gin.H{
			"token": token,
		})
	*/
}

// запрос к API Google для получение данных о пользователе по access token`у
func getUserDataFromYandex(c *gin.Context) ([]byte, error) {

	// получение access token при помощи кода подтверждения
	// accessToken, err := getAccessTokenFromYandex(c)
	// if err != nil {
	// 	return nil, fmt.Errorf("not received access token: %s", err.Error())
	// }

	accessToken, err := getAccessTokenFromYandex(c)
	if err != nil {
		return nil, err
	}

	// создание GET запроса с заголовком токена доступа для получения данных о пользователе
	req, err := http.NewRequest("GET", oauthYandexUrlAPI, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("not create GET request: %s", err.Error())
	}
	req.Header.Set("Authorization", "OAuth "+accessToken)
	// logrus.Printf("request Head = %s\n", req.Header.Get("Authorization"))
	// logrus.Printf("request = %v\n", req)
	// отправление GET запроса
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("not create DO request: %s", err.Error())
	}

	defer response.Body.Close()

	// чтение тела ответа на запрос GET
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed read response by API Яндекс ID: %s", err.Error())
	}

	return contents, nil
}

// запрос к Яндекс.OAuth для получение access token при помощи кода подтверждения
func getAccessTokenFromYandex(c *gin.Context) (string, error) {

	// выделим code из URL ответа при согласии пользователя
	codeRequest := c.Request.FormValue("code")

	// создадим заголовок авторизации для запроса данных о пользователе
	/*
		Идентификатор и пароль приложения также можно отправить в заголовке Authorization, закодировав строку <client_id>:<client_secret> методом base64.
		Если Яндекс.OAuth получает заголовок Authorization, параметры client_id и client_secret в теле запроса игнорируются.
	*/

	// создадим тело запроса
	body := strings.NewReader(fmt.Sprintf("grant_type=authorization_code&code=%s&client_id=%s&client_secret=%s", codeRequest, yandexOauthConfig.ClientID, yandexOauthConfig.ClientSecret))

	// отправление Post запроса для получения access token
	// response, err := http.Post(yandexOauthConfig.Endpoint.TokenURL, "application/x-www-form-urlencoded", bytes.NewBuffer(body))
	// if err != nil {
	// 	return "", fmt.Errorf("failed request to Яндекс.OAuth: %s", err.Error())
	// }
	// defer response.Body.Close()
	req, err := http.NewRequest("POST", yandexOauthConfig.Endpoint.TokenURL, body)
	if err != nil {
		return "", fmt.Errorf("not create POST request: %s", err.Error())
	}

	// в запрос добавим заголовок следующего вида:
	// [Authorization: Basic <закодированная строка client_id:client_secret>]
	clId := url.QueryEscape(yandexOauthConfig.ClientID)
	clSecret := url.QueryEscape(yandexOauthConfig.ClientSecret)
	req.SetBasicAuth(clId, clSecret)
	// добавление заголовка
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")
	// logrus.Printf("request Head = %s\n", req.Header.Get("Authorization"))
	// logrus.Printf("request = %v\n", req)

	// отправление POST запроса
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("not create DO request: %s", err.Error())
	}

	defer response.Body.Close()

	// чтение тела ответа на запрос Post
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed read response by Яндекс.OAuth: %s", err.Error())
	}

	var responceAccessToken = &responceYandex{}
	json.Unmarshal(contents, responceAccessToken)

	// logrus.Println(responceAccessToken)
	// написать продление токена, если нужно

	return responceAccessToken.AccessToken, nil
}

// запрос для обновления access token
/*func updateAccessTokenFromGoogle(c *gin.Context) (string, error) {
	var newToken string
	/*
	   Обновить token (https://yandex.ru/dev/id/doc/dg/oauth/reference/refresh-client.html#refresh-client__get-token)
	   POST /token HTTP/1.1
	   Host: oauth.yandex.ru
	   Content-type: application/x-www-form-urlencoded
	   Content-Length: <длина тела запроса>
	   [Authorization: Basic <закодированная строка client_id:client_secret>]

	   grant_type=refresh_token
	    & refresh_token=<refresh_token>
	   [& client_id=fe3918e6e2ee4e68b1391c55e117ebc5]
	   [& client_secret=b06823335f3f436d9882b0c9cd0f8a34]

	return newToken, nil
}*/
