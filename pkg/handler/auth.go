package handler

import (
	"encoding/json"
	"errors"
	"fmt"

	"net/http"
	"net/url"

	app "github.com/AlexKomzzz/collectivity"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const startListURI = "https://localhost:8080/startList"

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

// создание пользователя при получении данных с помощью OAuth (Google или Яндекс)
func (h *Handler) createUserOAuth(c *gin.Context) {

	dataUser := &app.User{}

	// определение JWT_API из URL
	tokenAPI := c.Query("token")
	if tokenAPI == "" {
		logrus.Println("Handler/createUserOAuth()/Query(): отсутствие токена в URL при задании нового пароля")
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"error": "Ошибка запроса. Повторите процедуру.",
		})
		return
	}

	middleName := c.PostForm("middle-name")
	if middleName == "" {
		logrus.Println("не передано отчество в теле запроса при регистрации через OAuth2")
		c.HTML(http.StatusBadRequest, "middle_names.html", gin.H{
			"err":    true,
			"msgErr": "Ошибка запроса. Введите, пожалуйста, данные снова",
		})
		return
	}

	// восстановить idUserAPI из JWT
	idUserAPI, err := h.service.ParseToken(tokenAPI)
	if err != nil {
		logrus.Println("Handler/createUserOAuth()/ParseToken()/ ошибка при парсе JWT: ", err)
		if idUserAPI == -1 {
			c.HTML(http.StatusBadRequest, "login.html", gin.H{
				"error": "Истекло выделенное время, повторите процедуру",
			})
		} else {
			c.HTML(http.StatusInternalServerError, "login.html", gin.H{
				"error": "Непредвиденная ошибка, пожалуйста, повторите.",
			})
		}
		return
	}

	// считать данные пользователя из кэша по idUserAPI
	dataCash, err := h.service.GetUserCash(idUserAPI)
	if err != nil {
		logrus.Println("Handler/createUserOAuth()/GetUserCash()/ ошибка при считывании данных из кэша: ", err)
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{
			"error": "Непредвиденная ошибка, пожалуйста, повторите.",
		})
		return
	}

	err = json.Unmarshal(dataCash, &dataUser)
	if err != nil {
		logrus.Println("Handler/createUserOAuth()/Unmarshal()/ ошибка при декодировании данных кэша в структуру пользователя: ", err)
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{
			"error": "Непредвиденная ошибка, пожалуйста, повторите.",
		})
		return
	}

	dataUser.MiddleName = middleName
	// составление ФИО
	dataUser.Username = fmt.Sprintf("%s %s %s", dataUser.LastName, dataUser.FirstName, middleName)

	// создание пользователя в БД (или обновление, если уже создан)
	idUser, err := h.service.CreateUser(dataUser)
	if err != nil {
		logrus.Println("Handler/createUserOAuth(): ", dataUser.Username, " - ", err)
		if idUser == -1 {
			c.HTML(http.StatusInternalServerError, "login.html", gin.H{
				"error": fmt.Sprintf("Пользователь с ФИО \"%s\" не может быть зарегистрирован", dataUser.Username),
			})
			return
		}
		errorServerResponse(c, err)
		return
	}

	/*idUser, err := h.service.CreateUserAPI("yandex", userData.Id, userData.FirstName, userData.LastName, userData.Email)
	if err != nil {
		logrus.Println(err)
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{
			"error": "Непредвиденная ошибка, пожалуйста, повторите.",
		})
		return
	}*/

	// получение JWT по id пользователя
	token, err := h.service.GenerateJWTbyID(idUser)
	if err != nil {
		logrus.Println("Handler/createUserOAuth(): ", err)
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{
			"error": "Непредвиденная ошибка, пожалуйста, повторите.",
		})
		return
	}

	logrus.Printf("JWT для пользователя %d: %s\n", idUser, token)

	// передача JWT токена пользователю
	// c.JSON(http.StatusOK, gin.H{
	// 	"token": token,
	// })

	// создание сессии
	session := sessions.Default(c)

	// запись в куки JWT
	session.Set("token", token)
	session.Save()

	logrus.Println("Запись куки-файла при OAuth2")

	c.Redirect(http.StatusTemporaryRedirect, startListURI)
}

