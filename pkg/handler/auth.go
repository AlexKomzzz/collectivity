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
		c.HTML(http.StatusBadRequest, "forma_auth.html", gin.H{
			"err":    true,
			"msgErr": "Ошибка запроса. Введите, пожалуйста, все данные снова.",
		})
		return
	}

	// выделим данные из body и запишем в структуру User
	// разделим поля данных в запросе
	res := bytes.Split(body, []byte{13, 10})
	for _, params := range res {
		// делим строки по знаку равенства
		paramsSl := strings.Split(string(params), "=")

		if paramsSl[0] == "first-name" {
			if paramsSl[1] == "" {
				logrus.Println("не передано имя при создании нового пользователя")
				c.HTML(http.StatusBadRequest, "forma_auth.html", gin.H{
					"err":    true,
					"msgErr": "Ошибка запроса. Введите, пожалуйста, все данные снова.",
				})
				return
			}
			dataUser.FirstName = strings.TrimSpace(paramsSl[1])
			// log.Println(paramsSl[1])
		} else if paramsSl[0] == "last-name" {
			if paramsSl[1] == "" {
				logrus.Println("не передана фамилия при создании нового пользователя")
				c.HTML(http.StatusBadRequest, "forma_auth.html", gin.H{
					"err":    true,
					"msgErr": "Ошибка запроса. Введите, пожалуйста, все данные снова.",
				})
				return
			}
			dataUser.LastName = strings.TrimSpace(paramsSl[1])
			// log.Println(paramsSl[1])
		} else if paramsSl[0] == "middle-name" {
			if len(paramsSl) > 1 {
				dataUser.MiddleName = strings.TrimSpace(paramsSl[1])
			}
			// log.Println(paramsSl[1])
		} else if paramsSl[0] == "email" {
			if paramsSl[1] == "" {
				logrus.Println("не передан email при создании нового пользователя")
				c.HTML(http.StatusBadRequest, "forma_auth.html", gin.H{
					"err":    true,
					"msgErr": "Ошибка запроса. Введите, пожалуйста, все данные снова.",
				})
				return
			}
			dataUser.Email = strings.TrimSpace(paramsSl[1])
			// log.Println(paramsSl[1])
		} else if paramsSl[0] == "psw" {
			if paramsSl[1] == "" {
				logrus.Println("не передан пароль при создании нового пользователя")
				c.HTML(http.StatusBadRequest, "forma_auth.html", gin.H{
					"err":    true,
					"msgErr": "Ошибка запроса. Введите, пожалуйста, все данные снова.",
				})
				return
			}
			dataUser.Password = strings.TrimSpace(paramsSl[1])
			// log.Println(paramsSl[1])
		} else if paramsSl[0] == "psw-repeat" {
			if paramsSl[1] == "" {
				logrus.Println("не передан повторный пароль при создании нового пользователя")
				c.HTML(http.StatusBadRequest, "forma_auth.html", gin.H{
					"err":    true,
					"msgErr": "Ошибка запроса. Введите, пожалуйста, все данные снова.",
				})
				return
			}
			passRepeat = strings.TrimSpace(paramsSl[1])
			// log.Println(paramsSl[1])
		}
	}

	// составление ФИО
	dataUser.Username = fmt.Sprintf("%s %s %s", dataUser.LastName, dataUser.FirstName, dataUser.MiddleName)

	// Проверка на отсутствие пользователя с таким email в БД
	ok, err := h.service.CheckUserByEmail(dataUser.Email)
	if err != nil {
		logrus.Println("Handler/signUp(): ", err.Error())
		errorServerResponse(c, err)
		return
	}
	if ok {
		c.HTML(http.StatusBadRequest, "forma_auth.html", gin.H{
			"err":    true,
			"msgErr": "Пользователь с указанной электронной почтой уже зарегистрирован.",
		})
		return
	}

	// создание пользователя в БД authdata, проверка паролей
	idUser, err := h.service.CreateUserByAuth(&dataUser, passRepeat)
	if err != nil {
		if err.Error() == "пароли не совпадают" {
			logrus.Println("несовпадение паролей при регистации.", err)
			c.HTML(http.StatusBadRequest, "forma_auth.html", gin.H{
				"pass": true,
			})
		} else {
			logrus.Println("ошибка при создании пользователя в БД authdata: ", err)
			errorServerResponse(c, err)
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
		logrus.Println("Handler/signUp():", err)
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
		logrus.Println("Handler/signAdd(): отсутствие токена или email в URL при подтверждении почты")
		newErrorResponse(c, http.StatusBadRequest, "invalid URL")
		return
	}

	// определение пользователя по JWT
	idUserAuth, err := h.service.ParseTokenEmail(tokenByURL)
	if err != nil {
		logrus.Println("Handler/signAdd(): ", err)
		errorServerResponse(c, err)
		return
	}

	// получение данных пользователя из БД authdata
	dataUser, err := h.service.GetUserFromAuth(idUserAuth)
	if err != nil {
		logrus.Println("Handler/signAdd(): ", err)
		errorServerResponse(c, err)
		return
	}

	// сравнение email полученного и сохраненного в БД
	err = h.service.ComparisonEmail(dataUser.Email, emailByURL)
	if err != nil {
		logrus.Println("Handler/signAdd(): ", err)
		newErrorResponse(c, http.StatusBadRequest, "пользователь с данным адресом электронной почты не регистрировался")
		return
	}

	// создание пользователя в БД
	idUser, err := h.service.CreateUser(&dataUser)
	if err != nil {
		if idUser == -1 {
			logrus.Println("пользователь с таким ФИО не может быть зарегистрирован - ", dataUser.Username, ": ", err)
			c.HTML(http.StatusBadRequest, "forma_auth", gin.H{
				"err":    true,
				"msgErr": fmt.Sprintf("Пользователь с ФИО \"%s\" не может быть зарегистрирован", dataUser.Username),
			})
			return
		}
		logrus.Println("Handler/signAdd(): ", err)
		errorServerResponse(c, err)
		return
	}
	logrus.Println("Создан новый пользователь: ", idUser)

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
			c.HTML(http.StatusBadRequest, "login.html", gin.H{
				"error": "Пользователя с такой эл. почтой не существует. Проверьте правильность введенных данных или зарегистрируйтесь.",
			})
		} else if err.Error() == "пароль" {
			c.HTML(http.StatusBadRequest, "login.html", gin.H{
				"error": "Неверный пароль. Попробуйте снова.",
			})
		} else {
			logrus.Println("Handler/signIn(): ", err)
			errorServerResponse(c, err)
		}

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
		logrus.Println("Handler/definitionUser()/ReadAll(): ", err)
		c.HTML(http.StatusBadRequest, "new_pass_email.html", gin.H{
			"error": "Ошибка запроса. Повторите процедуру",
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
			if paramsSl[1] == "" {
				logrus.Println("не передан email для восстановления пароля")
				c.HTML(http.StatusBadRequest, "new_pass_email.html", gin.H{
					"error": "Ошибка запроса. Повторите процедуру.",
				})
				return
			}
			emailUser = strings.TrimSpace(paramsSl[1])
			// log.Println(paramsSl[1])
		}
	}

	// идентифицируем пользователя по email, получени id пользователя по email
	idUser, err := h.service.GetUserByEmail(emailUser)
	if err != nil {
		logrus.Println(err)
		if idUser == -1 {
			logrus.Println("Handler/definitionUser(): ", err)
			c.HTML(http.StatusBadRequest, "new_pass_email.html", gin.H{
				"error": "Пользователя с таким адресом электронной почты не существует.",
			})
		} else {
			logrus.Println("Handler/definitionUser(): ", err)
			errorServerResponse(c, err)
		}
		return
	}

	// генерация токена для отправки на почту для восстановление пароля
	token, err := h.service.GenerateJWTtoEmail(idUser)
	if err != nil {
		logrus.Println("Handler/definitionUser(): ", err)
		errorServerResponse(c, err)
		return
	}

	// формирование URL
	URL := fmt.Sprintf("%s/auth/pass/definition-userJWT?token=%s", viper.GetString("url"), url.PathEscape(token))
	logrus.Printf("URL: %s", URL)

	// текст письма
	msg := fmt.Sprintf("To: %s\r\n"+
		"Subject: Восстановление пароля\r\n"+
		"\r\n"+
		"Для восстановления пароля перейдите по ссылке: %s.\r\n", emailUser, URL)

	// отпрвка сообщения на почту пользователя с ссылкой для восстановления пароля
	err = h.service.SendMessageByMail(emailUser, URL, msg)
	if err != nil {
		logrus.Println("Handler/definitionUser(): ", err)
		errorServerResponse(c, err)
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
		logrus.Println("Handler/definitionUserJWT()/Query(): отсутствие токена в URL при восстановлении пароля при переходде по ссылке с почты")
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"error": "Ошибка запроса. Повторите процедуру.",
		})
		return
	}

	// определяем пользователя по JWT
	_, err := h.service.ParseTokenEmail(token)
	if err != nil {
		logrus.Println("Handler/definitionUserJWT(): ", err)
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"error": "Ошибка запроса. Повторите процедуру.",
		})
		return
	}

	// URLrequest := fmt.Sprintf("/auth/pass/recovery-pass?token=%s", url.PathEscape(token))

	// отправляем форму для нового пароля
	// передаем токен для последующего определения id
	c.HTML(http.StatusOK, "new_pass.html", gin.H{
		//	"id":     true,
		"id":    true,
		"token": token,
	})
}

