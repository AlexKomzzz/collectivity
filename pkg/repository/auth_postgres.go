package repository

import (
	"fmt"

	app "github.com/AlexKomzzz/collectivity"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

const (
	DBusers = "users"
	DBauth  = "authdata"
)

type AuthPostgres struct {
	db *sqlx.DB
}

func NewAuthPostgres(db *sqlx.DB) *AuthPostgres {
	return &AuthPostgres{
		db: db,
	}
}

// создание пользователя в БД (при создании нового пользователя и потверждении эл.почты)
// необходимо передать структуру User с зашифрованным паролем
func (r *AuthPostgres) CreateUser(user *app.User) (int, error) {

	query := fmt.Sprintf("INSERT INTO %s (first_name, last_name, middle_name, password_hash, email) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (email) DO UPDATE SET first_name = EXCLUDED.first_name, last_name=EXCLUDED.last_name, middle_name=EXCLUDED.middle_name, password_hash=EXCLUDED.password_hash RETURNING id", DBusers)

	row := r.db.QueryRow(query, user.FirstName, user.LastName, user.MiddleName, user.Password, user.Email)
	var id int
	err := row.Scan(&id)
	if err != nil {
		logrus.Printf("error Scan by CreateUser: %s\n", err.Error())
		return 0, err
	}

	return id, nil
}

// создание пользователя в БД регистрации (при создании нового пользователя, когда эл. почта не потверждена)
func (r *AuthPostgres) CreateUserByAuth(user *app.User) (int, error) {

	query := fmt.Sprintf("INSERT INTO %s (first_name, last_name, middle_name, password_hash, email) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (email) DO UPDATE SET first_name = EXCLUDED.first_name, last_name=EXCLUDED.last_name, middle_name=EXCLUDED.middle_name, password_hash=EXCLUDED.password_hash RETURNING id", DBauth)

	row := r.db.QueryRow(query, user.FirstName, user.LastName, user.MiddleName, user.Password, user.Email)
	var id int
	err := row.Scan(&id)
	if err != nil {
		logrus.Printf("error Scan by CreateUser: %s\n", err.Error())
		return 0, err
	}

	return id, nil
}

// создание пользователя в БД при авторизации через Google или Яндекс
func (r *AuthPostgres) CreateUserAPI(typeAPI, idAPI, firstName, lastName, email string) (int, error) {

	query := fmt.Sprintf("INSERT INTO users (id_%s, first_name, last_name, email) VALUES ($1, $2, $3, $4) ON CONFLICT (id_%s) DO UPDATE SET first_name = EXCLUDED.first_name, last_name=EXCLUDED.last_name RETURNING id", typeAPI, typeAPI)

	row := r.db.QueryRow(query, idAPI, firstName, lastName, email)
	var idUser int
	err := row.Scan(&idUser)
	if err != nil {
		logrus.Printf("error Scan by CreateUserAPIGoogle: %s\n", err.Error())
		return 0, err
	}

	return idUser, nil
}

// определение id пользователя по email и паролю
func (r *AuthPostgres) GetUser(email, password string) (int, error) {
	query := "SELECT id FROM users WHERE email=$1 AND password_hash=$2"
	var id int
	err := r.db.Get(&id, query, email, password)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// получение данных о пользователе (с неподтвержденным email) из БД authdata
func (r *AuthPostgres) GetUserFromAuth(idUserAuth int) (app.User, error) {
	var dataUser app.User

	query := fmt.Sprintf("SELECT (first_name, last_name, middle_name, email, password_hash) FROM %s WHERE id=$1", DBauth)
	logrus.Println("id = ", idUserAuth)
	err := r.db.Get(&dataUser, query, idUserAuth)
	if err != nil {
		return dataUser, err
	}

	return dataUser, nil
}

// определение id пользователя по email и id для Google и Яндекс API
// в переменную typeAPI необходимо передать 'google' либо 'yandex'
/*func (r *AuthPostgres) GetUserAPI(typeAPI, idAPI, email string) (int, error) {
	query := fmt.Sprintf("SELECT id FROM users WHERE id_%s=$1 AND email=$2", typeAPI)
	var id int
	err := r.db.Get(&id, query, idAPI, email)
	if err != nil {
		return 0, err
	}

	return id, nil
}*/

// определение id пользователя по email
func (r *AuthPostgres) GetUserByEmail(email string) (int, error) {
	query := "SELECT id FROM users WHERE email=$1"
	var id int
	err := r.db.Get(&id, query, email)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// обновление пароля у пользователя
func (service *AuthPostgres) UpdatePass(idUser int, newHashPsw string) error {

	query := "UPDATE password_hash=$1 FROM users WHERE id=$2" // ПРОВЕРИТЬ!!!!!
	_, err := service.db.Exec(query, newHashPsw, idUser)
	if err != nil {
		return err
	}

	return nil
}

// проверка роли пользователя по id
func (service *AuthPostgres) GetRole(idUser int) (string, error) {
	var roleUser string

	query := "SELECT role FROM users WHERE id=$1"
	err := service.db.Get(&roleUser, query, idUser)

	return roleUser, err
}

/*
// определение username пользователя по id
func (r *AuthPostgres) GetUsername(userId int) (string, error) {
	query := "SELECT username FROM users WHERE id=$1"
	var username string
	err := r.db.Get(&username, query, userId)
	if err != nil {
		return "", err
	}

	return username, nil
}
*/
