package config

import (
	"github.com/gofiber/session/v2"
)

var (
	Store       = session.New()
	EmailTokens = make(map[string]string)
)