// восстановление пароля
func (h *Handler) recoveryPass(c *gin.Context) {

	var psw, refreshPsw string

	// определение JWT из URL
	token := c.Query("token")
	if token == "" {
		logrus.Println("Handler/definitionUserJWT()/Query(): отсутствие токена в URL при задании нового пароля")
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"error": "Ошибка запроса. Повторите процедуру.",
		})
		return
	}

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
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// выделим новый пароль и повторный пароль и idUser из body
	// разделим поля данных в запросе
	res := bytes.Split(body, []byte{13, 10})
	for _, params := range res {
		// делим строки по знаку равенства
		paramsSl := strings.Split(string(params), "=")

		// if paramsSl[0] == "token" {
		// 	if paramsSl[1] == "" {
		// 		logrus.Println("не передан token при восстановлении пароля")
		// 		c.HTML(http.StatusBadRequest, "login.html", gin.H{
		// 			"error": "Ошибка запроса. Повторите процедуру.",
		// 		})
		// 		return
		// 	}
		// 	token = paramsSl[1]
		// 	// log.Println(paramsSl[1])
		if paramsSl[0] == "refresh_psw" {
			if paramsSl[1] == "" {
				logrus.Println("не передан повторный пароль для изменения")
				c.HTML(http.StatusBadRequest, "new_pass.html", gin.H{
					"id":     true,
					"token":  token,
					"err":    true,
					"msgErr": "Повторите новый пароль.",
				})
				return
			}
			refreshPsw = strings.TrimSpace(paramsSl[1])
			// log.Println(paramsSl[1])
		} else if paramsSl[0] == "psw" {
			if paramsSl[1] == "" {
				logrus.Println("не передан новый пароль для изменения")
				c.HTML(http.StatusBadRequest, "new_pass.html", gin.H{
					"id":     true,
					"token":  token,
					"err":    true,
					"msgErr": "Вы не указали новый пароль.",
				})
				return
			}
			psw = strings.TrimSpace(paramsSl[1])
			// log.Println(paramsSl[1])
		}
	}

	// определяем пользователя по JWT
	idUser, err := h.service.ParseTokenEmail(token)
	if err != nil {
		logrus.Println(err)
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"error": "Ошибка запроса. Повторите процедуру.",
		})
		return
	}

	// захэшируем пароли и проверим, что они совпадают
	err = h.service.CheckPass(&psw, &refreshPsw)
	if err != nil {
		logrus.Println("Handler/recoveryPass(): ", err)
		c.HTML(http.StatusBadRequest, "new_pass.html", gin.H{
			"id":     true,
			"token":  token,
			"err":    true,
			"msgErr": "Пароли не совпадают",
		})
		return
	}

	// перезапишем новый пароль в БД
	err = h.service.UpdatePass(idUser, psw)
	if err != nil {
		logrus.Println("Handler/recoveryPass(): ", err)
		errorServerResponse(c, err)
		return
	}

	logrus.Println("Обновлен пароль для пользователя: ", idUser)

	// сгенерируем JWT
	// tokenJWT, err := h.service.GenerateJWT_API(idUser)
	// if err != nil {
	// 	errorServerResponse(c, err)
	// 	return
	// }

	// перенаправить на страницу авторизации
	// либо стразу выдать JWT???
	c.HTML(http.StatusOK, "login.html", gin.H{
		"error": "Пароль успешно обновлен",
	})
	// c.JSON(http.StatusOK, gin.H{
	// 	"token": tokenJWT,
	// })
}

// создание админа
func (h *Handler) createAdm(c *gin.Context) {
	err := h.service.CreateAdmin()
	if err != nil {
		logrus.Println("Handler/createAdm(): ", err)
		errorServerResponse(c, err)
		return
	}

	logrus.Println("Создание Админа в БД")
	c.JSON(http.StatusOK, gin.H{
		"admin": "ok",
	})
}
