package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"auth/config"
	e "auth/email"
)

type InMemoryStorageI interface {
	Set(key, value string, exp time.Duration) error
	Get(key string) (string, error)
	Del(key string) error
	SaveToken(email string, token string, exp time.Duration) error
}

type storageRedis struct {
	client *redis.Client
}

func NewInMemoryStorage(rdb *redis.Client) InMemoryStorageI {
	return &storageRedis{
		client: rdb,
	}
}

func (r *storageRedis) Set(key, value string, exp time.Duration) error {
	err := r.client.Set(context.Background(), key, value, exp).Err()
	if err != nil {
		return err
	}
	return nil
}


func (r *storageRedis) Del(key string) error {
	err := r.client.Del(context.Background(), key).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *storageRedis) Get(key string) (string, error) {
	val, err := r.client.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key '%s' does not exist", key)
	} else if err != nil {
		return "", fmt.Errorf("error retrieving key '%s' from redis: %v", key, err)
	}
	return val, nil
}

func (r *storageRedis) SaveToken(email string, token string, exp time.Duration) error {
	cnf := config.Load()
	fmt.Println(token)

	st := e.SendEmailRequest{
		To:      []string{email},
		Type:    "Verification email",
		Subject: "Verification",
		Code:    token,
	}

	err := e.SendEmail(&cnf, &st)
	if err != nil {
		return fmt.Errorf("failed to send verification email to %s: %v", email, err)
	}

	err = r.Set(email, token, exp)
	if err != nil {
		return fmt.Errorf("failed to save token for %s in redis: %v", email, err)
	}

	return nil
}