// получение данных при создании нового пользователя, запись в БД регистрации, отправка ссылки с токеном и email на почту для подтверждения эл.почты
func (h *Handler) signUp(c *gin.Context) { // Обработчик для регистрации

	dataUser := &app.User{}

	//  извлечение данных из пост запроса
	dataUser.FirstName = c.PostForm("first-name")
	dataUser.LastName = c.PostForm("last-name")
	dataUser.MiddleName = c.PostForm("middle-name")
	dataUser.Email = c.PostForm("email")
	dataUser.Password = c.PostForm("psw")
	passRepeat := c.PostForm("psw-repeat")

	if dataUser.FirstName == "" || dataUser.LastName == "" || dataUser.Password == "" || passRepeat == "" || dataUser.Email == "" {
		logrus.Println("получены не все данные при создании нового пользователя")
		c.HTML(http.StatusBadRequest, "forma_auth.html", gin.H{
			"err":    true,
			"msgErr": "Ошибка запроса. Введите, пожалуйста, все данные снова.",
		})
		return
	}

	// составление ФИО
	dataUser.Username = fmt.Sprintf("%s %s %s", dataUser.LastName, dataUser.FirstName, dataUser.MiddleName)

	// проверка на существование пользователя с таким ФИО в БД users и получение idUser
	idUser, err := h.service.GetUserByUsername(dataUser.Username)
	if err != nil {
		logrus.Println("Handler/signUp()/CheckUserByMiddleNames(): ", err.Error())
		errorServerResponse(c, err)
		return
	}
	if idUser == -1 {
		c.HTML(http.StatusBadRequest, "forma_auth.html", gin.H{
			"err":    true,
			"msgErr": fmt.Sprintf("Пользователь с ФИО \"%s\" не может быть зарегистрирован", dataUser.Username),
		})
		return
	}

	dataUser.Id = idUser

	// Проверка на отсутствие пользователя с таким email в БД
	ok, err := h.service.CheckUserByEmail(dataUser.Email)
	if err != nil {
		logrus.Println("Handler/signUp()/CheckUserByEmail(): ", err.Error())
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

	// проверка на совпадение паролей
	err = h.service.CheckPass(&dataUser.Password, &passRepeat)
	if err != nil {
		logrus.Println("Handler/signUp()/CheckPass(): ", err)
		c.HTML(http.StatusBadRequest, "forma_auth.html", gin.H{
			"pass": true,
		})
		return
	}

	// кэширование данных пользователя
	// декодировка структуры пользователя в слайз байт для кэширования в redis
	dataCash, err := json.Marshal(dataUser)
	if err != nil {
		logrus.Println("Handler/signUp()/Marshal()/ ошибка при декодировке данных пользователя: ", err)
		errorServerResponse(c, err)
		return
	}

	// запись данных о пользователе в кэш
	// ключ в кэше - ID
	err = h.service.SetUserCash(idUser, dataCash)
	if err != nil {
		logrus.Println("Handler/signUp()/SetUserCash()/ ошибка при записи данных в кэш: ", err)
		errorServerResponse(c, err)
		return
	}

	// создание пользователя в БД authdata, проверка паролей
	/*idUser, err := h.service.CreateUserByAuth(&dataUser, passRepeat)
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
	}*/
	// logrus.Printf("idUser: %d", idUser)

	// генерация токена для подтверждения почты
	tokenVerific, err := h.service.GenerateJWTtoEmail(idUser)
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
	URL := fmt.Sprintf("%s/auth/sign-add?token=%s&email=%s", viper.GetString("url"), url.PathEscape(tokenVerific), url.PathEscape(dataUser.Email))
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
// сравнение email полученного и сохраненного в БД, создание нового пользователя в БД users, редирект на стр входа
func (h *Handler) signAdd(c *gin.Context) { // Обработчик для регистрации

	dataUser := &app.User{}

	// определение токена и email из URL
	tokenVerific := c.Query("token")
	emailByURL := c.Query("email")
	if tokenVerific == "" || emailByURL == "" {
		logrus.Println("Handler/signAdd(): отсутствие токена или email в URL при подтверждении почты")
		newErrorResponse(c, http.StatusBadRequest, "invalid URL")
		return
	}

	// определение пользователя по JWT
	idUser, err := h.service.ParseTokenEmail(tokenVerific)
	if err != nil {
		logrus.Println("Handler/signAdd(): ", err)
		if idUser == -1 {
			c.HTML(http.StatusBadRequest, "login.html", gin.H{
				"error": "Истекло выделенное время, повторите процедуру",
			})
		} else {
			errorServerResponse(c, err)
		}
		return
	}

	// получение данных пользователя из кэша
	dataUserCash, err := h.service.GetUserCash(idUser)
	if err != nil {
		logrus.Println("Handler/signAdd()/GetUserCash()/ ошибка при считывании данных из кэша: ", err)
		errorServerResponse(c, err)
		return
	}

	err = json.Unmarshal(dataUserCash, dataUser)
	if err != nil {
		logrus.Println("Handler/signAdd()/Unmarshal()/ ошибка при декодировании данных кэша в структуру пользователя: ", err)
		errorServerResponse(c, err)
		return
	}

	// получение данных пользователя из БД authdata
	/*dataUser, err := h.service.GetUserFromAuth(idUserAuth)
	if err != nil {
		logrus.Println("Handler/signAdd(): ", err)
		errorServerResponse(c, err)
		return
	}*/

	// сравнение idUser из JWT и из кэша
	err = h.service.ComparisonId(idUser, dataUser.Id)
	if err != nil {
		logrus.Println("Handler/signAdd()/ComparisonId(): ", err)
		errorServerResponse(c, err)
		return
	}

	// сравнение email полученного из кэша и из URL
	err = h.service.ComparisonEmail(dataUser.Email, emailByURL)
	if err != nil {
		logrus.Println("Handler/signAdd()/ComparisonEmail(): ", err)
		errorServerResponse(c, errors.New("пользователь с данным адресом электронной почты не регистрировался"))
		return
	}

	// создание пользователя в БД auth (idUser уже известен)
	err = h.service.CreateUserById(dataUser)
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
	// token, err := h.service.GenerateJWTbyID(idUser)
	// if err != nil {
	// 	logrus.Println("ошибка при генерации JWT в signAdd: ", err)
	// 	errorServerResponse(c, err)
	// 	return
	// }
	// редирект на стартовую страницу
	// c.Redirect(http.StatusTemporaryRedirect, startList+token)

	// выдача JWT
	// c.JSON(http.StatusOK, gin.H{
	// 	"token": token,
	// })

	c.HTML(http.StatusOK, "login.html", gin.H{
		"msg": "Учетная запись успешно создана. Используйте адрес электронной почты и пароль для входа",
	})
}

// аутентификация пользователя, генерация JWT, создание сессии
func (h *Handler) signIn(c *gin.Context) {

	var dataUser app.User

	// получение данных о пользователе из формы
	dataUser.Email = c.PostForm("email")
	dataUser.Password = c.PostForm("password")
	if dataUser.Email == "" || dataUser.Password == "" {
		logrus.Println("не передан email или пароль при аутентификации")
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"error": "Ошибка при передачи данных, пожалуйста, повторите",
		})
		return
	}
	// log.Println("newemail: ", email)
	// log.Println("newpass: ", pass)

	/*
		// ContentType = text/plain
		// выделим тело запроса
		 структура тела запроса {
			email=<your_email>
			password=<your_pass>
			btn_login=
		}
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
				if paramsSl[1] == "" {
					logrus.Println("не передан email при аутентификации")
					c.HTML(http.StatusBadRequest, "login.html", gin.H{
						"error": "Не получен адрес эл. почты, повторите",
					})
					return
				}
				dataUser.Email = strings.TrimSpace(paramsSl[1])
				log.Println("email: ", paramsSl[1])
			} else if paramsSl[0] == "password" {
				if paramsSl[1] == "" {
					logrus.Println("не передан пароль при аутентификации")
					c.HTML(http.StatusBadRequest, "login.html", gin.H{
						"error": "Не получен пароль, повторите",
					})
					return
				}
				dataUser.Password = strings.TrimSpace(paramsSl[1])
				log.Println("pass: ", paramsSl[1])
			}
		}*/

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
			// } else if err.Error() == "api" {
			// 	c.HTML(http.StatusBadRequest, "login.html", gin.H{
			// 		"error": "Для входа в аккаунт с данной эл. почтой используйте кнопки входа через Google или Яндекс почту",
			// 	})
		} else {
			logrus.Println("Handler/signIn(): ", err)
			errorServerResponse(c, err)
		}

		return
	}

	// создание сессии
	session := sessions.Default(c)

	// запись в куки JWT
	session.Set("token", token)
	session.Save()

	logrus.Println("Запись куки-файла при email и пароле")

	// редирект на стартовую страницу
	c.Redirect(http.StatusTemporaryRedirect, startListURI)
}

