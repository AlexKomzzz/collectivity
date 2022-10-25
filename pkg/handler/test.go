package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Resp struct {
	Email    string `form:"email"`
	Password string `form:"password"`
}

func (h *Handler) test(c *gin.Context) {
	c.HTML(http.StatusBadRequest, "login.html", gin.H{
		"error": "Такого пользователя не существует",
	})
}
