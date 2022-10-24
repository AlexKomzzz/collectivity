package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Resp struct {
	Email    string `form:"email"`
	Password string `form:"password"`
}

func (h *Handler) test(c *gin.Context) {
	// c.HTML(http.StatusOK, "ex.html", gin.H{
	// 	"req": "Alex",
	// })
	logrus.Println("test")
	logrus.Println(c.ContentType())
	// body, _ := ioutil.ReadAll(c.Request.Body)
	// res := strings.ReplaceAll(string(body), "\n", " ")
	// resSL := strings.Split(res, " ")
	// log.Println(string(body))
}
