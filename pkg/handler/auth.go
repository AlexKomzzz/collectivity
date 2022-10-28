package handler

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	app "github.com/AlexKomzzz/collectivity"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// стартовая страница
func (h *Handler) startList(c *gin.Context) {

	// идентифицируем пользователя
	idUser, err := h.userIdentity(c)
	if err != nil {
		logrus.Println("Вход без идентификации")
		c.Redirect(http.StatusTemporaryRedirect, "/auth/login")
		return
	}

	// проверка роли пользователя
	roleUser, err := h.service.GetRole(idUser)
	if err != nil {
		logrus.Println(err)
		if err.Error() == "sql: no rows in result set" {
			// значит это обычный пользователь
			c.HTML(http.StatusOK, "start_list.html", gin.H{})
			return
		} else {
			newErrorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}
	}

	logrus.Println("roleUser: ", roleUser)
	if roleUser == "admin" {
		// значит это админ
		c.HTML(http.StatusOK, "start_list.html", gin.H{
			"role": roleUser,
		})
		return
	} else {
		logrus.Println("ошибка в определении роли пользователя из БД")
		newErrorResponse(c, http.StatusInternalServerError, "ошибка в определении роли пользователя из БД")
		return
	}
}

// Создание нового пользователя, добавление в БД и выдача токена авторизации
func (h *Handler) signUp(c *gin.Context) { // Обработчик для регистрации

	var dataUser app.User
	var passRepeat string

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
		} else if paramsSl[0] == "psw-repeat" {
			passRepeat = paramsSl[1]
			// log.Println(paramsSl[1])
		}
	}

	//logrus.Printf("dataUser: %v", dataUser)

	idUser, err := h.service.CreateUser(&dataUser, passRepeat)
	if err != nil {
		if err.Error() == "пароли не совпадают" {
			c.HTML(http.StatusBadRequest, "forma_auth.html", gin.H{
				"error": err,
			})
			// c.Redirect(http.StatusTemporaryRedirect, "/auth/sign-form?error=пароли%20не%20совпадают")
			// c.Redirect(http.StatusTemporaryRedirect, "/test")

		} else {
			newErrorResponse(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	token, err := h.service.GenerateJWT_API(idUser)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":    idUser,
		"token": token,
	})
}

/*
// отправление формы для создания нового пользователя
func (h *Handler) formAuth(c *gin.Context) {
	// вытаскиваем с URL ошибку
	errorString := c.Query("error")

	// передаем форму для создание акк
	c.HTML(http.StatusBadRequest, "forma_auth.html", gin.H{
		"error": errorString,
	})

}*/

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
		if err.Error() == "sql: no rows in result set" {
			c.HTML(http.StatusBadRequest, "login.html", gin.H{
				"error": "Такого пользователя не существует.\nПроверьте правильность введенных данных или зарегистрируйтесь",
			})
		} else {
			c.HTML(http.StatusBadRequest, "login.html", gin.H{
				"error": err,
			})
		}
		//newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

// определение пользователя по email
func (h *Handler) definitionUser(c *gin.Context) {

	var emailUser string

	// ContentType = text/plain
	// выделим тело запроса
	/* структура тела запроса {
		email=<your_email>
		btn_login=
	}*/
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		logrus.Println(err)
		c.HTML(http.StatusBadRequest, "recovery_pass.html", gin.H{
			"error": err,
		})
		return
	}

	// выделим email из body
	// разделим поля данных в запросе
	res := bytes.Split(body, []byte{13, 10})
	for _, params := range res {
		// делим строки по знаку равенства
		paramsSl := strings.Split(string(params), "=")

		if paramsSl[0] == "email" {
			emailUser = paramsSl[1]
			// log.Println(paramsSl[1])
		}
	}

	// идентифицируем пользователя по email
	token, err := h.service.DefinitionUserByEmail(emailUser)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			logrus.Println(err)
			c.HTML(http.StatusBadRequest, "recovery_pass.html", gin.H{
				"error": "Пользователя с такой почтой не существует.",
			})
			return
		} else {
			logrus.Println(err)
			c.HTML(http.StatusBadRequest, "recovery_pass.html", gin.H{
				"error": err,
			})
			return
		}
	}

	// формирование URL
	URLs := fmt.Sprintf("%s/auth/pass/definition-userJWT?token=%s", viper.GetString("url"), url.PathEscape(token))

	// отпрвка сообщения на почту пользователя с ссылкой для восстановления пароля
	err = h.service.SendMessage(emailUser, URLs)
	if err != nil {
		logrus.Println(err)
		c.HTML(http.StatusBadRequest, "recovery_pass.html", gin.H{
			"error": err,
		})
		return
	}

}

// // определение пользователя по JWT
func (h *Handler) definitionUserJWT(c *gin.Context) {

	// определение JWT из URL
	token := c.Query("token")
	if token == "" {
		logrus.Println("empty auth token")
		newErrorResponse(c, http.StatusBadRequest, "empty auth token")
		return
	}

	// определяем пользователя по JWT
	idUser, err := h.service.ParseToken(token)
	if err != nil {
		logrus.Println(err)
		newErrorResponse(c, http.StatusServiceUnavailable, err.Error())
		return
	}

	// отправляем форму для нового пароля
	c.HTML(http.StatusOK, "new_pass.html", gin.H{
		"id": idUser,
	})
}

// восстановление пароля
func (h *Handler) recoveryPass(c *gin.Context) {

	var psw, refreshPsw, idUser string

	// передать idUser!!!

	// ContentType = text/plain
	// выделим тело запроса
	/* структура тела запроса {
		psw=<new_password>
		refresh_psw=<refresh_psw>
		id_user=<your_id>
	}*/
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		logrus.Println(err)
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	// выделим новый пароль и повторный пароль из body
	// разделим поля данных в запросе
	res := bytes.Split(body, []byte{13, 10})
	for _, params := range res {
		// делим строки по знаку равенства
		paramsSl := strings.Split(string(params), "=")

		if paramsSl[0] == "psw" {
			if paramsSl[1] == "" {
				logrus.Println("не передан пароль")
				newErrorResponse(c, http.StatusBadRequest, "Повторите попытку")
				return
			}
			psw = paramsSl[1]
			// log.Println(paramsSl[1])
		} else if paramsSl[0] == "refresh_psw" {
			if paramsSl[1] == "" {
				logrus.Println("не передан повторный пароль")
				newErrorResponse(c, http.StatusBadRequest, "Повторите попытку")
				return
			}
			refreshPsw = paramsSl[1]
			// log.Println(paramsSl[1])
		} else if paramsSl[0] == "id_user" {
			if paramsSl[1] == "" {
				logrus.Println("не передан idUser при восстановлении пароля")
				newErrorResponse(c, http.StatusBadRequest, "Повторите попытку")
				return
			}
			idUser = paramsSl[1]
			// log.Println(paramsSl[1])
		}
	}

	// захэшируем пароли и проверим, что они совпадают
	err = h.service.CheckPass(&psw, &refreshPsw)
	if err != nil {
		logrus.Println(err)
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// перезапишем новый пароль в БД
	err = h.service.UpdatePass(idUser, psw)
	if err != nil {
		logrus.Println(err)
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	// перенаправить на страницу авторизации
	// либо стразу выдать JWT???
	c.HTML(http.StatusOK, "login.html", gin.H{
		"error": "Пароль обновлен",
	})
}
