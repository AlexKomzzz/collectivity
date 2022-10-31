package repository

import (
	app "github.com/AlexKomzzz/collectivity"
	"github.com/jmoiron/sqlx"
)

type Authorization interface {
	// создание пользователя в БД (при создании нового пользователя, при потверждении эл.почты)
	// необходимо передать структуру User с хэшифрованным паролем
	CreateUser(user *app.User) (int, error)
	// создание пользователя в БД регистрации (при создании нового пользователя, когда эл. почта не потверждена)
	CreateUserByAuth(user *app.User) (int, error)
	// создание пользователя в БД при авторизации через Google  или Яндекс
	CreateUserAPI(typeAPI, idAPI, firstName, lastName, email string) (int, error)
	// определение id пользователя по email и паролю
	GetUser(email, password string) (int, error)
	// получение данных о пользователе (с неподтвержденным email) из БД authdata
	GetUserFromAuth(idUserAuth int) (app.User, error)
	// определение id пользователя по email
	GetUserByEmail(email string) (int, error)
	// обновление пароля у пользователя
	UpdatePass(idUser int, newHashPsw string) error
	// проверка роли пользователя по id
	GetRole(idUser int) (string, error)
	// определение id пользователя по email и id для Google и Яндекс API
	// в переменную typeAPI необходимо передать 'google' либо 'yandex'
	// GetUserAPI(typeAPI, idAPI, email string) (int, error)
}

type Repository struct {
	Authorization
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Authorization: NewAuthPostgres(db),
	}
}
