package service

import (
	app "github.com/AlexKomzzz/collectivity"
	"github.com/AlexKomzzz/collectivity/pkg/repository"
)

type Authorization interface {
	CreateUser(user app.User) (int, error)
	GenerateJWT(email, password string) (string, error)
}

type Service struct {
	Authorization
}

func NewService(repos *repository.Repository) *Service {
	return &Service{
		Authorization: NewAuthService(repos),
	}
}