// удаление сессии пользователя
func (h *Handler) signOut(c *gin.Context) {

	session := sessions.Default(c)
	session.Clear()
	session.Save()

	logrus.Println("Выход из учетной записи")

	c.HTML(http.StatusBadRequest, "login.html", gin.H{})
}

// определение пользователя по email при восстановлении пароля, отправка письма на почту с токеном
func (h *Handler) definitionUser(c *gin.Context) {

	emailUser := c.PostForm("email")
	if emailUser == "" {
		logrus.Println("не передан email для восстановления пароля")
		c.HTML(http.StatusBadRequest, "new_pass_email.html", gin.H{
			"error": "Ошибка запроса. Повторите процедуру.",
		})
		return
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
	URL := fmt.Sprintf("%s/auth/pass/definition-userJWT?token=%s&email=%s", viper.GetString("url"), url.PathEscape(token), url.PathEscape(emailUser))
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

	// определение JWT и email из URL
	token := c.Query("token")
	if token == "" {
		logrus.Println("Handler/definitionUserJWT()/Query(): отсутствие токена в URL при восстановлении пароля при переходде по ссылке с почты")
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"error": "Ошибка запроса. Повторите процедуру.",
		})
		return
	}

	emailUser := c.Query("email")
	if emailUser == "" {
		logrus.Println("Handler/definitionUserJWT()/Query(): отсутствие email в URL при восстановлении пароля при переходде по ссылке с почты")
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"error": "Ошибка запроса. Повторите процедуру.",
		})
		return
	}

	// определяем пользователя по JWT
	idUser, err := h.service.ParseTokenEmail(token)
	if err != nil {
		logrus.Println("Handler/definitionUserJWT(): ", err)
		if idUser == -1 {
			c.HTML(http.StatusBadRequest, "login.html", gin.H{
				"error": "Истекло выделенное время, повторите процедуру",
			})
		} else {
			c.HTML(http.StatusBadRequest, "login.html", gin.H{
				"error": "Ошибка запроса. Повторите процедуру.",
			})
		}
		return
	}

	// URLrequest := fmt.Sprintf("/auth/pass/recovery-pass?token=%s", url.PathEscape(token))

	// отправляем форму для нового пароля
	// передаем токен для последующего определения id
	c.HTML(http.StatusOK, "new_pass.html", gin.H{
		//	"id":     true,
		"id":    true,
		"token": token,
		"email": emailUser,
	})
}

