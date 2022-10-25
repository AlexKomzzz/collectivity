package handler

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"

	app "github.com/AlexKomzzz/collectivity"
	"github.com/gin-gonic/gin"
)

// Создание нового пользователя, добавление в БД и выдача токена авторизации
func (h *Handler) signUp(c *gin.Context) { // Обработчик для регистрации

	var dataUser app.User

	// ContentType = text/plain
	// выделим тело запроса
	/* структура тела запроса {
		first-name=<first-name>
		last-name=<last-name>
		middle-name=<middle-name> OR nil !!!!
		email=<your_email>
		psw=<password>
		btn_login=
	}*/
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// выделим данные из body и запишем в структуру User
	// разделим поля данных в запросе
	res := bytes.Split(body, []byte{13, 10})
	for _, params := range res {
		// делим строки по знаку равенства
		paramsSl := strings.Split(string(params), "=")

		if paramsSl[0] == "first-name" {
			dataUser.FirstName = paramsSl[1]
			// log.Println(paramsSl[1])
		} else if paramsSl[0] == "last-name" {
			dataUser.LastName = paramsSl[1]
			// log.Println(paramsSl[1])
		} else if paramsSl[0] == "middle-name" {
			if len(paramsSl) > 1 {
				dataUser.MiddleName = paramsSl[1]
			}
			// log.Println(paramsSl[1])
		} else if paramsSl[0] == "email" {
			dataUser.Email = paramsSl[1]
			// log.Println(paramsSl[1])
		} else if paramsSl[0] == "psw" {
			dataUser.Password = paramsSl[1]
			// log.Println(paramsSl[1])
		}
	}

	id, err := h.service.CreateUser(&dataUser)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	token, err := h.service.GenerateJWT(dataUser.Email, dataUser.Password)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":    id,
		"token": token,
	})
}

// type signInInput struct { // Структура для идентификации
// 	Email    string `json:"email" binding:"required"`
// 	Password string `json:"password" binding:"required"`
// }

// авторизация пользователя, выдача JWT
func (h *Handler) signIn(c *gin.Context) { // Обработчик для аутентификации и получения токена

	var dataUser app.User

	// ContentType = text/plain
	// выделим тело запроса
	/* структура тела запроса {
		email=<your_email>
		password=<your_pass>
		btn_login=
	}*/
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// выделим email и password из body
	// разделим поля данных в запросе
	res := bytes.Split(body, []byte{13, 10})
	for _, params := range res {
		// делим строки по знаку равенства
		paramsSl := strings.Split(string(params), "=")

		if paramsSl[0] == "email" {
			dataUser.Email = paramsSl[1]
			// log.Println(paramsSl[1])
		} else if paramsSl[0] == "password" {
			dataUser.Password = paramsSl[1]
			// log.Println(paramsSl[1])
		}
	}

	token, err := h.service.GenerateJWT(dataUser.Email, dataUser.Password)
	if err != nil {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"error": "Такого пользователя не существует",
		})
		//newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}
