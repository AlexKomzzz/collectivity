package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type errorResponse struct {
	Message string `json:"message"`
}

func newErrorResponse(c *gin.Context, statusCode int, message string) { // Ф-я обработчик ошибки
	logrus.Error(message)
	c.AbortWithStatusJSON(statusCode, errorResponse{Message: message})
}

func errorServerResponse(c *gin.Context, err error) { // Ф-я обработчик ошибки
	logrus.Error(err)
	c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse{Message: "ошибка сервера, пожалуйста, обновите страницу"})
}

func Response404(c *gin.Context) {
	logrus.Println("Error 404. Not found. Invalid url")
	c.HTML(http.StatusNotFound, "error.html", gin.H{
		"code":  http.StatusNotFound,
		"error": "It looks like one of the  developers fell asleep",
	})
}
