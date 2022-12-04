package handler

import (
	"encoding/json"
	"errors"

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

	dataUser.Email = c.PostForm("email")
	dataUser.Password = c.PostForm("password")

	if dataUser.Email == "" || dataUser.Password == "" {
		logrus.Println("не получены email или пароль при аутентификации для тлг бота")
		c.HTML(http.StatusBadRequest, "login_bot.html", gin.H{
			"err":    true,
			"errMsg": "Ошибка в запросе, пожалуйста, повторите.",
			"URL":    redirectURL,
		})
		return
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

	// отправление POST запроса
	response, err := http.Post(redirectURL, "application/json", bodyReq)
	// response, err := http.DefaultClient.Do(req)
	if err != nil {
		logrus.Println("Handler/signInBot()/Post()/ ошибка при отправке запроса на редирект ссылку: ", err)
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
