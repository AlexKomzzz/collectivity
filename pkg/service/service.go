package service

import (
	app "github.com/AlexKomzzz/collectivity"
	"github.com/AlexKomzzz/collectivity/pkg/repository"
)

type Authorization interface {
	// функция создания Пользователя в БД
	// возвращяет id
	CreateUser(user app.User) (int, error)
	// функция создания Пользователя при авторизации через Google или Яндекс
	CreateUserAPI(typeAPI, idAPI, firstName, lastName, email string) (int, error)
	// генерация JWT по email и паролю
	GenerateJWT(email, password string) (string, error)
	// генерация JWT при Google или Яндекс авторизации
	// в переменную typeAPI необходимо передать 'google' либо 'yandex'
	GenerateJWT_API(idUser int) (string, error)
}

type Service struct {
	Authorization
}

func NewService(repos *repository.Repository) *Service {
	return &Service{
		Authorization: NewAuthService(repos),
	}
}
