package db

import (
	"github.com/go-redis/redis"
	"log"
)

var Redis *redis.Client

func ConnectRedis() {
	Redis = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	pong, err := Redis.Ping().Result()
	if err != nil {
		log.Println(err.Error())
	}
	log.Println("redis", pong)
}
