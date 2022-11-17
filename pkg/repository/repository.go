package repository

import (
	app "github.com/AlexKomzzz/collectivity"
	"github.com/go-redis/redis"
	"github.com/jmoiron/sqlx"
)

type Authorization interface {
	// создание админа
	CreateAdmin(admin app.User) error
	// создание пользователя в БД (при создании нового пользователя, при потверждении эл.почты)
	// необходимо передать структуру User с хэшифрованным паролем
	CreateUser(user *app.User) (int, error)
	// создание пользователя в БД auth (idUser уже известен)
	CreateUserById(user *app.User) error
	// создание пользователя в БД регистрации (при создании нового пользователя, когда эл. почта не потверждена)
	// CreateUserByAuth(user *app.User) (int, error)

	// создание пользователя в БД при авторизации через Google  или Яндекс
	//CreateUserAPI(typeAPI, idAPI, firstName, lastName, email string) (int, error)

	// проверка на существование пользователя в таблице users
	CheckUser(username string) (bool, error)
	// проверка на отсутствие пользователя с таким email в БД
	CheckUserByEmail(email string) (bool, error)
	// определение id пользователя по email и паролю
	GetUser(email, password string) (int, error)
	// получение idUser по ФИО
	GetUserByUsername(username string) (int, error)
	// получение данных о пользователе (с неподтвержденным email) из БД authdata
	// GetUserFromAuth(idUserAuth int) (app.User, error)
	// определение id пользователя по email
	GetUserByEmail(email string) (int, error)
	// обновление пароля у пользователя
	UpdatePass(idUser int, emailUser, newHashPsw string) error
	// определение роли пользователя по id
	GetRole(idUser int) (string, error)
	// определение долго пользователя
	GetDebtUser(idUser int) (string, error)
	// определение id пользователя по email и id для Google и Яндекс API
	// в переменную typeAPI необходимо передать 'google' либо 'yandex'
	// GetUserAPI(typeAPI, idAPI, email string) (int, error)
}

type AuthCash interface {
	// добавление в кэш данных о пользователе с ключом idUser
	SetUserCash(userIdAPI int, userData []byte) error
	// получение из кэша данных о пользователе по ключу idUser
	GetUserCash(userIdAPI int) ([]byte, error)
}

type DataClients interface {
	// добавление данных по долгу у клиентв
	AddDebtByClient(client *app.User) error
}

type Repository struct {
	Authorization
	AuthCash
	DataClients
}

func NewRepository(db *sqlx.DB, redisClient *redis.Client) *Repository {
	return &Repository{
		Authorization: NewAuthPostgres(db),
		AuthCash:      NewAuthRedis(redisClient),
		DataClients:   NewDataClientsPostgres(db),
	}
}
