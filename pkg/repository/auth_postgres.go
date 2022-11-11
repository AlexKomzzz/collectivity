package repository

import (
	"errors"
	"fmt"

	app "github.com/AlexKomzzz/collectivity"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

const (
	// названия таблиц в БД
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

// создание админа
func (r *AuthPostgres) CreateAdmin(admin app.User) error {

	query := fmt.Sprintf("INSERT INTO %s (first_name, last_name, password_hash, email, role_user) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (email) DO UPDATE SET first_name = EXCLUDED.first_name, last_name=EXCLUDED.last_name, password_hash=EXCLUDED.password_hash", DBusers)

	_, err := r.db.Exec(query, admin.FirstName, admin.LastName, admin.Password, admin.Email, "admin")

	return err
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

// проверка на отсутствие пользователя с таким email в БД
func (r *AuthPostgres) CheckUserByEmail(email string) (bool, error) {
	var yes bool
	query := fmt.Sprintf("SELECT EXISTS (SELECT id FROM %s WHERE email=$1)", DBusers)
	err := r.db.Get(&yes, query, email)

	return yes, err
}

// определение id пользователя по email и паролю
// id = -1 значит пользователя с данным email нет в БД
// id = -2 значит неправильно введен пароль
func (r *AuthPostgres) GetUser(email, password string) (int, error) {

	var yes bool
	var id int

	// открытие транзакции
	tx, err := r.db.Begin()
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	defer tx.Rollback()

	queryYes := fmt.Sprintf("SELECT EXISTS (SELECT id FROM %s WHERE email=$1 AND password_hash=$2)", DBusers)

	// проверка на наличие пользователя в БД
	row := tx.QueryRow(queryYes, email, password)
	err = row.Scan(&yes)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	// при отсутствии данного пользователя, проверяется отдельно наличие пользователя по email и по паролю
	if !yes {
		var yesEmail bool

		// проверка на наличие пользователя в БД c таким email
		queryEmail := fmt.Sprintf("SELECT EXISTS (SELECT id FROM %s WHERE email=$1)", DBusers)
		row := tx.QueryRow(queryEmail, email)
		err = row.Scan(&yesEmail)
		if err != nil {
			tx.Rollback()
			return 0, err
		}

		if !yesEmail {
			// значит пользователя с данным емейл нет в БД
			tx.Rollback()
			return -1, nil
		} else {
			//  проверка по паролю
			var yesPass bool

			queryPass := fmt.Sprintf("SELECT EXISTS (SELECT id FROM %s WHERE password_hash=$1)", DBusers)
			row := tx.QueryRow(queryPass, password)
			err = row.Scan(&yesPass)
			if err != nil {
				tx.Rollback()
				return 0, err
			}

			if !yesPass {
				// значит неправильно набран пароль
				tx.Rollback()
				return -2, nil
			} else {
				// такого случая быть не должно
				tx.Rollback()
				return 0, errors.New("ошибка БД при проверке пользователя по email и паролю")
			}
		}

	} else {
		query := fmt.Sprintf("SELECT id FROM %s WHERE email=$1 AND password_hash=$2", DBusers)
		row := tx.QueryRow(query, email, password)
		err := row.Scan(&id)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
	}

	// проверка
	if id == 0 {
		tx.Rollback()
		logrus.Println("ошибка: при определении id пользователя через email и pass")
		return 0, errors.New("непредвиденная ошибка при запросе к БД")
	}
	tx.Commit()

	return id, nil
}

// получение данных о пользователе (с неподтвержденным email) из БД authdata
func (r *AuthPostgres) GetUserFromAuth(idUserAuth int) (app.User, error) {
	var dataUser app.User
	// dataUser := app.User{
	// 	Id: idUserAuth,
	// }

	query := fmt.Sprintf("SELECT first_name, last_name, middle_name, email, password_hash FROM %s WHERE id=$1", DBauth)
	// logrus.Println("id = ", idUserAuth)
	err := r.db.Get(&dataUser, query, idUserAuth)
	// logrus.Println("dataUser = ", dataUser)
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
	var yes bool
	// проверка на наличие пользователя в БД
	query := fmt.Sprintf("SELECT EXISTS (SELECT id FROM %s WHERE email=$1)", DBusers)
	err := r.db.Get(&yes, query, email)
	if err != nil {
		return -1, err
	}

	// если пользователя нет
	if !yes {
		return -1, errors.New("sql: no rows in result set")
	}

	queryGet := "SELECT id FROM users WHERE email=$1"
	var id int
	err = r.db.Get(&id, queryGet, email)
	if err != nil {
		return -1, err
	}

	return id, nil
}

// обновление пароля у пользователя
func (r *AuthPostgres) UpdatePass(idUser int, newHashPsw string) error {

	query := "UPDATE users SET password_hash=$1 WHERE id=$2"
	_, err := r.db.Exec(query, newHashPsw, idUser)
	if err != nil {
		return err
	}

	return nil
}

// проверка роли пользователя по id
func (r *AuthPostgres) GetRole(idUser int) (string, error) {
	var roleUser string

	query := "SELECT role_user FROM users WHERE id=$1"
	err := r.db.Get(&roleUser, query, idUser)

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
