package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

const (
	duration time.Duration = 11 * time.Minute // тайм-айт ключа в кэше Redis
)

type AuthRedis struct {
	redisClient *redis.Client
}

func NewAuthRedis(redisClient *redis.Client) *AuthRedis {
	return &AuthRedis{
		redisClient: redisClient,
	}
}

// добавление в кэш данных о пользователе с ключом idUser
func (r *AuthRedis) SetUserCash(userIdAPI int, userData []byte) error {

	return r.redisClient.Set(fmt.Sprintf("user:%d", userIdAPI), userData, duration).Err()
}

// получение из кэша данных о пользователе по ключу idUser
func (r *AuthRedis) GetUserCash(userIdAPI int) ([]byte, error) {

	val, err := r.redisClient.Get(fmt.Sprintf("user:%d", userIdAPI)).Result()
	if err != nil {
		return nil, errors.New("AuthRedis/GetUserCash()/ ошибка при получении данных о пользователе из кэша: " + err.Error())
	}

	return []byte(val), nil
}
