package handler

import (
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
)

// название файла
const fileName = "table.xlsx"

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
	// os.Mkdir("./web/files/", 0777)
	saveFile, err := os.OpenFile("./web/files/"+fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		logrus.Println("ошибка при создании файла в системе: ", err)
		errorServerResponse(c, err)
		return
	}
	defer saveFile.Close()

	// копирование данных из запроса в созданный файл
	io.Copy(saveFile, file)

	logrus.Println("Новая таблица загружена администратором")

	c.HTML(http.StatusOK, "start_list.html", gin.H{})
}

// отправка данных клиенту
func (h *Handler) dataClient(c *gin.Context) {

	// открытие файла
	f, err := excelize.OpenFile("./web/files/" + fileName)
	if err != nil {
		logrus.Println("ошибка при открытии файла из системы: ", err)
		errorServerResponse(c, err)
		return
	}
	defer func() {
		// Close the spreadsheet.
		if err := f.Close(); err != nil {
			logrus.Println("ошибка при закрытии файла: ", err)

		}
	}()

	// получение данных из ячейки таблицы
	// cell, err := f.GetCellValue("Sheet1", "A2")
	// if err != nil {
	// 	logrus.Println("ошибка при получении данных из таблицы: ", err)
	// 	errorServerResponse(c, err)
	// 	return
	// }

	// получение строк из таблицы на листе Sheet1
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		logrus.Println("ошибка при получении данных из таблицы: ", err)
		errorServerResponse(c, err)
		return
	}
	for i, row := range rows {
		for j, colCell := range row {

			if colCell == "bob" {
				logrus.Println("значение Боба равно ", rows[i][j+1])
				c.JSON(http.StatusOK, gin.H{
					"значение Боба": rows[i][j+1],
				})
			}

			//logrus.Print(colCell, "\t")
		}
		//logrus.Println()
	}
}
