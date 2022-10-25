package service

import (
	"crypto/sha1"
	"errors"
	"fmt"
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
func (service *AuthService) CreateUser(user *app.User) (int, error) {
	// захешим пароль
	user.Password = generatePasswordHash(user.Password)

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

/*
// функция получения username по id
func (service *AuthService) GetUsername(userId int) (string, error) {
	return service.repos.GetUsername(userId)
}
*/
