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

	if roleUser == "admin" {
		// значит это админ
		logrus.Println("вход админа")
		c.HTML(http.StatusOK, "start_list.html", gin.H{
			"role": roleUser,
		})
		return
	} else {
		logrus.Println("обычный пользователь")
		c.HTML(http.StatusOK, "start_list.html", gin.H{})
		return
	}
}

// получение данных при создании нового пользователя, запись в БД регистрации, отправка ссылки с токеном и email на почту для подтверждения эл.почты
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

	// logrus.Printf("dataUser: %v", dataUser)

	// создание пользователя в БД
	idUser, err := h.service.CreateUserByAuth(&dataUser, passRepeat)
	if err != nil {
		if err.Error() == "пароли не совпадают" {
			logrus.Println("несовпадение паролей при регистации.", err)
			c.HTML(http.StatusBadRequest, "forma_auth.html", gin.H{
				"pass": err,
			})
		} else {
			logrus.Println("ошибка при создании пользователя в БД authdata: ", err)
			newErrorResponse(c, http.StatusInternalServerError, err.Error())
		}
		return
	}
	// logrus.Printf("idUser: %d", idUser)

	// генерация токена для подтверждения почты
	token, err := h.service.GenerateJWTtoEmail(idUser)
	if err != nil {
		logrus.Println("ошибка при генерации токена для подтверждения почты: ", err)
		errorServerResponse(c, err)
		// c.HTML(http.StatusInternalServerError, "login.html", gin.H{
		// 	"error": err,
		// })
		return
	}
	// logrus.Printf("token: %s", token)

	// формирование ссылки для отправки, в которой содержится токен
	URL := fmt.Sprintf("%s/auth/sign-add?token=%s&email=%s", viper.GetString("url"), url.PathEscape(token), url.PathEscape(dataUser.Email))
	logrus.Printf("URL: %s", URL)

	// текст письма
	msg := fmt.Sprintf("To: %s\r\n"+
		"Subject: Подтверждение адреса электронной почты\r\n"+
		"\r\n"+
		"Для подтверждения эл. почты перейдите по ссылке: %s.\r\n", dataUser.Email, URL)

	// отправка письма на почту пользователя, для подтверждения email
	err = h.service.SendMessageByMail(dataUser.Email, URL, msg)
	if err != nil {
		logrus.Println("ошибка при отправке ссылки на подтверждение email на почту пользователя: ", err)
		errorServerResponse(c, err)
		return
	}

	// отправка HTML формы
	c.HTML(http.StatusOK, "go_email.html", gin.H{
		"msg": "подтверждения электронного адреса",
	})
}

