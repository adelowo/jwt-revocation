package main

import (
	"errors"

	"github.com/go-redis/redis/v7"
)

type Client struct {
	redis *redis.Client
}

const (
	blackListKey string = "jwt_blacklist"
)

var (
	errJTIBlacklisted = errors.New("jwt token has been blacklisted")
)

func NewRediClient(dsn string) (*Client, error) {

	client := redis.NewClient(&redis.Options{
		Addr:     dsn,
		Password: "",
		DB:       0,
	})

	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}

	return &Client{redis: client}, nil
}

func (c *Client) IsBlacklisted(jti string) error {
	m, err := c.redis.SMembersMap(blackListKey).Result()
	if err != nil {
		return err
	}

	if _, ok := m[jti]; ok {
		return errJTIBlacklisted
	}

	return nil
}

func (c *Client) AddToBlacklist(jti string) error {
	_, err := c.redis.SAdd(blackListKey, jti).Result()
	return err
}
