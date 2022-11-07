package handler

import (
	"github.com/gin-gonic/gin"
)

type Resp struct {
	Email    string `form:"email"`
	Password string `form:"password"`
}

func (h *Handler) test(c *gin.Context) {
	// err := h.service.SendMessage("komalex203@gmail.com", "ссылка")
	// if err != nil {
	// 	logrus.Println(err)
	// 	newErrorResponse(c, http.StatusServiceUnavailable, err.Error())
	// 	return
	// }
}
