package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"log"
	"net/http"
	"strings"

	app "github.com/AlexKomzzz/collectivity"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type dataClient struct {
	Debt        string `json:"debt"`
	AccessToken string `json:"token"`
}

// выдача формы для авторизации по почте и паролю с передачей ссылки редиректа
func (h *Handler) loginBot(c *gin.Context) {

	// парсинг URL, вытаскиевае ссылку редиректа
	redirectURL := c.Query("redirect_url")
	if redirectURL == "" {
		errorServerResponse(c, errors.New("invalid request"))
		return
	}

	// выдача формы с передачей ссылки
	c.HTML(http.StatusOK, "login_bot.html", gin.H{
		"URL": redirectURL,
	})
}

// аутентификация пользователя, выдача JWT
func (h *Handler) signInBot(c *gin.Context) { // Обработчик для аутентификации и получения токена

	var dataUser app.User

	// парсинг URL, вытаскиевае ссылку редиректа
	redirectURL := c.Query("redirect_url")
	if redirectURL == "" {
		errorServerResponse(c, errors.New("invalid request"))
		return
	}

	// ContentType = text/plain
	// выделим тело запроса
	/* структура тела запроса {
		email=<your_email>
		password=<your_pass>
		btn_login=
	}*/
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logrus.Println("ошибка при выделении тела запроса в signInBot: ", err)
		c.HTML(http.StatusBadRequest, "login_bot.html", gin.H{
			"err":    true,
			"errMsg": "Ошибка в запросе, пожалуйста, повторите.",
			"URL":    redirectURL,
		})
		return
	}

	// выделим email и password из body
	// разделим поля данных в запросе
	res := bytes.Split(body, []byte{13, 10})
	for _, params := range res {
		// делим строки по знаку равенства
		paramsSl := strings.Split(string(params), "=")

		if paramsSl[0] == "email" {
			dataUser.Email = strings.TrimSpace(paramsSl[1])
			// log.Println(paramsSl[1])
		} else if paramsSl[0] == "password" {
			dataUser.Password = strings.TrimSpace(paramsSl[1])
			// log.Println(paramsSl[1])
		}
	}

	token, err := h.service.GenerateJWT(dataUser.Email, dataUser.Password)
	if err != nil {
		if err.Error() == "нет пользователя" {
			logrus.Println("Handler/signInBot()/GenerateJWT(): ", err)
			c.HTML(http.StatusBadRequest, "login_bot.html", gin.H{
				"err":    true,
				"errMsg": "Пользователя с такой эл. почтой не существует. Проверьте правильность введенных данных или зарегистрируйтесь.",
				"URL":    redirectURL,
			})
		} else if err.Error() == "пароль" {
			logrus.Println("Handler/signInBot()/GenerateJWT(): ", err)
			c.HTML(http.StatusBadRequest, "login_bot.html", gin.H{
				"err":    true,
				"errMsg": "Неверный пароль. Попробуйте снова.",
				"URL":    redirectURL,
			})
		} else {
			logrus.Println("Handler/signInBot()/GenerateJWT(): ", err)
			c.HTML(http.StatusInternalServerError, "login_bot.html", gin.H{
				"err":    true,
				"errMsg": "Непредвиденная ошибка, пожалуйста, повторите.",
				"URL":    redirectURL,
			})
			return
		}

		return
	}

	idUser, err := h.service.ParseToken(token)
	if err != nil {
		logrus.Println("Handler/signInBot()/ParseToken(): ", err)
		c.HTML(http.StatusInternalServerError, "login_bot.html", gin.H{
			"err":    true,
			"errMsg": "Непредвиденная ошибка, пожалуйста, повторите.",
			"URL":    redirectURL,
		})
	}

	debt, err := h.service.GetDebtUser(idUser)
	if err != nil {
		logrus.Println("Handler/signInBot()/GetDebtUser(): ", err)
		c.HTML(http.StatusInternalServerError, "login_bot.html", gin.H{
			"err":    true,
			"errMsg": "Непредвиденная ошибка, пожалуйста, повторите.",
			"URL":    redirectURL,
		})
	}
	// c.JSON(http.StatusOK, gin.H{
	// 	"token": token,
	// })

	// редирект на стартовую страницу
	// c.Redirect(http.StatusTemporaryRedirect, startList+token)

	dataClient := &dataClient{
		Debt:        debt,
		AccessToken: token,
	}

	dataReq, err := json.Marshal(dataClient)
	if err != nil {
		logrus.Println("Handler/signInBot()/Marshal()/ ошибка при маршалинге структуру данных о клиенте: ", err)
		c.HTML(http.StatusInternalServerError, "login_bot.html", gin.H{
			"err":    true,
			"errMsg": "Непредвиденная ошибка, пожалуйста, повторите.",
			"URL":    redirectURL,
		})
	}

	bodyReq := strings.NewReader(string(dataReq))

	// отправка ПОСТ запроса на ссылку редиректа с debt в теле
	req, err := http.NewRequest("POST", redirectURL, bodyReq)
	if err != nil {
		logrus.Println("Handler/signInBot()/NewRequest()/ ошибка при создании запроса на редирект ссылку: ", err)
		c.HTML(http.StatusInternalServerError, "login_bot.html", gin.H{
			"err":    true,
			"errMsg": "Непредвиденная ошибка, пожалуйста, повторите.",
			"URL":    redirectURL,
		})
		return
	}
	// logrus.Printf("request Head = %s\n", req.Header.Get("Authorization"))
	// logrus.Printf("request = %v\n", req)
	// отправление GET запроса
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		logrus.Println("Handler/signInBot()/Do()/ ошибка при отправке запроса на редирект ссылку: ", err)
		c.HTML(http.StatusInternalServerError, "login_bot.html", gin.H{
			"err":    true,
			"errMsg": "Непредвиденная ошибка, пожалуйста, повторите.",
			"URL":    redirectURL,
		})
		return
	}

	defer response.Body.Close()

	// Проверка кода ответа
	if response.StatusCode != http.StatusOK {
		logrus.Println("Handler/signInBot()/StatusCode/ получен статус код: ", response.StatusCode, " (ожидание: 200)")
		c.HTML(http.StatusBadRequest, "login_bot.html", gin.H{
			"err":    true,
			"errMsg": "Непредвиденная ошибка, пожалуйста, повторите.",
			"URL":    redirectURL,
		})
		return
	}

	// перенаправление на тлг бота
	botURL := viper.GetString("bot_url")
	c.Header("Location", botURL)
	c.Writer.WriteHeader(http.StatusMovedPermanently)
	// c.AbortWithStatus(http.StatusMovedPermanently)
	log.Println("ссылка на бота отправлена: ", botURL)
}
