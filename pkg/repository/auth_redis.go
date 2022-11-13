package repository

import (
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

	status := r.redisClient.Set(fmt.Sprintf("user:%d", userIdAPI), userData, duration)

	return status.Err()
}

// получение из кэша данных о пользователе по ключу idUser
func (r *AuthRedis) GetUserCash(userIdAPI int) ([]byte, error) {

	status := r.redisClient.Get(fmt.Sprintf("user:%d", userIdAPI))
	if status.Err() != nil {
		return nil, status.Err()
	}

	return []byte(status.Val()), nil
}
