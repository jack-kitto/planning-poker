package session

import (
	"time"

	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/fiber/v2/utils"
)

var (
	Store = session.New(
		session.Config{
			Expiration:   24 * time.Hour,
			KeyLookup:    "cookie:session_id",
			KeyGenerator: utils.UUIDv4,
		},
	)
	EmailTokens = make(map[string]string)
)
