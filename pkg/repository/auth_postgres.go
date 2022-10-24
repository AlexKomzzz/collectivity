package repository

import (
	app "github.com/AlexKomzzz/collectivity"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type AuthPostgres struct {
	db *sqlx.DB
}

func NewAuthPostgres(db *sqlx.DB) *AuthPostgres {
	return &AuthPostgres{
		db: db,
	}
}

// созданиепользователя в БД
// необходимо передать структуру User с зашифрованным паролем
func (r *AuthPostgres) CreateUser(user app.User) (int, error) {

	query := "INSERT INTO users (first_name, last_name, password_hash, email) VALUES ($1, $2, $3) ON CONFLICT (email) DO UPDATE SET first_name = EXCLUDED.first_name, last_name=EXCLUDED.last_name, password_hash=EXCLUDED.password_hash RETURNING id"

	row := r.db.QueryRow(query, user.FirstName, user.LastName, user.Email)
	var id int
	err := row.Scan(&id)
	if err != nil {
		logrus.Printf("error Scan by CreateUser: %s\n", err.Error())
		return 0, err
	}

	return id, nil
}

// определение id пользователя по email и паролю
func (r *AuthPostgres) GetUser(email, password string) (int, error) {
	query := "SELECT id FROM users WHERE email=$1 AND password=$2"
	var id int
	err := r.db.Get(&id, query, email, password)
	if err != nil {
		return 0, err
	}

	return id, nil
}

/*
// определение idпользователя по email
func (r *Repository) GetUserByEmail(email string) (int, error) {
	query := "SELECT id FROM users WHERE email=$1"
	var id int
	err := r.db.Get(&id, query, email)
	if err != nil {
		return 0, fmt.Errorf("error: 'select' from GetUserByEmail (repos): нет пользователя с таким email: %v", err)
	}

	return id, nil
}

// определение username пользователя по id
func (r *Repository) GetUsername(userId int) (string, error) {
	query := "SELECT username FROM users WHERE id=$1"
	var username string
	err := r.db.Get(&username, query, userId)
	if err != nil {
		return "", err
	}

	return username, nil
}
*/