// поинт при переходе по ссылке на подтверждение эл.почты для создания нового пользователя
// выделение токена и email из URL, выделение idUser из токена, получение данных пользователя из БД authdata,
// сравнение email полученного и сохраненного в БД, создание нового пользователя в БД users, генерация и выдача JWT
func (h *Handler) signAdd(c *gin.Context) { // Обработчик для регистрации

	// определение токена и email из URL
	tokenByURL := c.Query("token")
	emailByURL := c.Query("email")
	if tokenByURL == "" || emailByURL == "" {
		logrus.Println("отсутствие токена или email в URL при подтверждении почты")
		newErrorResponse(c, http.StatusBadRequest, "invalid URL")
		return
	}

	// определение пользователя по JWT
	idUserAuth, err := h.service.ParseTokenEmail(tokenByURL)
	if err != nil {
		logrus.Println("ошибка при парсе токена при подтверждении почты: ", err)
		errorServerResponse(c, err)
		return
	}

	// получение данных пользователя из БД authdata
	dataUser, err := h.service.GetUserFromAuth(idUserAuth)
	if err != nil {
		logrus.Println("ошибка при получении данных пользователя из БД authdata: ", err)
		errorServerResponse(c, err)
		return
	}

	// сравнение email полученного и сохраненного в БД
	err = h.service.ComparisonEmail(dataUser.Email, emailByURL)
	if err != nil {
		logrus.Println("неуспешное сравнение emails в signAdd: ", err)
		newErrorResponse(c, http.StatusBadRequest, "пользователь с данным адресом электронной почты не регистрировался")
		return
	}

	// создание пользователя в БД
	idUser, err := h.service.CreateUser(&dataUser)
	if err != nil {
		logrus.Println("ошибка при создании пользователя в БД users: ", err)
		errorServerResponse(c, err)
		return
	}

	// генерация JWT по id
	token, err := h.service.GenerateJWT_API(idUser)
	if err != nil {
		logrus.Println("ошибка при генерации JWT в signAdd: ", err)
		errorServerResponse(c, err)
		return
	}

	// выдача JWT
	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

// аутентификация пользователя, выдача JWT
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
		logrus.Println("ошибка при выделении тела запроса в signIn: ", err)
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
		logrus.Println("ошибка при генерации JWT в signIn: ", err)
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

// определение пользователя по email при восстановлении пароля, отправка письма на почту с токеном
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

	// идентифицируем пользователя по email, получени id пользователя по email
	idUser, err := h.service.GetUserByEmail(emailUser)
	if err != nil {
		logrus.Println(err)
		if err.Error() == "sql: no rows in result set" {
			c.HTML(http.StatusBadRequest, "recovery_pass.html", gin.H{
				"error": "Пользователя с такой почтой не существует.",
			})
		} else {
			c.HTML(http.StatusBadRequest, "recovery_pass.html", gin.H{
				"error": err,
			})
		}
		return
	}

	// генерация токена для отправки на почту для восстановление пароля
	token, err := h.service.GenerateJWTtoEmail(idUser)
	if err != nil {
		logrus.Println("ошибка при генерации токена для восстановления пароля: ", err)
		c.HTML(http.StatusBadRequest, "recovery_pass.html", gin.H{
			"error": err,
		})
		return
	}

	// формирование URL
	URL := fmt.Sprintf("%s/auth/pass/definition-userJWT?token=%s", viper.GetString("url"), url.PathEscape(token))

	// текст письма
	msg := fmt.Sprintf("To: %s\r\n"+
		"Subject: Восстановление пароля\r\n"+
		"\r\n"+
		"Для восстановления пароля перейдите по ссылке: %s.\r\n", emailUser, URL)

	// отпрвка сообщения на почту пользователя с ссылкой для восстановления пароля
	err = h.service.SendMessageByMail(emailUser, URL, msg)
	if err != nil {
		logrus.Println(err)
		c.HTML(http.StatusBadRequest, "recovery_pass.html", gin.H{
			"error": err,
		})
		return
	}

	// отправка HTML формы
	c.HTML(http.StatusOK, "go_email.html", gin.H{
		"msg": "восстановления пароля",
	})
}

// поинт при переходе по ссылке на восстановление пароля
// выделение токена из url, определение пользователя по JWT
func (h *Handler) definitionUserJWT(c *gin.Context) {

	// определение JWT из URL
	token := c.Query("token")
	if token == "" {
		logrus.Println("отсутствие токена в URL при восстановлении пароля")
		newErrorResponse(c, http.StatusBadRequest, "empty auth token")
		return
	}

	// определяем пользователя по JWT
	idUser, err := h.service.ParseTokenEmail(token)
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

	var psw, refreshPsw, idUserStr string

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

	// выделим новый пароль и повторный пароль и idUser из body
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
			idUserStr = paramsSl[1]
			// log.Println(paramsSl[1])
		}
	}

	// конвертируем idUser в числовое представление из строкового
	idUser, err := h.service.ConvIdUser(idUserStr)
	if err != nil {
		logrus.Println("ошибка конвертации idUser из string в int при восстановлении пароля: ", err)
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
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

	// сгенерируем JWT
	token, err := h.service.GenerateJWT_API(idUser)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{
			"error": err,
		})
		//newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	// перенаправить на страницу авторизации
	// либо стразу выдать JWT???
	// c.HTML(http.StatusOK, "login.html", gin.H{
	// 	"error": "Пароль обновлен",
	// })
	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}
