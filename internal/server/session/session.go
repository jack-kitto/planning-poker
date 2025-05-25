package session

import (
	"log"
	"planning-poker/internal/server/config"
	"time"

	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/gofiber/storage/redis"
)

var (
	Store       *session.Store
	EmailTokens = make(map[string]string)
)

func InitSessionStore() {
	config, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	redisStore := redis.New(redis.Config{
		Host:     config.RedisBaseURL,
		Port:     config.RedisPort,
		Password: "",
		Database: 0,
		Reset:    false,
	})

	Store = session.New(session.Config{
		Storage:      redisStore,
		Expiration:   24 * time.Hour,
		KeyLookup:    "cookie:session_id",
		KeyGenerator: utils.UUIDv4,
	})
}
