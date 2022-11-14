package repository

import (
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

type ConfigRedis struct {
	Addr     string
	Password string
	DB       int
}

func NewRedisCache(cfg ConfigRedis) (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	status := redisClient.Ping()
	err := status.Err()

	logrus.Print("Connect status server Redis: ", status)

	// redisClient.FlushAll(context) // Очистить Redis

	return redisClient, err
}
