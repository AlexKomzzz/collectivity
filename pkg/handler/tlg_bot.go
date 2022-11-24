package handler

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	app "github.com/AlexKomzzz/collectivity"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

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
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		logrus.Println("ошибка при выделении тела запроса в signIn: ", err)
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
			c.HTML(http.StatusBadRequest, "login_bot.html", gin.H{
				"err":    true,
				"errMsg": "Пользователя с такой эл. почтой не существует. Проверьте правильность введенных данных или зарегистрируйтесь.",
				"URL":    redirectURL,
			})
		} else if err.Error() == "пароль" {
			c.HTML(http.StatusBadRequest, "login.html", gin.H{
				"err":    true,
				"errMsg": "Неверный пароль. Попробуйте снова.",
				"URL":    redirectURL,
			})
		} else {
			logrus.Println("Handler/signIn(): ", err)
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
		logrus.Println("parseToken: ", err)
		c.HTML(http.StatusInternalServerError, "login_bot.html", gin.H{
			"err":    true,
			"errMsg": "Непредвиденная ошибка, пожалуйста, повторите.",
			"URL":    redirectURL,
		})
	}

	debt, err := h.service.GetDebtUser(idUser)
	if err != nil {
		logrus.Println("parseToken: ", err)
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

	dataReq := strings.NewReader(fmt.Sprintf("\"debt\":\"%s\",\n\"token\":\"%s\"", debt, token))

	// отправка ПОСТ запроса на ссылку редиректа с debt в теле
	req, err := http.NewRequest("POST", redirectURL, dataReq)
	if err != nil {
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
		c.HTML(http.StatusBadRequest, "login_bot.html", gin.H{
			"err":    true,
			"errMsg": "Непредвиденная ошибка, пожалуйста, повторите.",
			"URL":    redirectURL,
		})
		return
	}

}
