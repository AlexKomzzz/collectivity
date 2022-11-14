package service

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"

	app "github.com/AlexKomzzz/collectivity"
	"github.com/AlexKomzzz/collectivity/pkg/repository"
	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (

	// данные админа
	emailAdmin = "admin@admin.le"
	fNameAdmin = "collect"
	lNameAdmin = "house"

	SOLT            = "bt,&13#Rkm*54FS#$WR2@#nasf!ds5fre%"
	tokenTTL        = 300 * time.Hour               // время жизни токена аутентификации
	tokenTTLtoEmail = 15 * time.Minute              // время жизни токена при восстановлении пароля или подтверждении почты
	JWT_SECRET      = "rkjk#4#%35FSFJlja#4353KSFjH" // секрет для JWT
	JWTemail_SECRET = "r2sk#4#gdfoij743*#weg(FjH"   // секрет для токена при восстановлении пароля и подтверждения почты
)

type AuthService struct {
	repos *repository.Repository
}

func NewAuthService(repos *repository.Repository) *AuthService {
	return &AuthService{repos: repos}
}

type tokenClaims struct {
	jwt.StandardClaims
	UserId int `json:"user_id"`
}

// хэширование пароля
func generatePasswordHash(password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))
	return fmt.Sprintf("%x", hash.Sum([]byte(SOLT)))
}

// генерация JWT с id пользователя
func generateJWT(idUser int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{ // генерация токена
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(tokenTTL).Unix(), // время действия токена
			IssuedAt:  time.Now().Unix(),               //время создания
		},
		idUser,
	})

	return token.SignedString([]byte(JWT_SECRET))
}

// создание админа
func (service *AuthService) CreateAdmin() error {
	psw := os.Getenv("pswad")
	admin := app.User{
		Username:  fmt.Sprintf("%s %s", lNameAdmin, fNameAdmin),
		FirstName: fNameAdmin,
		LastName:  lNameAdmin,
		Email:     emailAdmin,
		Password:  generatePasswordHash(psw),
	}

	return service.repos.CreateAdmin(admin)
}

// создание пользователя в БД регистрации (при создании нового пользователя, когда эл. почта не потверждена)
// возвращяет id созданного пользователя из таблицы authdata
/*func (service *AuthService) CreateUserByAuth(user *app.User, passRepeat string) (int, error) {

	// захешим и сравним переданные пароли
	err := service.CheckPass(&user.Password, &passRepeat)
	if err != nil {
		return 0, err
	}

	return service.repos.CreateUserByAuth(user)
}*/

// создание пользователя в БД (при создании нового пользователя и потверждении эл.почты)
// возвращяет id созданного пользователя из таблицы users
func (service *AuthService) CreateUser(user *app.User) (int, error) {
	ok, err := service.repos.CheckUser(user.Username)
	if err != nil {
		return 0, errors.New("AuthService/CreateUser(): " + err.Error())
	}
	if ok {
		return service.repos.CreateUser(user)
	} else {
		return -1, errors.New("данный пользователь не может быть зарегистрирован")
	}
}

// создание пользователя в БД auth (idUser уже известен)
func (service *AuthService) CreateUserById(user *app.User) error {
	return service.repos.CreateUserById(user)
}

// проверка на отсутствие пользователя с таким email в БД
func (service *AuthService) CheckUserByEmail(email string) (bool, error) {
	return service.repos.CheckUserByEmail(email)
}

// проверка на существование пользователя с таким ФИО в БД users и получение idUser
func (service *AuthService) GetUserByUsername(username string) (int, error) {
	ok, err := service.repos.CheckUser(username)
	if err != nil {
		return 0, err
	}
	if !ok {
		return -1, nil
	} else {
		return service.repos.GetUserByUsername(username)

	}
}

// хэширование и проверка паролей на соответсвие
func (service *AuthService) CheckPass(psw, refreshPsw *string) error {
	// захешим пароли
	// logrus.Printf("psw: %s, psw_double: %s", *psw, *psw)
	*psw = generatePasswordHash(*psw)
	*refreshPsw = generatePasswordHash(*refreshPsw)

	// Сравним переданные пароли
	if *psw != *refreshPsw {
		return errors.New("AuthService/CheckPass(): пароли не совпадают")
	}

	return nil
}

/*
// функция создания Пользователя при авторизации через Google или Яндекс
func (service *AuthService) CreateUserAPI(typeAPI, idAPI, firstName, lastName, email string) (int, error) {

	//

	return service.repos.CreateUserAPI(typeAPI, idAPI, firstName, lastName, email)
}*/

// конвертация idUser из строки в число
func (service *AuthService) ConvertID(idUser string) (int, error) {
	if len(idUser) > 18 {
		// обрезать до 17 символов
		idUser = string([]byte(idUser)[:19])
	}
	return strconv.Atoi(idUser)
}

// определение idUser по email
func (service *AuthService) GetUserByEmail(email string) (int, error) {
	return service.repos.GetUserByEmail(email)
}

