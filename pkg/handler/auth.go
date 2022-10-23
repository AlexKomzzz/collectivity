package handler

import (
	"fmt"
	"net/http"

	app "github.com/AlexKomzzz/collectivity"
	"github.com/gin-gonic/gin"
)

func (h *Handler) signUp(c *gin.Context) { // Обработчик для регистрации
	var input app.User

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body") //fmt.Sprintf("invalid input body: %s", err.Error()))
		return
	}

	id, err := h.service.CreateUser(input)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}

type signInInput struct { // Структура для идентификации
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) signIn(c *gin.Context) { // Обработчик для аутентификации и получения токена
	var input signInInput

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("invalid input body: %s", err.Error()))
		return
	}

	// token, err := h.service.GenerateJWT(input.Username, input.Password)
	// if err != nil {
	// 	newErrorResponse(c, http.StatusInternalServerError, err.Error())
	// 	return
	// }

	c.JSON(http.StatusOK, map[string]interface{}{
		"token": "token",
	})
}
