package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Resp struct {
	Email    string `form:"email"`
	Password string `form:"password"`
}

func (h *Handler) test(c *gin.Context) {
	err := h.service.SendMessage("komalex203@gmail.com", "ссылка")
	if err != nil {
		logrus.Println(err)
		newErrorResponse(c, http.StatusServiceUnavailable, err.Error())
		return
	}
}
