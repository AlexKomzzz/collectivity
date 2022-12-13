package handler

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// стартовая страница
func (h *Handler) startList(c *gin.Context) {

	// идентифицируем пользователя
	// idUser, err := h.userIdentity(c)
	// if err != nil {
	// 	if idUser == -1 {
	// 		logrus.Println("Вход c протухшим JWT")
	// 		c.HTML(http.StatusBadRequest, "login.html", gin.H{
	// 			"error": "Истекло выделенное время, повторите процедуру",
	// 		})
	// 	} else {
	// 		logrus.Println("Вход без идентификации")
	// 		c.Redirect(http.StatusTemporaryRedirect, "/auth/login")
	// 	}
	// 	return
	// }

	// получение токена из URL
	// определение JWT и email из URL
	// token := c.Query("token")
	// if token == "" {
	// 	logrus.Println("Вход без идентификации")
	// 	c.HTML(http.StatusBadRequest, "login.html", gin.H{})
	// 	return
	// }

	// Получение тока из файла куки
	session := sessions.Default(c)
	sessionToken := session.Get("token")
	if sessionToken == nil {
		logrus.Println("Вход без идентификации")
		c.HTML(http.StatusBadRequest, "login.html", gin.H{})
		return
	}

	// парсинг токена
	idUser, err := h.service.ParseToken(sessionToken.(string))
	if err != nil {
		if idUser == -1 {
			logrus.Println("Вход c протухшим JWT")
			c.HTML(http.StatusBadRequest, "login.html", gin.H{
				"error": "Истекло время сеанса",
			})
		} else {
			logrus.Println("Вход с неверным JWT")
			c.HTML(http.StatusBadRequest, "login.html", gin.H{})
		}
		return
	}

	// проверка роли пользователя
	roleUser, err := h.service.GetRole(idUser)
	if err != nil {
		logrus.Println("ошибка при проверке роли пользователя в БД при входе на стартовую страницу: ", err)
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	if roleUser == "admin" {
		// значит это админ
		logrus.Println("вход админа")
		c.HTML(http.StatusOK, "start_list.html", gin.H{
			"role": true,
		})
		return
	}

	logrus.Println("Вход пользователя: ", idUser)

	// определения debt для пользователя
	debtUser, err := h.service.GetDebtUser(idUser)
	if err != nil {
		logrus.Println("Handler/startList()/GetDebtUser(): ошибка при определении долга пользователя при входе на стартовую страницу: ", err)
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"error": "Ошибка запроса. Повторите процедуру.",
		})
		return
	}

	c.HTML(http.StatusOK, "start_list.html", gin.H{
		"resp": true,
		"debt": debtUser,
	})
}
