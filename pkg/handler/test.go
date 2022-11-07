package handler

import (
	"github.com/gin-gonic/gin"
)

type Resp struct {
	Email    string `form:"email"`
	Password string `form:"password"`
}

func (h *Handler) test(c *gin.Context) {
	// c.HTML(http.StatusBadRequest, "forma_auth.html", gin.H{
	// 	"pass":   true,
	// 	"msgErr": "Ошибка запроса. Введите, пожалуйста, все данные снова.",
	// })
}
