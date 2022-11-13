package service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"strings"

	app "github.com/AlexKomzzz/collectivity"
	"github.com/AlexKomzzz/collectivity/pkg/repository"
	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
)

// название файла
const fileName = "table.xlsx"

type FileService struct {
	repos *repository.Repository
}

func NewFileService(repos *repository.Repository) *FileService {
	return &FileService{repos: repos}
}

// сохранение файла на сервере
func (service *FileService) SaveFileToServe(file multipart.File) error {

	// создание пустого файла в файловой системе
	// os.Mkdir("./web/files/", 0777)
	saveFile, err := os.OpenFile("./web/files/"+fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return errors.New("FileService/SaveFileToServe()/ ошибка при создании файла в системе: " + err.Error())
	}
	defer saveFile.Close()

	// копирование данных из запроса в созданный файл
	io.Copy(saveFile, file)
	return nil
}

// запись данных из файла в БД
func (service *FileService) ParseDataFileToDB() error {
	// открытие файла
	f, err := excelize.OpenFile("./web/files/" + fileName)
	if err != nil {
		return errors.New("FileService/ParseDataFileToDB()/ ошибка при открытии файла: " + err.Error())
	}
	defer func() {
		// Close the spreadsheet.
		if err := f.Close(); err != nil {
			logrus.Println("ошибка при закрытии таблицы: ", err)
		}
	}()

	// массив, для хранения данных о пользователях
	//clients := make([]*app.User, 0)

	// получение строк из таблицы на листе Sheet1
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return errors.New("FileService/ParseDataFileToDB()/ ошибка при получении строк из файла: " + err.Error())
	}

	// проверка каждой ячейки
	for i, row := range rows {
		// пропустим первую строчку (где содержится шапка таблицы)
		if i == 0 {
			continue
		}

		client := &app.User{}

		for j, colCell := range row {

			switch j {
			case 0:
				client.LastName = strings.TrimSpace(colCell)
			case 1:
				client.FirstName = strings.TrimSpace(colCell)
			case 2:
				client.MiddleName = strings.TrimSpace(colCell)
			case 3:
				client.Debt = colCell
			}
		}
		client.Username = fmt.Sprintf("%s %s %s", client.LastName, client.FirstName, client.MiddleName)
		logrus.Println(*client)

		// запись данных пользователя в БД
		err := service.repos.AddDebtByClient(client)
		if err != nil {
			return errors.New("FileService/ParseDataFileToDB(): " + err.Error())
		}
		//clients = append(clients, client)
	}

	return nil
}

// открыть файл
/*
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
}*/
