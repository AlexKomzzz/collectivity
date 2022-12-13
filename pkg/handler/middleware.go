package handler

const (
// authorizationHeader = "Authorization"
// userCtx             = "userId"
)

/*func (h *Handler) userIdentity(c *gin.Context) (int, error) {

	// Получение токена из файла куки
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

	// запись idUser в контекст
	c.Set(userCtx, userId)
}*/

/*
func getUserId(c *gin.Context) (int, error) {
	id, ok := c.Get(userCtx)
	if !ok {
		//newErrorResponse(c, http.StatusInternalServerError, "user id not found")
		return 0, errors.New("user id not found")
	}

	idInt, ok := id.(int)
	if !ok {
		//newErrorResponse(c, http.StatusInternalServerError, "user id is of invalid type")
		return 0, errors.New("user id is of invalid type")
	}

	return idInt, nil
}*/
