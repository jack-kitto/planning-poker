package session

import (
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
	redisStore := redis.New(redis.Config{
		Host:     "localhost", // Change if your Redis is elsewhere
		Port:     6379,
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