// получение данных о пользователе (с неподтвержденным email) из БД authdata
/*func (service *AuthService) GetUserFromAuth(idUserAuth int) (app.User, error) {
	return service.repos.GetUserFromAuth(idUserAuth)
}*/

// получение данных о пользователе из кэша
func (service *AuthService) GetUserCash(idUserAPI int) ([]byte, error) {
	return service.repos.GetUserCash(idUserAPI)
}

// проверка роли пользователя по id
func (service *AuthService) GetRole(idUser int) (string, error) {

	return service.repos.GetRole(idUser)
}

// генерация JWT по email и паролю
func (service *AuthService) GenerateJWT(email, password string) (string, error) {
	// определим id пользователя
	idUser, err := service.repos.GetUser(email, generatePasswordHash(password))
	if err != nil {
		return "", err
	}
	if idUser == -1 {
		return "", errors.New("нет пользователя")
	} else if idUser == -2 {
		return "", errors.New("пароль")
	}

	return generateJWT(idUser)
}

// генерация JWT с указанием idUser
func (service *AuthService) GenerateJWTbyID(idUser int) (string, error) {

	return generateJWT(idUser)
}

// генерация токена для восстановления пароля или подтверждения почты
func (service *AuthService) GenerateJWTtoEmail(idUser int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{ // генерация токена
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(tokenTTLtoEmail).Unix(), // время действия токена
			IssuedAt:  time.Now().Unix(),                      //время создания
		},
		idUser,
	})

	return token.SignedString([]byte(JWTemail_SECRET))
}

// проверка токена на валидность
func (service *AuthService) ValidToken(headerAuth string) (int, error) {

	headerParts := strings.Split(headerAuth, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" || headerParts[1] == "" {
		return -1, errors.New("invalid auth header")
	}

	return service.ParseToken(headerParts[1])
}

// Парс токена (получаем из токена id)
func (service *AuthService) ParseToken(accesstoken string) (int, error) {
	token, err := jwt.ParseWithClaims(accesstoken, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method JWT")
		}
		return []byte(JWT_SECRET), nil
	})
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok {
		return 0, errors.New("token claims are not of type *tokenClaims")
	}

	return claims.UserId, nil
}

// Парс токена при восстановлении пароля или подтверждении почты
func (service *AuthService) ParseTokenEmail(accesstoken string) (int, error) {
	token, err := jwt.ParseWithClaims(accesstoken, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("AuthService/ParseTokenEmail()/ParseWithClaims()/token.Method(): invalid signing method")
		}
		return []byte(JWTemail_SECRET), nil
	})
	if err != nil {
		return 0, errors.New("AuthService/ParseTokenEmail()/ParseWithClaims(): " + err.Error())
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok {
		return 0, errors.New("AuthService/ParseTokenEmail()/token.Claims(): token claims are not of type *tokenClaims")
	}

	return claims.UserId, nil
}

// запись данных о пользователе в кэш
func (service *AuthService) SetUserCash(idUserAPI int, userData []byte) error {
	return service.repos.SetUserCash(idUserAPI, userData)
}

// отправка сообщения пользователю на почту для передачи ссылки
func (service *AuthService) SendMessageByMail(emailUser, url, msg string) error {

	// почта от куда отправляется ссылка
	emailAPI := os.Getenv("emailAddr")
	passwordAPI := os.Getenv("SMTPpwd")
	host := viper.GetString("email.host")
	port := viper.GetString("email.port")
	address := host + ":" + port

	// Настройка аутентификации отправителя
	// auth := sasl.NewPlainClient("", emailAPI, passwordAPI)
	auth := smtp.PlainAuth("", emailAPI, passwordAPI, host)
	// logrus.Println("auth: ", auth)
	// logrus.Println("address: ", address)
	// logrus.Println("emailAPI: ", emailAPI)

	// список рассылки
	to := []string{emailUser}
	logrus.Println("send mes START")
	err := smtp.SendMail(address, auth, emailAPI, to, []byte(msg))
	logrus.Println("send mes OK")

	if err != nil {
		return errors.New("AuthService/SendMessageByMail()/ ошибка при отправке письма на почту пользователя: " + err.Error())
	}

	return nil
}

// обновление пароля у пользователя
func (service *AuthService) UpdatePass(idUser int, newHashPsw string) error {

	return service.repos.UpdatePass(idUser, newHashPsw)
}

// сравнение email полученного и сохраненного в БД
func (service *AuthService) ComparisonEmail(emailUser, emailURL string) error {

	// Сравним emails
	if emailUser != emailURL {
		return errors.New("AuthService/ComparisonEmail(): emails не совпадают")
	}

	return nil
}

// сравнение idUser из JWT и из кэша
func (service *AuthService) ComparisonId(idJWT, idCash int) error {

	if idJWT != idCash {
		return errors.New("AuthService/ComparisonId(): id не совпадают")
	}

	return nil
}
