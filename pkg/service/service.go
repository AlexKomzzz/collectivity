package service

import (
	"mime/multipart"

	app "github.com/AlexKomzzz/collectivity"
	"github.com/AlexKomzzz/collectivity/pkg/repository"
)

type Authorization interface {
	// создание админа
	CreateAdmin() error
	// создание пользователя в БД регистрации (при создании нового пользователя, когда эл. почта не потверждена)
	// возвращяет id созданного пользователя из таблицы authdata
	CreateUserByAuth(user *app.User, passRepeat string) (int, error)
	// создание пользователя в БД (при создании нового пользователя и потверждении эл.почты)
	// возвращяет id созданного пользователя из таблицы users
	CreateUser(user *app.User) (int, error)
	// функция создания Пользователя при авторизации через Google или Яндекс
	// в переменную typeAPI необходимо передать 'google' либо 'yandex'
	//CreateUserAPI(typeAPI, idAPI, firstName, lastName, email string) (int, error)
	// проверка на отсутствие пользователя с таким email в БД
	CheckUserByEmail(email string) (bool, error)
	// хэширование и проверка паролей на соответсвие
	CheckPass(psw, refreshPsw *string) error
	// конвертация idUser из строки в число
	ConvertID(idUser string) (int, error)
	// определение idUser по email
	GetUserByEmail(email string) (int, error)
	// получение данных о пользователе (с неподтвержденным email) из БД authdata
	GetUserFromAuth(idUserAuth int) (app.User, error)
	// получение данных о пользователе из кэша
	GetUserCash(idUserAPI int) ([]byte, error)
	// проверка роли пользователя по id
	GetRole(idUser int) (string, error)
	// генерация JWT по email и паролю
	GenerateJWT(email, password string) (string, error)
	// генерация JWT с указанием idUser
	GenerateJWTbyID(idUser int) (string, error)
	// генерация токена для восстановления пароля или подтверждения почты
	GenerateJWTtoEmail(idUser int) (string, error)
	// проверка заголовка авторизации на валидность
	ValidToken(headerAuth string) (int, error)
	// парс токена (получаем из токена id)
	ParseToken(accesstoken string) (int, error)
	// парс токена при восстановлении пароля или подтверждении почты
	ParseTokenEmail(accesstoken string) (int, error)
	// запись данных о пользователе в кэш
	SetUserCash(idUserAPI int, userData []byte) error
	// отправка сообщения пользователю на почту для передачи ссылки
	SendMessageByMail(emailUser, url, msg string) error
	// обновление пароля у пользователя
	UpdatePass(idUser int, newHashPsw string) error
	// сравнение email полученного и сохраненного в БД
	ComparisonEmail(emailUser, emailURL string) error
}

type FileHanding interface {
	// сохранение файла на сервере
	SaveFileToServe(file multipart.File) error
	// запись данных из файла в БД
	ParseDataFileToDB() error
}

type Service struct {
	Authorization
	FileHanding
}

func NewService(repos *repository.Repository) *Service {
	return &Service{
		Authorization: NewAuthService(repos),
		FileHanding:   NewFileService(repos),
	}
}
