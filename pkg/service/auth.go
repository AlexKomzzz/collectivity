package service

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"net/smtp"
	"os"
	"strings"
	"time"

	app "github.com/AlexKomzzz/collectivity"
	"github.com/AlexKomzzz/collectivity/pkg/repository"
	"github.com/dgrijalva/jwt-go"
)

const SOLT = "bt,&13#Rkm*54FS#$WR2@#nasf!ds5fre%"

type AuthService struct {
	repos *repository.Repository
}

func NewAuthService(repos *repository.Repository) *AuthService {
	return &AuthService{repos: repos}
}

const tokenTTL = 300 * time.Hour

const JWT_SECRET = "rkjk#4#%35FSFJlja#4353KSFjH"

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

// функция создания Пользователя в БД
// возвращяет id
func (service *AuthService) CreateUser(user *app.User, passRepeat string) (int, error) {
	// захешим пароли
	user.Password = generatePasswordHash(user.Password)
	passRepeat = generatePasswordHash(passRepeat)

	// Сравним переданные пароли
	if user.Password != passRepeat {
		return 0, errors.New("пароли не совпадают")
	}

	return service.repos.CreateUser(user)
}

// функция создания Пользователя при авторизации через Google или Яндекс
func (service *AuthService) CreateUserAPI(typeAPI, idAPI, firstName, lastName, email string) (int, error) {

	return service.repos.CreateUserAPI(typeAPI, idAPI, firstName, lastName, email)
}

// генерация JWT по email и паролю
func (service *AuthService) GenerateJWT(email, password string) (string, error) {
	// определим id пользователя
	idUser, err := service.repos.GetUser(email, generatePasswordHash(password))
	if err != nil {
		return "", err
	}

	return generateJWT(idUser)
}

// генерация JWT при Google или Яндекс авторизации
// в переменную typeAPI необходимо передать 'google' либо 'yandex'
func (service *AuthService) GenerateJWT_API(idUser int) (string, error) {
	// определим id пользователя
	// idUser, err := service.repos.GetUserAPI(typeAPI, idAPI, email)
	// if err != nil {
	// 	return "", err
	// }

	return generateJWT(idUser)
}

// проверка токена на валидность
func (service *AuthService) ValidToken(headerAuth string) (int, error) {

	headerParts := strings.Split(headerAuth, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" || headerParts[1] == "" {
		return -1, errors.New("invalid auth header")
	}

	userId, err := service.ParseToken(headerParts[1])
	if err != nil {
		return -1, err
	}

	return userId, nil
}

// Парс токена (получаем из токена id)
func (service *AuthService) ParseToken(accesstoken string) (int, error) {
	token, err := jwt.ParseWithClaims(accesstoken, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
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

// функция проверки пользователя по email
func (service *AuthService) DefinitionUserByEmail(email string) (string, error) {
	idUser, err := service.repos.GetUserByEmail(email)
	if err != nil {
		return "", nil
	}
	return generateJWT(idUser)
}

// отправка сообщения пользователю на почту для передачи ссылки на восстановление пароля
func (service *AuthService) SendMessage(emailUser, url string) error {

	// почта от куда отправляется ссылка
	emailAPI := os.Getenv("emailAddr")
	passwordAPI := os.Getenv("SMTPpwd")
	// emailAPI := "komalex203@gmail.com"
	// passwordAPI := "3411019873svetA"
	// host := "smtp.gmail.com"
	host := "smtp.yandex.ru"
	// port := "587"
	port := "465"
	address := host + ":" + port

	// Настройка аутентификации отправителя
	// auth := sasl.NewPlainClient("", emailAPI, passwordAPI)
	auth := smtp.PlainAuth("", emailAPI, passwordAPI, host)

	// список рассылки
	to := []string{emailUser}

	msg := fmt.Sprintf("To: %s\r\n"+
		"Subject: Восстановление пароля\r\n"+
		"\r\n"+
		"Для восстановления пароля перейдите по ссылке: %s.\r\n", to[0], url)
	err := smtp.SendMail(address, auth, emailAPI, to, []byte(msg))
	if err != nil {
		return err
	}

	return nil
}
