package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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
	} else {
		logrus.Println("обычный пользователь")
		c.HTML(http.StatusOK, "start_list.html", gin.H{})
		return
	}
}
