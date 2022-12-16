package handler

import (
	"encoding/json"
	"errors"

	"log"
	"net/http"
	"strings"

	app "github.com/AlexKomzzz/collectivity"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	keyTokenCtx    = "token"
	keyRedirectURL = "redirectURL"
)

type dataClient struct {
	AccessToken string `json:"token"`
}

// проврка куков на наличие токена, при отсутствии токена выдача формы для авторизации по почте и паролю с передачей ссылки редиректа
func (h *Handler) loginBot(c *gin.Context) {

	// парсинг URL, вытаскиваем ссылку редиректа
	redirectURL := c.Query("redirect_url")
	if redirectURL == "" {
		// при остуствии ссылки редиректа отправляем ошибку в АПИ бота
		logrus.Println("отсутствует ссылка редиректа в URL регистрации для тлг бота")
		c.HTML(http.StatusBadRequest, "login_bot.html", gin.H{
			"err":    true,
			"errMsg": "Непредвиденная ошибка, пожалуйста, создайте заного ссылку в телеграмм боте",
		})
		return
	}

	// проверка JWT в куках
	session := sessions.Default(c)
	sessionToken := session.Get("token")

	// если в куках отсутствует токен, ссылку редиректа передаем в куках, перенаправляем на форму авторизации
	if sessionToken == nil {
		logrus.Println("Вход без идентификации")

		// создаем куки
		c.SetCookie("redirectTLG", redirectURL, 60*60*24, "/", viper.GetString("host"), true, true)

		logrus.Println("Запись куки-файла регистации для тлг")
		// при отсутствии токена в куках выдача формы на авторизацию
		c.HTML(http.StatusBadRequest, "login_bot.html", gin.H{})
		return
	}

	// если в куках уже есть токен, проверяем его на валидность и отдаем обратно
	_, err := h.service.ParseToken(sessionToken.(string))

	// если токен не валиден или протух выдаем форму авторизации
	if err != nil {
		logrus.Println("не валидный или протухший токен в куках от бота при авторизации: ", err)
		// создаем куки
		c.SetCookie("redirectTLG", redirectURL, 60*60*24, "/", viper.GetString("host"), true, true)
		logrus.Println("Запись куки-файла регистации для тлг")

		c.HTML(http.StatusBadRequest, "login_bot.html", gin.H{})
		return
	}

	// если токен валиден, отправляем его в апи бота
	c.Set(keyTokenCtx, sessionToken.(string))
	c.Set(keyRedirectURL, redirectURL)
	// отправка токена в АПИ бота
	h.sendTokenToBot(c)
}

// получение данных аутентификации пользователя, генерация JWT
func (h *Handler) signInBot(c *gin.Context) {

	var dataUser app.User

	// получение ссылки редиректа из куки
	redirectURL, err := c.Cookie("redirectTLG")
	if err != nil {
		logrus.Println("Handler/signInBot()/GetDebtUser(): ", err)
		errorServBot(c)
		return
	}

	// получение данных от пользователя из пост запроса
	dataUser.Email = c.PostForm("email")
	dataUser.Password = c.PostForm("password")

	if dataUser.Email == "" || dataUser.Password == "" {
		logrus.Println("не получены email или пароль при аутентификации для тлг бота")
		c.HTML(http.StatusBadRequest, "login_bot.html", gin.H{
			"err":    true,
			"errMsg": "Ошибка в запросе, пожалуйста, повторите.",
		})
		return
	}

	// генерация JWT
	token, err := h.service.GenerateJWT(dataUser.Email, dataUser.Password)
	if err != nil {
		if err.Error() == "нет пользователя" {
			logrus.Println("Handler/signInBot()/GenerateJWT(): ", err)
			c.HTML(http.StatusBadRequest, "login_bot.html", gin.H{
				"err":    true,
				"errMsg": "Пользователя с указанным адресом эл. почты не существует. Проверьте правильность введенных данных или зарегистрируйтесь",
			})
		} else if err.Error() == "пароль" {
			logrus.Println("Handler/signInBot()/GenerateJWT(): ", err)
			c.HTML(http.StatusBadRequest, "login_bot.html", gin.H{
				"err":    true,
				"errMsg": "Неверный пароль. Попробуйте снова",
			})
		} else {
			logrus.Println("Handler/signInBot()/GenerateJWT(): ", err)
			errorServBot(c)
			return
		}

		return
	}

	c.Set(keyTokenCtx, token)
	c.Set(keyRedirectURL, redirectURL)
	// отправка токена в АПИ бота
	h.sendTokenToBot(c)
}

