package service

import (
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
	CreateUserAPI(typeAPI, idAPI, firstName, lastName, email string) (int, error)
	// Проверка на отсутствие пользователя с таким email в БД
	CheckUserByEmail(email string) (bool, error)
	// хэширование и проверка паролей на соответсвие
	CheckPass(psw, refreshPsw *string) error
	// определение idUser по email
	GetUserByEmail(email string) (int, error)
	// получение данных о пользователе (с неподтвержденным email) из БД authdata
	GetUserFromAuth(idUserAuth int) (app.User, error)
	// проверка роли пользователя по id
	GetRole(idUser int) (string, error)
	// генерация JWT по email и паролю
	GenerateJWT(email, password string) (string, error)
	// генерация JWT при Google или Яндекс авторизации
	// в переменную typeAPI необходимо передать 'google' либо 'yandex'
	GenerateJWT_API(idUser int) (string, error)
	// генерация токена для восстановления пароля или подтверждения почты
	GenerateJWTtoEmail(idUser int) (string, error)
	// проверка заголовка авторизации на валидность
	ValidToken(headerAuth string) (int, error)
	// Парс токена (получаем из токена id)
	ParseToken(accesstoken string) (int, error)
	// Парс токена при восстановлении пароля или подтверждении почты
	ParseTokenEmail(accesstoken string) (int, error)
	// отправка сообщения пользователю на почту для передачи ссылки
	SendMessageByMail(emailUser, url, msg string) error
	// обновление пароля у пользователя
	UpdatePass(idUser int, newHashPsw string) error
	// сравнение email полученного и сохраненного в БД
	ComparisonEmail(emailUser, emailURL string) error
}

type Service struct {
	Authorization
}

func NewService(repos *repository.Repository) *Service {
	return &Service{
		Authorization: NewAuthService(repos),
	}
}
