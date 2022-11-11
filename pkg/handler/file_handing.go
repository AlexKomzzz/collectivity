package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// получение файла от клиента
func (h *Handler) getFile(c *gin.Context) {

	// получение файла из запроса от клиента по ключевому слову "file"
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		logrus.Println("ошибка при получении файла от клиента: ", err)
		errorServerResponse(c, err)
		return
	}
	defer file.Close()

	// создание пустого файла в файловой системе
	err = h.service.SaveFileToServe(file)
	if err != nil {
		logrus.Println(errors.New("Handler/getFile()/ошибка при сохранении файла в системе: " + err.Error()))
		errorServerResponse(c, err)
		return
	}

	logrus.Println("Новая таблица загружена администратором")

	// запись данных из файла в БД
	err = h.service.ParseDataFileToDB()
	if err != nil {
		logrus.Println(errors.New("Handler/getFile()/ошибка при сохранении данных из файла в БД: " + err.Error()))
		errorServerResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"debt": "OK",
	})
	// c.HTML(http.StatusOK, "start_list.html", gin.H{
	// 	"role": true,
	// })
}

// отправка данных клиенту
func (h *Handler) dataClient(c *gin.Context) {

}