// передача токена АПИ боту
func (h *Handler) sendTokenToBot(c *gin.Context) {

	// получение токена из контекста
	token, err := getTokenCtx(c)
	if err != nil {
		logrus.Println("Handler/getDebtBot()/getTokenCtx(): ", err)
		errorServBot(c)
		return
	}

	// получение редирект ссылки на API tlg bot из контекста
	redirectURL, err := getRedirectURLCtx(c)
	if err != nil {
		logrus.Println("Handler/getDebtBot()/getRedirectURLCtx(): ", err)
		errorServBot(c)
		return
	}

	// структура для отправки на API tlg_bot
	dataClient := &dataClient{
		AccessToken: token,
	}

	dataReq, err := json.Marshal(dataClient)
	if err != nil {
		logrus.Println("Handler/getDebtBot()/Marshal()/ ошибка при маршалинге структуру данных о клиенте: ", err)
		errorServBot(c)
		return
	}

	bodyReq := strings.NewReader(string(dataReq))

	// отправление POST запроса
	response, err := http.Post(redirectURL, "application/json", bodyReq)
	if err != nil {
		logrus.Println("Handler/getDebtBot()/Post()/ ошибка при отправке запроса на редирект ссылку: ", err)
		errorServBot(c)
		return
	}
	defer response.Body.Close()

	// Проверка кода ответа
	if response.StatusCode != http.StatusOK {
		logrus.Println("Handler/signInBot()/StatusCode/ получен статус код: ", response.StatusCode, " (ожидание: 200)")
		errorServBot(c)
		return
	}

	log.Println("успешная регистрация тлг бота для пользователя")

	// перенаправление на тлг бота
	botURL := viper.GetString("bot_url")
	c.Header("Location", botURL)
	c.Writer.WriteHeader(http.StatusMovedPermanently)
}

// передача данных о задолженности
func (h *Handler) getDebtBot(c *gin.Context) {

	// проверка JWT в куках
	session := sessions.Default(c)
	sessionToken := session.Get("token")

	// если в куках отсутствует токен - ответ bad request
	if sessionToken == nil {
		logrus.Println("Запрос долга от бота без идентификации пользователя")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "token",
		})
		return
	}

	// определение idUser по JWT
	idUser, err := h.service.ParseToken(sessionToken.(string))
	if err != nil {
		logrus.Println("Протухший или невалидный токен, при запросе долга от бота: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "token",
		})
		return
	}

	// получение данных о задолженности по idUser
	debt, err := h.service.GetDebtUser(idUser)
	if err != nil {
		logrus.Println("Handler/getDebtBot()/GetDebtUser(): ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "debt",
		})
		return
	}

	// отправка данных в апи бота
	c.JSON(http.StatusOK, gin.H{
		"debt": debt,
	})
}

// вытащить токен из контекста
func getTokenCtx(c *gin.Context) (string, error) {

	tokenCtx, ok := c.Get(keyTokenCtx)
	if !ok {
		return "", errors.New("token not found by contect")
	}

	token, ok := tokenCtx.(string)
	if !ok {
		return "", errors.New("token is of invalid type by contect")
	}

	return token, nil
}

// вытащить редирект ссылку из контекста
func getRedirectURLCtx(c *gin.Context) (string, error) {

	redirectURLCtx, ok := c.Get(keyRedirectURL)
	if !ok {
		return "", errors.New("redirectURL not found by contect")
	}

	redirectURL, ok := redirectURLCtx.(string)
	if !ok {
		return "", errors.New("redirectURL is of invalid type by contect")
	}

	return redirectURL, nil
}

func errorServBot(c *gin.Context) {
	c.HTML(http.StatusInternalServerError, "login_bot.html", gin.H{
		"err":    true,
		"errMsg": "Непредвиденная ошибка, пожалуйста, повторите.",
	})
}
