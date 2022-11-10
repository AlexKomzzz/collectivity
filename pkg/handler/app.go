package handler

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// стартовая страница
func (h *Handler) startList(c *gin.Context) {

	// идентифицируем пользователя
	idUser, err := h.userIdentity(c)
	if err != nil {
		logrus.Println("Вход без идентификации")
		c.Redirect(http.StatusTemporaryRedirect, "/auth/login")
		return
	}

	// проверка роли пользователя
	roleUser, err := h.service.GetRole(idUser)
	if err != nil {
		logrus.Println("ошибка при проверке роли пользователя в БД при входе на стартовую страницу: ", err)
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	if roleUser == "admin" {
		// значит это админ
		logrus.Println("вход админа")
		c.HTML(http.StatusOK, "start_list.html", gin.H{
			"role": true,
		})
		return
	} else {
		logrus.Println("обычный пользователь")
		c.HTML(http.StatusOK, "start_list.html", gin.H{})
		return
	}
}

// получение файла от клиента
func (h *Handler) parsFile(c *gin.Context) {

	// получение файла из запроса от клиента по ключевому слову "file"
	file, hdr, err := c.Request.FormFile("file")
	if err != nil {
		logrus.Println(err)
		errorServerResponse(c, err)
		return
	}
	defer file.Close()

	// создание пустого файла в файловой системе
	// os.Mkdir("./web/files/", 0777)
	saveFile, err := os.OpenFile("./web/files/"+hdr.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		logrus.Println(err)
		errorServerResponse(c, err)
		return
	}
	defer saveFile.Close()

	// копирование данных из запроса в созданный файл
	io.Copy(saveFile, file)

	// чтение файла
	readFile, err := os.Open(fmt.Sprintf("./web/files/%s", fileName))
	if err != nil {
		logrus.Println(err)
		errorServerResponse(c, err)
		return
	}
	defer readFile.Close()

	buff, err := ioutil.ReadAll(readFile)
	if err != nil {
		logrus.Println(err)
		errorServerResponse(c, err)
		return
	}

	logrus.Println("записанный файл: ", string(buff))

	c.JSON(http.StatusOK, gin.H{
		"file": string(buff),
	})
}