// восстановление пароля
func (h *Handler) recoveryPass(c *gin.Context) {

	// определение JWT из URL
	token := c.Query("token")
	if token == "" {
		logrus.Println("Handler/definitionUserJWT()/Query(): отсутствие токена в URL при задании нового пароля")
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"error": "Ошибка запроса. Повторите процедуру.",
		})
		return
	}

	// определение email из URL
	emailUser := c.Query("email")
	if emailUser == "" {
		logrus.Println("Handler/definitionUserJWT()/Query(): отсутствие email в URL при задании нового пароля")
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"error": "Ошибка запроса. Повторите процедуру.",
		})
		return
	}

	psw := c.PostForm("psw")
	refreshPsw := c.PostForm("refresh_psw")
	if psw == "" || refreshPsw == "" {
		logrus.Println("не передан повторный или основной пароль для изменения")
		c.HTML(http.StatusBadRequest, "new_pass.html", gin.H{
			"id":     true,
			"token":  token,
			"err":    true,
			"msgErr": "Повторите введение паролей",
		})
		return
	}

	// определяем пользователя по JWT
	idUser, err := h.service.ParseTokenEmail(token)
	if err != nil {
		logrus.Println("Handler/recoveryPass()/ParseTokenEmail():", err)

		if idUser == -1 {
			c.HTML(http.StatusBadRequest, "login.html", gin.H{
				"error": "Истекло выделенное время, повторите процедуру",
			})
		} else {
			c.HTML(http.StatusBadRequest, "login.html", gin.H{
				"error": "Ошибка запроса. Повторите процедуру.",
			})
		}
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
	err = h.service.UpdatePass(idUser, emailUser, psw)
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
		"error": "Пароль успешно обновлен. Можете использовать его для входа в учетную запись",
	})
	// c.JSON(http.StatusOK, gin.H{
	// 	"token": tokenJWT,
	// })
}
