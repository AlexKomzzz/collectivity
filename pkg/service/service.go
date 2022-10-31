package service

import (
	app "github.com/AlexKomzzz/collectivity"
	"github.com/AlexKomzzz/collectivity/pkg/repository"
)

type Authorization interface {
	// функция создания Пользователя в БД
	// возвращяет id
	CreateUser(user *app.User, passRepeat string) (int, error)
	// функция создания Пользователя при авторизации через Google или Яндекс
	CreateUserAPI(typeAPI, idAPI, firstName, lastName, email string) (int, error)
	// генерация JWT по email и паролю
	GenerateJWT(email, password string) (string, error)
	// генерация JWT при Google или Яндекс авторизации
	// в переменную typeAPI необходимо передать 'google' либо 'yandex'
	GenerateJWT_API(idUser int) (string, error)
	// проверка заголовка авторизации на валидность
	ValidToken(headerAuth string) (int, error)
	// Парс токена (получаем из токена id)
	ParseToken(accesstoken string) (int, error)
	// Парс токена при восстановлении пароля или подтверждении почты
	ParseTokenEmail(accesstoken string) (int, error)
	// идентификация пользователя по email
	DefinitionUserByEmail(email string) (string, error)
	// отправка сообщения пользователю на почту для передачи ссылки на восстановление пароля
	SendMessage(emailUser, url string) error
	// хэширование и проверка паролей на соответсвие
	CheckPass(psw, refreshPsw *string) error
	// обновление пароля у пользователя
	UpdatePass(idUser, newHashPsw string) error
	// проверка роли пользователя по id
	GetRole(idUser int) (string, error)
}

type Service struct {
	Authorization
}

func NewService(repos *repository.Repository) *Service {
	return &Service{
		Authorization: NewAuthService(repos),
	}
}
